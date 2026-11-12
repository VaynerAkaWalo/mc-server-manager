package main

import (
	"context"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/jobs"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/save"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/server"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/versions"
	"log/slog"
)

func runJob(ctx context.Context, name string) int {
	return jobs.Run(ctx, name, map[string]jobs.Job{
		"expire-servers": expireServers,
		"expire-saves":   expireSaves,
	})
}

func expireServers(ctx context.Context) error {
	serverStore, err := connectStore(ctx)
	if err != nil {
		return err
	}
	defer serverStore.Close()

	clusterService, err := newClusterService()
	if err != nil {
		return err
	}

	service := server.CreateServerService(clusterService, serverStore, versions.NewSource())

	names, err := service.ShutdownExpired()
	slog.Info("Expired servers", "count", len(names), "servers", names)

	return err
}

func expireSaves(ctx context.Context) error {
	serverStore, err := connectStore(ctx)
	if err != nil {
		return err
	}
	defer serverStore.Close()

	saveBucket, err := newSaveBucket()
	if err != nil {
		return err
	}

	ids, err := save.NewService(serverStore, saveBucket).ExpireIdle(ctx)
	slog.Info("Expired saves", "count", len(ids), "saves", ids)

	return err
}
