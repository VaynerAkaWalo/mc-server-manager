package server

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/VaynerAkaWalo/mc-server-manager/internal/mcping"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/store"
	"github.com/VaynerAkaWalo/mc-server-manager/pkg/server"
	"github.com/google/uuid"
)

const (
	monitorInterval    = 15 * time.Second
	pingTimeout        = 2 * time.Second
	maxConcurrentPings = 20
)

var idleTimeouts = store.IdleTimeouts{
	Expiring: 15 * time.Minute,
	Expired:  30 * time.Minute,
	Startup:  time.Hour,
}

type pingFunc func(ctx context.Context, address string) (*mcping.Players, error)

type playerCache struct {
	mu      sync.RWMutex
	players map[uuid.UUID]server.Players
}

func newPlayerCache() *playerCache {
	return &playerCache{players: map[uuid.UUID]server.Players{}}
}

func (c *playerCache) get(id uuid.UUID) *server.Players {
	c.mu.RLock()
	defer c.mu.RUnlock()

	players, found := c.players[id]
	if !found {
		return nil
	}

	return &players
}

func (c *playerCache) set(id uuid.UUID, players server.Players) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.players[id] = players
}

func (c *playerCache) forget(id uuid.UUID) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.players, id)
}

func (c *playerCache) retain(ids map[uuid.UUID]bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for id := range c.players {
		if !ids[id] {
			delete(c.players, id)
		}
	}
}

func (s *Service) RunMonitor(ctx context.Context) {
	ticker := time.NewTicker(monitorInterval)
	defer ticker.Stop()

	for {
		s.checkServers(ctx)

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (s *Service) checkServers(ctx context.Context) {
	records, err := s.store.ListMonitoredServers(ctx)
	if err != nil {
		slog.Error("Failed to list servers to monitor: " + err.Error())
		return
	}

	var wg sync.WaitGroup
	slots := make(chan struct{}, maxConcurrentPings)
	monitored := make(map[uuid.UUID]bool, len(records))
	for _, record := range records {
		monitored[record.ID] = true

		wg.Add(1)
		go func() {
			defer wg.Done()

			slots <- struct{}{}
			defer func() { <-slots }()

			s.checkServer(ctx, record)
		}()
	}
	wg.Wait()

	s.players.retain(monitored)
	s.record(s.store.MarkIdle(ctx, idleTimeouts))
}

func (s *Service) checkServer(ctx context.Context, record store.ServerRecord) {
	pingCtx, cancel := context.WithTimeout(ctx, pingTimeout)
	defer cancel()

	players, err := s.ping(pingCtx, clusterAddress(record.Name))
	if err != nil {
		s.players.forget(record.ID)
		if record.Status == string(server.StatusRunning) {
			slog.Warn("Server " + record.Name + " stopped responding: " + err.Error())
			s.record(s.store.MarkDegraded(ctx, record.ID))
		}
		return
	}

	if players != nil {
		s.players.set(record.ID, server.Players{Online: players.Online, Max: players.Max})
	} else {
		s.players.forget(record.ID)
	}
	if record.Status == string(server.StatusNew) || record.Status == string(server.StatusDegraded) {
		slog.Info("Server " + record.Name + " is responding")
		s.record(s.store.MarkRunning(ctx, record.ID))
	}
	// Mods can hide player counts, so a server that hides them never counts as idle.
	if players == nil || players.Online > 0 {
		s.record(s.store.MarkActive(ctx, record.ID))
	}
}

func (s *Service) record(err error) {
	if err != nil {
		slog.Error("Failed to update server status: " + err.Error())
	}
}
