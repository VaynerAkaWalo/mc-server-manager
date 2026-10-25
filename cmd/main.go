package main

import (
	"context"
	"github.com/VaynerAkaWalo/go-toolkit/xhttp"
	"github.com/VaynerAkaWalo/go-toolkit/xlog"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/catalog"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/cluster"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/frontend"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/healthcheck"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/internalapi"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/server"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/store"
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

	serverService := server.CreateServerService(cluster.CreateClusterService(*clusterClient, *dynamicClient), serverStore)
	serverHandlers := server.NewServerHandlers(serverService)

	healthcheckHandlers := healthcheck.NewHealthcheckHandlers()
	frontendHandlers := frontend.NewFrontendHandlers(os.DirFS("static"))

	internalHandlers := internalapi.NewHandlers(serverStore.Ping, &serverService)

	// The public /health stays until the probes in the deployments repo point at the internal listener.
	publicServer := xhttp.Server{
		Addr:     ":8080",
		Handlers: []xhttp.RouteHandler{healthcheckHandlers, serverHandlers, catalog.NewCatalogHandlers(), frontendHandlers},
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
