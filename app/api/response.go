package api

import (
	"encoding/json"
	"log"
	"net/http"
)

// RespondWithJSON is a reusable helper function to set headers and encode JSON.
func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	// Marshal the payload into JSON bytes
	data, err := json.Marshal(payload)
	if err != nil {
		// Log the error internally and send a generic 500 error response
		log.Printf("ERROR: failed to marshal JSON response: %v", err)
		http.Error(w, "Internal server error formatting response.", http.StatusInternalServerError)
		return
	}

	// Write headers and status code
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	// Write the JSON bytes
	if _, err := w.Write(data); err != nil {
		// Log write errors (client disconnect, etc.)
		log.Printf("ERROR: failed to write response body: %v", err)
	}
}
