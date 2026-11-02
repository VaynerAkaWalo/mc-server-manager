package cluster

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

func serviceAgainst(t *testing.T, handler http.HandlerFunc) *Service {
	t.Helper()

	api := httptest.NewServer(handler)
	t.Cleanup(api.Close)

	client, err := dynamic.NewForConfig(&rest.Config{Host: api.URL})
	if err != nil {
		t.Fatal(err)
	}

	return &Service{dynamicClient: *client}
}

func TestLateDeleteDoesNotRemoveNewerServerThatReusedTheName(t *testing.T) {
	var precondition struct {
		Preconditions struct {
			UID types.UID `json:"uid"`
		} `json:"preconditions"`
	}
	service := serviceAgainst(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &precondition)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"kind":"Status","apiVersion":"v1","status":"Failure","reason":"Conflict","code":409}`))
	})

	if err := service.DeleteServer("alpha", "old-uid"); err != nil {
		t.Fatalf("a delete rejected by the UID precondition should be ignored, got %v", err)
	}
	if precondition.Preconditions.UID != "old-uid" {
		t.Errorf("delete must carry the UID of the server it meant to remove, got %q", precondition.Preconditions.UID)
	}
}
