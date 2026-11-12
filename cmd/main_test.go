package main

import (
	"slices"
	"testing"
)

func TestSplitJobArgsLeavesFlagsAfterTheJobNameForTheKubeconfigLoader(t *testing.T) {
	name, rest, ok := splitJobArgs([]string{"job", "expire-servers", "-kubeconfig", "staging"})

	if !ok || name != "expire-servers" || !slices.Equal(rest, []string{"-kubeconfig", "staging"}) {
		t.Fatalf("name = %q, rest = %v, ok = %v", name, rest, ok)
	}
}

func TestSplitJobArgsIsNotJobModeWithoutTheJobCommand(t *testing.T) {
	for _, args := range [][]string{nil, {"-kubeconfig", "staging"}, {"serve"}} {
		if _, _, ok := splitJobArgs(args); ok {
			t.Fatalf("args %v must start the server", args)
		}
	}
}

func TestSplitJobArgsReportsAMissingJobName(t *testing.T) {
	name, _, ok := splitJobArgs([]string{"job"})

	if !ok || name != "" {
		t.Fatalf("name = %q, ok = %v", name, ok)
	}
}
