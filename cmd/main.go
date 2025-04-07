package main

import (
	"fmt"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/healthcheck"
	"net/http"
	"os"
)

func main() {
	router := http.NewServeMux()

	server := http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	healthcheckHandlers := healthcheck.NewHealthcheckHandlers()
	healthcheckHandlers.RegisterHealthcheckRoutes(router)

	fmt.Println("Starting server at port :8080")
	err := server.ListenAndServe()
	if err != nil {
		fmt.Println("Failed to start application server")
		os.Exit(1)
	}
}
