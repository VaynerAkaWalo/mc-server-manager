package bucket

import (
	"context"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

type fakeStore struct {
	mu      sync.Mutex
	objects map[string]bool
}

func (f *fakeStore) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()

	key := strings.TrimPrefix(r.URL.Path, "/saves-bucket/")
	switch {
	case r.Method == http.MethodDelete:
		delete(f.objects, key)
		w.WriteHeader(http.StatusNoContent)
	case r.Method == http.MethodGet && r.URL.Query().Get("list-type") == "2":
		prefix := r.URL.Query().Get("prefix")
		type object struct{ Key string }
		result := struct {
			XMLName  xml.Name `xml:"ListBucketResult"`
			Contents []object
		}{}
		for stored := range f.objects {
			if strings.HasPrefix(stored, prefix) {
				result.Contents = append(result.Contents, object{Key: stored})
			}
		}
		w.Header().Set("Content-Type", "application/xml")
		_ = xml.NewEncoder(w).Encode(result)
	default:
		http.NotFound(w, r)
	}
}

func TestPurgeRemovesOnlyTheObjectsOfOneSave(t *testing.T) {
	objects := &fakeStore{objects: map[string]bool{
		"saves/4f9d0a3e/config":         true,
		"saves/4f9d0a3e/data/ab/abcdef": true,
		"saves/4f9d0a3e/snapshots/1":    true,
		"saves/4f9d0a3e-other/config":   true,
		"saves/11111111/config":         true,
	}}
	api := httptest.NewServer(objects)
	t.Cleanup(api.Close)
	client, err := New(Config{Endpoint: api.URL, Bucket: "saves-bucket", Region: "eu-frankfurt-1", AccessKey: "key", SecretKey: "secret"})
	if err != nil {
		t.Fatal(err)
	}

	if err := client.Purge(context.Background(), "saves/4f9d0a3e"); err != nil {
		t.Fatal(err)
	}

	if len(objects.objects) != 2 || !objects.objects["saves/4f9d0a3e-other/config"] || !objects.objects["saves/11111111/config"] {
		t.Fatalf("objects left = %v", objects.objects)
	}
}

func TestNewRequiresEverySetting(t *testing.T) {
	full := Config{Endpoint: "https://ns.compat.objectstorage.eu-frankfurt-1.oraclecloud.com", Bucket: "b", Region: "r", AccessKey: "k", SecretKey: "s"}
	if _, err := New(full); err != nil {
		t.Fatal(err)
	}

	for name, config := range map[string]Config{
		"endpoint": {Bucket: "b", Region: "r", AccessKey: "k", SecretKey: "s"},
		"bucket":   {Endpoint: full.Endpoint, Region: "r", AccessKey: "k", SecretKey: "s"},
		"region":   {Endpoint: full.Endpoint, Bucket: "b", AccessKey: "k", SecretKey: "s"},
		"key":      {Endpoint: full.Endpoint, Bucket: "b", Region: "r", SecretKey: "s"},
		"secret":   {Endpoint: full.Endpoint, Bucket: "b", Region: "r", AccessKey: "k"},
	} {
		if _, err := New(config); err == nil {
			t.Errorf("missing %s was accepted", name)
		}
	}
}
