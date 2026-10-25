package mcping

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"time"
)

const (
	maxPacketSize   = 1 << 20
	defaultDeadline = 5 * time.Second
	statusState     = 1
	// -1 asks the server for its status without negotiating a protocol version.
	anyProtocol = -1
)

type Players struct {
	Online int
	Max    int
}

// Ping returns nil players when the server answers but hides its player counts.
func Ping(ctx context.Context, address string) (*Players, error) {
	host, portText, err := net.SplitHostPort(address)
	if err != nil {
		return nil, fmt.Errorf("parse address: %w", err)
	}
	port, err := strconv.ParseUint(portText, 10, 16)
	if err != nil {
		return nil, fmt.Errorf("parse port: %w", err)
	}

	var dialer net.Dialer
	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}
	defer conn.Close()

	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(defaultDeadline)
	}
	if err := conn.SetDeadline(deadline); err != nil {
		return nil, fmt.Errorf("set deadline: %w", err)
	}

	if err := writePacket(conn, handshake(host, uint16(port))); err != nil {
		return nil, fmt.Errorf("send handshake: %w", err)
	}
	if err := writePacket(conn, []byte{0x00}); err != nil {
		return nil, fmt.Errorf("send status request: %w", err)
	}

	return readStatus(bufio.NewReader(conn))
}

func handshake(host string, port uint16) []byte {
	var packet bytes.Buffer
	packet.WriteByte(0x00)
	packet.Write(appendVarInt(nil, anyProtocol))
	packet.Write(appendVarInt(nil, int32(len(host))))
	packet.WriteString(host)
	packet.Write(binary.BigEndian.AppendUint16(nil, port))
	packet.Write(appendVarInt(nil, statusState))

	return packet.Bytes()
}

func writePacket(w io.Writer, payload []byte) error {
	_, err := w.Write(append(appendVarInt(nil, int32(len(payload))), payload...))
	return err
}

func readStatus(r *bufio.Reader) (*Players, error) {
	length, err := readVarInt(r)
	if err != nil {
		return nil, fmt.Errorf("read packet length: %w", err)
	}
	if length <= 0 || length > maxPacketSize {
		return nil, fmt.Errorf("unexpected packet length %d", length)
	}

	payload := make([]byte, length)
	if _, err := io.ReadFull(r, payload); err != nil {
		return nil, fmt.Errorf("read packet: %w", err)
	}

	packet := bytes.NewReader(payload)
	id, err := readVarInt(packet)
	if err != nil || id != 0x00 {
		return nil, errors.New("response is not a status packet")
	}
	size, err := readVarInt(packet)
	if err != nil || int(size) != packet.Len() {
		return nil, errors.New("status string has an invalid length")
	}

	var status struct {
		Players *struct {
			Online int `json:"online"`
			Max    int `json:"max"`
		} `json:"players"`
	}
	if err := json.NewDecoder(packet).Decode(&status); err != nil {
		return nil, fmt.Errorf("decode status: %w", err)
	}
	if status.Players == nil {
		return nil, nil
	}

	return &Players{Online: status.Players.Online, Max: status.Players.Max}, nil
}

func appendVarInt(buffer []byte, value int32) []byte {
	unsigned := uint32(value)
	for unsigned >= 0x80 {
		buffer = append(buffer, byte(unsigned)|0x80)
		unsigned >>= 7
	}

	return append(buffer, byte(unsigned))
}

func readVarInt(r io.ByteReader) (int32, error) {
	var result uint32
	for shift := 0; shift < 35; shift += 7 {
		next, err := r.ReadByte()
		if err != nil {
			return 0, err
		}

		result |= uint32(next&0x7f) << shift
		if next&0x80 == 0 {
			return int32(result), nil
		}
	}

	return 0, errors.New("varint is too long")
}
