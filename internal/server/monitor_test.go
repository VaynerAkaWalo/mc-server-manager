package server

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/VaynerAkaWalo/mc-server-manager/internal/mcping"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/store"
	"github.com/VaynerAkaWalo/mc-server-manager/pkg/server"
)

func monitoredService(status string, ping pingFunc) (*Service, *memoryStore, store.ServerRecord) {
	record := testRecord(time.Now().Add(time.Hour))
	record.Status = status
	db := &memoryStore{records: []store.ServerRecord{record}}
	svc := CreateServerService(nil, db, testVersions)
	svc.ping = ping

	return &svc, db, record
}

func answering(players *mcping.Players) pingFunc {
	return func(context.Context, string) (*mcping.Players, error) { return players, nil }
}

func silent(context.Context, string) (*mcping.Players, error) {
	return nil, errors.New("connection refused")
}

func TestNewServerBecomesRunningAndReportsPlayers(t *testing.T) {
	svc, db, record := monitoredService("new", answering(&mcping.Players{Online: 2, Max: 10}))

	svc.checkServers(context.Background())

	if db.records[0].Status != "running" {
		t.Fatalf("status = %s", db.records[0].Status)
	}
	response, err := svc.getServer(context.Background(), record.ID)
	if err != nil {
		t.Fatal(err)
	}
	if response.Players == nil || *response.Players != (server.Players{Online: 2, Max: 10}) {
		t.Fatalf("players = %+v", response.Players)
	}
}

func TestNewServerThatDoesNotAnswerStaysNew(t *testing.T) {
	svc, db, _ := monitoredService("new", silent)

	svc.checkServers(context.Background())

	if db.records[0].Status != "new" {
		t.Fatalf("status = %s", db.records[0].Status)
	}
}

func TestRunningServerThatStopsAnsweringIsDegraded(t *testing.T) {
	svc, db, record := monitoredService("running", answering(&mcping.Players{Online: 1, Max: 10}))
	svc.checkServers(context.Background())

	svc.ping = silent
	svc.checkServers(context.Background())

	if db.records[0].Status != "degraded" {
		t.Fatalf("status = %s", db.records[0].Status)
	}
	if svc.players.get(record.ID) != nil {
		t.Fatal("players of an unresponsive server should be unknown")
	}
}

func TestDegradedServerRecoversWhenItAnswersAgain(t *testing.T) {
	svc, db, _ := monitoredService("degraded", answering(&mcping.Players{Max: 10}))

	svc.checkServers(context.Background())

	if db.records[0].Status != "running" {
		t.Fatalf("status = %s", db.records[0].Status)
	}
}

func TestPlayersOfInactiveServersAreDropped(t *testing.T) {
	svc, db, record := monitoredService("running", answering(&mcping.Players{Online: 1, Max: 10}))
	svc.checkServers(context.Background())

	db.records[0].Status = "stopped"
	svc.checkServers(context.Background())

	if svc.players.get(record.ID) != nil {
		t.Fatal("players of a stopped server should be dropped")
	}
}

func TestServerHidingPlayersIsRunningWithUnknownPlayers(t *testing.T) {
	svc, db, record := monitoredService("new", answering(nil))

	svc.checkServers(context.Background())

	if db.records[0].Status != "running" {
		t.Fatalf("status = %s", db.records[0].Status)
	}
	if svc.players.get(record.ID) != nil {
		t.Fatal("players should be unknown")
	}
}

func TestEmptyServerStaysIdle(t *testing.T) {
	svc, db, _ := monitoredService("running", answering(&mcping.Players{Online: 0, Max: 10}))
	lastActive := time.Now().Add(-time.Hour)
	db.records[0].LastActiveAt = &lastActive

	svc.checkServers(context.Background())

	if !db.records[0].LastActiveAt.Equal(lastActive) {
		t.Fatalf("last active = %v", db.records[0].LastActiveAt)
	}
}

func TestServerWithPlayersIsActive(t *testing.T) {
	svc, db, _ := monitoredService("running", answering(&mcping.Players{Online: 1, Max: 10}))
	lastActive := time.Now().Add(-time.Hour)
	db.records[0].LastActiveAt = &lastActive

	svc.checkServers(context.Background())

	if time.Since(*db.records[0].LastActiveAt) > time.Minute {
		t.Fatalf("last active = %v", db.records[0].LastActiveAt)
	}
}

func TestServerHidingPlayersNeverGoesIdle(t *testing.T) {
	svc, db, _ := monitoredService("running", answering(nil))
	lastActive := time.Now().Add(-time.Hour)
	db.records[0].LastActiveAt = &lastActive

	svc.checkServers(context.Background())

	if time.Since(*db.records[0].LastActiveAt) > time.Minute {
		t.Fatalf("last active = %v", db.records[0].LastActiveAt)
	}
}

func TestStartedServerBecomesActive(t *testing.T) {
	svc, db, _ := monitoredService("new", answering(&mcping.Players{Online: 0, Max: 10}))

	svc.checkServers(context.Background())

	if db.records[0].LastActiveAt == nil {
		t.Fatal("a server that started should count as active")
	}
}

func idleFor(db *memoryStore, duration time.Duration) {
	lastActive := time.Now().Add(-duration)
	db.records[0].LastActiveAt = &lastActive
}

func TestServerWithoutPlayersBecomesExpiring(t *testing.T) {
	svc, db, _ := monitoredService("running", answering(&mcping.Players{Max: 10}))
	idleFor(db, 16*time.Minute)

	svc.checkServers(context.Background())

	if db.records[0].Status != "expiring" {
		t.Fatalf("status = %s", db.records[0].Status)
	}
}

func TestExpiringServerWithoutPlayersExpires(t *testing.T) {
	svc, db, _ := monitoredService("expiring", answering(&mcping.Players{Max: 10}))
	idleFor(db, 31*time.Minute)

	svc.checkServers(context.Background())

	if db.records[0].Status != "expired" {
		t.Fatalf("status = %s", db.records[0].Status)
	}
}

func TestExpiringServerRunsAgainWhenPlayersJoin(t *testing.T) {
	svc, db, _ := monitoredService("expiring", answering(&mcping.Players{Online: 1, Max: 10}))
	idleFor(db, 20*time.Minute)

	svc.checkServers(context.Background())

	if db.records[0].Status != "running" {
		t.Fatalf("status = %s", db.records[0].Status)
	}
}

func TestExpiringServerStaysExpiringWhileEmpty(t *testing.T) {
	svc, db, _ := monitoredService("expiring", answering(&mcping.Players{Max: 10}))
	idleFor(db, 20*time.Minute)

	svc.checkServers(context.Background())

	if db.records[0].Status != "expiring" {
		t.Fatalf("status = %s", db.records[0].Status)
	}
}

func TestServerThatNeverStartsExpires(t *testing.T) {
	svc, db, _ := monitoredService("new", silent)
	db.records[0].CreatedAt = time.Now().Add(-61 * time.Minute)

	svc.checkServers(context.Background())

	if db.records[0].Status != "expired" {
		t.Fatalf("status = %s", db.records[0].Status)
	}
}
