package main

import (
	"context"
	"github.com/VaynerAkaWalo/go-toolkit/xlog"
	"log/slog"
	"os"
)

func main() {
	slog.SetDefault(slog.New(xlog.NewPreConfiguredHandler()))

	if len(os.Args) > 1 && os.Args[1] == "job" {
		name := ""
		if len(os.Args) > 2 {
			name = os.Args[2]
		}
		os.Exit(runJob(context.Background(), name))
	}

	runServer()
}
