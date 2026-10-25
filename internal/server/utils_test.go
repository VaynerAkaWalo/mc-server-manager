package server

import (
	"errors"
	"testing"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestRejectedByCluster(t *testing.T) {
	gr := schema.GroupResource{Resource: "mcservers"}

	cases := map[string]struct {
		err      error
		rejected bool
	}{
		"already exists": {apierrors.NewAlreadyExists(gr, "a"), true},
		"forbidden":      {apierrors.NewForbidden(gr, "a", errors.New("no")), true},
		"timeout":        {apierrors.NewTimeoutError("slow", 1), false},
		"server error":   {apierrors.NewInternalError(errors.New("boom")), false},
		"connection":     {errors.New("connection reset by peer"), false},
	}

	for name, c := range cases {
		if got := rejectedByCluster(c.err); got != c.rejected {
			t.Errorf("%s: expected %v, got %v", name, c.rejected, got)
		}
	}
}
