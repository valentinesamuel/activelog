package response

import (
	"encoding/json"
	"net/http"
)

func SendJSON(w http.ResponseWriter, statusCode int, data interface{}) error {
	w.Header().Set("Content Type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(data)
}

func Error(w http.ResponseWriter, statusCode int, message string) error {
	return SendJSON(w, statusCode, map[string]string{
		"error": message,
	})
}
