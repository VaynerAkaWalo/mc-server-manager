package main

import (
	"context"
	"github.com/VaynerAkaWalo/go-toolkit/xhttp"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/auth"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/catalog"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/frontend"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/internalapi"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/save"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/server"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/versions"
	"log/slog"
	"os"
)

func runServer() {
	serverStore, err := connectStore(context.Background())
	if err != nil {
		slog.Error(err.Error())
		os.Exit(-1)
	}
	defer serverStore.Close()

	err = serverStore.Migrate(context.Background())
	if err != nil {
		slog.Error("Failed to migrate database: " + err.Error())
		os.Exit(-1)
	}

	authHandlers, err := auth.NewHandlers(os.Getenv("OIDC_ISSUER"), os.Getenv("OIDC_CLIENT_ID"), os.Getenv("APP_URL"), serverStore)
	if err != nil {
		slog.Error("Failed to configure authentication", "error", err)
		os.Exit(1)
	}
	go authHandlers.RunCleanup(context.Background())

	saveBucket, err := newSaveBucket()
	if err != nil {
		slog.Error("Failed to configure the saves bucket", "error", err)
		os.Exit(1)
	}
	saveService := save.NewService(serverStore, saveBucket)

	gameVersions := versions.NewSource()

	clusterService, err := newClusterService()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(-1)
	}
	serverService := server.CreateServerService(clusterService, serverStore, gameVersions)
	serverHandlers := server.NewServerHandlers(serverService, authHandlers)

	go serverService.RunMonitor(context.Background())

	frontendHandlers := frontend.NewFrontendHandlers(os.DirFS("static"))

	internalHandlers := internalapi.NewHandlers(serverStore.Ping, &serverService, saveService)

	publicServer := xhttp.Server{
		Addr:     ":8080",
		Handlers: []xhttp.RouteHandler{authHandlers, serverHandlers, save.NewHandlers(saveService, authHandlers), catalog.NewCatalogHandlers(gameVersions), frontendHandlers},
	}
	internalServer := xhttp.Server{
		Addr:     ":8081",
		Handlers: []xhttp.RouteHandler{internalHandlers},
	}

	failed := make(chan error, 2)
	go func() { failed <- publicServer.ListenAndServe() }()
	go func() { failed <- internalServer.ListenAndServe() }()

	slog.Error((<-failed).Error())
	os.Exit(1)
}
