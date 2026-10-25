package problem

import (
	"encoding/json"
	"net/http"
)

const ContentType = "application/problem+json"

type Problem struct {
	Type   string `json:"type"`
	Title  string `json:"title"`
	Status int    `json:"status"`
	Detail string `json:"detail,omitempty"`
}

func Write(w http.ResponseWriter, status int, detail string) error {
	w.Header().Set("Content-Type", ContentType)
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(Problem{
		Type:   "about:blank",
		Title:  http.StatusText(status),
		Status: status,
		Detail: detail,
	})
}
