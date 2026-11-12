package jobs

import (
	"context"
	"errors"
	"testing"
)

func TestRunExitsZeroWhenTheNamedJobSucceeds(t *testing.T) {
	ran := false
	jobs := map[string]Job{"a": func(context.Context) error { ran = true; return nil }}

	if code := Run(context.Background(), "a", jobs); code != 0 || !ran {
		t.Fatalf("code = %d, ran = %v", code, ran)
	}
}

func TestRunExitsNonZeroWhenTheJobFails(t *testing.T) {
	jobs := map[string]Job{"a": func(context.Context) error { return errors.New("boom") }}

	if code := Run(context.Background(), "a", jobs); code == 0 {
		t.Fatal("a failed job must exit non-zero so Kubernetes retries it")
	}
}

func TestRunExitsNonZeroForAnUnknownJobWithoutRunningOthers(t *testing.T) {
	ran := false
	jobs := map[string]Job{"a": func(context.Context) error { ran = true; return nil }}

	if code := Run(context.Background(), "b", jobs); code == 0 || ran {
		t.Fatalf("code = %d, ran = %v", code, ran)
	}
}
