package main

import (
	"github.com/VaynerAkaWalo/go-toolkit/xlog"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/cluster"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/healthcheck"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/server"
	"log/slog"
	"net/http"
	"os"
)

func main() {
	slog.SetDefault(slog.New(xlog.NewPreConfiguredHandler()))

	router := http.NewServeMux()

	httpServer := http.Server{
		Addr:    ":8080",
		Handler: router,
	}

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

	serverService := server.CreateServerService(cluster.CreateClusterService(*clusterClient, *dynamicClient))
	serverHandlers := server.NewServerHandlers(serverService)
	serverHandlers.RegisterServerRoutes(router)

	healthcheckHandlers := healthcheck.NewHealthcheckHandlers()
	healthcheckHandlers.RegisterHealthcheckRoutes(router)

	slog.Info("Starting server at port :8080")
	err = httpServer.ListenAndServe()
	if err != nil {
		slog.Error("Failed to start application server")
		os.Exit(1)
	}
}
