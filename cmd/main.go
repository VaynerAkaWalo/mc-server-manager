package main

import (
	"github.com/VaynerAkaWalo/go-toolkit/xhttp"
	"github.com/VaynerAkaWalo/go-toolkit/xlog"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/cluster"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/healthcheck"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/server"
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

	serverService := server.CreateServerService(cluster.CreateClusterService(*clusterClient, *dynamicClient))
	serverHandlers := server.NewServerHandlers(serverService)

	healthcheckHandlers := healthcheck.NewHealthcheckHandlers()

	httpServer := xhttp.Server{
		Addr:     ":8080",
		Handlers: []xhttp.RouteHandler{healthcheckHandlers, serverHandlers},
	}

	slog.Error(httpServer.ListenAndServe().Error())
}
