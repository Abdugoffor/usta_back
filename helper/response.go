package helper

import (
	"encoding/json"
	"log"
	"net/http"
)

func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func WriteError(w http.ResponseWriter, status int, msg string) {
	WriteJSON(w, status, map[string]string{"error": msg})
}

// WriteInternalError logs the real error server-side and returns a generic message to the client.
func WriteInternalError(w http.ResponseWriter, err error) {
	log.Printf("ERROR: %v", err)
	WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
}

func WriteValidation(w http.ResponseWriter, errs map[string]string) {
	WriteJSON(w, http.StatusUnprocessableEntity, map[string]interface{}{"errors": errs})
}
