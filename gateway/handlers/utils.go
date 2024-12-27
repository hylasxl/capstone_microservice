package handlers

import (
	"encoding/json"
	"net/http"
)

func respondWithError(w http.ResponseWriter, statusCode int, message string, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errorResponse := map[string]interface{}{
		"success": false,
		"error":   message,
	}

	if err != nil {
		errorResponse["details"] = err.Error()
	}

	err = json.NewEncoder(w).Encode(errorResponse)
	if err != nil {
		return
	}
}
