package jobs

import (
	"context"
	"log/slog"
	"sort"
)

type Job func(ctx context.Context) error

func Run(ctx context.Context, name string, jobs map[string]Job) int {
	job, ok := jobs[name]
	if !ok {
		slog.Error("Unknown job", "job", name, "jobs", names(jobs))
		return 1
	}

	if err := job(ctx); err != nil {
		slog.Error("Job failed", "job", name, "error", err)
		return 1
	}

	return 0
}

func names(jobs map[string]Job) []string {
	result := make([]string, 0, len(jobs))
	for name := range jobs {
		result = append(result, name)
	}
	sort.Strings(result)

	return result
}
