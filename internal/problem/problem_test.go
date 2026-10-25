package problem

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWrite(t *testing.T) {
	rec := httptest.NewRecorder()

	if err := Write(rec, http.StatusConflict, "name in use"); err != nil {
		t.Fatal(err)
	}

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d", rec.Code)
	}
	if got := rec.Header().Get("Content-Type"); got != ContentType {
		t.Fatalf("content type = %q", got)
	}

	var body Problem
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	want := Problem{Type: "about:blank", Title: "Conflict", Status: 409, Detail: "name in use"}
	if body != want {
		t.Fatalf("body = %+v", body)
	}
}
