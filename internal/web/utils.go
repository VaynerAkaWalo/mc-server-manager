package web

import (
	"encoding/json"
	"net/http"
)

func SendJsonResponse(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func SendErrorResponse(w http.ResponseWriter, code int, message string) {
	response := ErrorHttpResponse{
		Error: message,
	}

	SendJsonResponse(w, code, response)
}
