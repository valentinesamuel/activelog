package response

import (
	"encoding/json"
	"net/http"

	"github.com/valentinesamuel/activelog/pkg/errors"
)

func SendJSON(w http.ResponseWriter, statusCode int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(data)
}

func Error(w http.ResponseWriter, statusCode int, message string) error {
	return SendJSON(w, statusCode, map[string]string{
		"error": message,
	})
}

func AppError(w http.ResponseWriter, err *errors.AppError) error {
	return SendJSON(w, err.Code, map[string]interface{}{
		"error": err.Message,
		"code":  err.Code,
	})
}
