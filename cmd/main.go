package main

import (
	"fmt"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/cluster"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/healthcheck"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/server"
	"log"
	"net/http"
	"os"
)

func main() {
	router := http.NewServeMux()

	httpServer := http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	clientLoader := cluster.CreateClientLoader()
	clusterClient, err := clientLoader.Client()
	if err != nil {
		log.Fatal("Failed to create cluster client")
	}
	dynamicClient, err := clientLoader.DynamicClient()
	if err != nil {
		log.Fatal("Failed to create crd cluster client")
	}

	serverService := server.CreateServerService(cluster.CreateClusterService(*clusterClient, *dynamicClient))
	serverHandlers := server.NewServerHandlers(serverService)
	serverHandlers.RegisterServerRoutes(router)

	healthcheckHandlers := healthcheck.NewHealthcheckHandlers()
	healthcheckHandlers.RegisterHealthcheckRoutes(router)

	fmt.Println("Starting server at port :8080")
	err = httpServer.ListenAndServe()
	if err != nil {
		fmt.Println("Failed to start application server")
		os.Exit(1)
	}
}
