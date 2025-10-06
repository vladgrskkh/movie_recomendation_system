package main

import (
	"encoding/json"
	"net/http"
)

// writeJSON is a helper method for writing JSON responses
func (app *application) writeJSON(w http.ResponseWriter, status int, data map[string]interface{}, headers http.Header) error {
	// Convert the data to JSON
	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	js = append(js, '\n')

	// Add provided headers
	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}
