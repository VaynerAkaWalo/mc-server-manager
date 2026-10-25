package mcping

import (
	"bufio"
	"bytes"
	"context"
	"net"
	"testing"
	"time"
)

func serveStatus(t *testing.T, body string) string {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { listener.Close() })

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		reader := bufio.NewReader(conn)
		for range 2 {
			length, err := readVarInt(reader)
			if err != nil {
				return
			}
			if _, err := reader.Discard(int(length)); err != nil {
				return
			}
		}

		payload := append([]byte{0x00}, appendVarInt(nil, int32(len(body)))...)
		_ = writePacket(conn, append(payload, body...))
	}()

	return listener.Addr().String()
}

func TestPingReadsPlayers(t *testing.T) {
	address := serveStatus(t, `{"version":{"name":"1.21"},"players":{"max":20,"online":3,"sample":[]}}`)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	players, err := Ping(ctx, address)
	if err != nil {
		t.Fatal(err)
	}

	if players == nil || *players != (Players{Online: 3, Max: 20}) {
		t.Fatalf("players = %+v", players)
	}
}

func TestPingAcceptsStatusWithoutPlayers(t *testing.T) {
	address := serveStatus(t, `{"version":{"name":"1.21"}}`)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	players, err := Ping(ctx, address)
	if err != nil || players != nil {
		t.Fatalf("players = %+v, err = %v", players, err)
	}
}

func TestPingFailsWhenNothingListens(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := listener.Addr().String()
	listener.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if _, err := Ping(ctx, address); err == nil {
		t.Fatal("expected an error")
	}
}

func TestVarIntRoundTrip(t *testing.T) {
	for _, value := range []int32{0, 1, 127, 128, 25565, 1 << 20, -1} {
		got, err := readVarInt(bytes.NewReader(appendVarInt(nil, value)))
		if err != nil || got != value {
			t.Fatalf("round trip of %d = %d, %v", value, got, err)
		}
	}
}
