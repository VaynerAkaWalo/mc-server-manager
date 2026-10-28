package main

import (
	"context"
	"github.com/VaynerAkaWalo/go-toolkit/xhttp"
	"github.com/VaynerAkaWalo/go-toolkit/xlog"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/auth"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/catalog"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/cluster"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/frontend"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/internalapi"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/server"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/store"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/versions"
	"log/slog"
	"os"
)

func main() {
	slog.SetDefault(slog.New(xlog.NewPreConfiguredHandler()))

	clientLoader := cluster.CreateClientLoader()
	clusterClient, err := clientLoader.Client()
	if err != nil {
		slog.Error("Failed to create cluster client")
		os.Exit(-1)
	}
	dynamicClient, err := clientLoader.DynamicClient()
	if err != nil {
		slog.Error("Failed to create crd cluster client")
		os.Exit(-1)
	}

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		slog.Error("DB_URL is not set")
		os.Exit(-1)
	}
	serverStore, err := store.Connect(context.Background(), dbURL)
	if err != nil {
		slog.Error("Failed to connect to database: " + err.Error())
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

	gameVersions := versions.NewSource()

	serverService := server.CreateServerService(cluster.CreateClusterService(*clusterClient, *dynamicClient), serverStore, gameVersions)
	serverHandlers := server.NewServerHandlers(serverService)

	go serverService.RunMonitor(context.Background())

	frontendHandlers := frontend.NewFrontendHandlers(os.DirFS("static"))

	internalHandlers := internalapi.NewHandlers(serverStore.Ping, &serverService)

	publicServer := xhttp.Server{
		Addr:     ":8080",
		Handlers: []xhttp.RouteHandler{authHandlers, serverHandlers, catalog.NewCatalogHandlers(gameVersions), frontendHandlers},
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
