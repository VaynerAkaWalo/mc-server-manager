package cluster

import (
	"encoding/json"
	serversv1alpha1 "github.com/VaynerAkaWalo/mc-server-operator/api/v1alpha1"
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

func TestDeployedServerCarriesTheSaveOnlyWhenThereIsOne(t *testing.T) {
	var created map[string]any
	service := serviceAgainst(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &created)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write(body)
	})

	spec := serversv1alpha1.McServerSpec{Name: "alpha", Save: &serversv1alpha1.SaveSpec{ID: "4f9d0a3e-7c1b-4d6f-9a52-0e1c8b7d2a10"}}
	if err := service.DeployServerSpec(spec); err != nil {
		t.Fatal(err)
	}
	if got := created["spec"].(map[string]any)["save"]; got == nil || got.(map[string]any)["id"] != spec.Save.ID {
		t.Fatalf("save = %v", got)
	}

	spec.Save = nil
	if err := service.DeployServerSpec(spec); err != nil {
		t.Fatal(err)
	}
	if _, present := created["spec"].(map[string]any)["save"]; present {
		t.Fatalf("spec = %v", created["spec"])
	}
}
