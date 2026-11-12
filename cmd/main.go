package main

import (
	"context"
	"github.com/VaynerAkaWalo/go-toolkit/xlog"
	"log/slog"
	"os"
)

func main() {
	slog.SetDefault(slog.New(xlog.NewPreConfiguredHandler()))

	if name, rest, ok := splitJobArgs(os.Args[1:]); ok {
		os.Args = append(os.Args[:1], rest...)

		os.Exit(runJob(context.Background(), name))
	}

	runServer()
}

func splitJobArgs(args []string) (name string, rest []string, ok bool) {
	if len(args) == 0 || args[0] != "job" {
		return "", nil, false
	}

	if len(args) == 1 {
		return "", nil, true
	}

	return args[1], args[2:], true
}
