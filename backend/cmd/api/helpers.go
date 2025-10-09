package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// readIDParam extracts and validates the ID parameter from the URL
func (app *application) readIDParam(r *http.Request) (int64, error) {
	movieID := chi.URLParam(r, "movieID")

	id, err := strconv.ParseInt(movieID, 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}

	return id, nil
}

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

// readJSON is a helper method for reading JSON requests
func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	maxBytes := 1_048_576 // 1 MB

	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	// TODO: improve error handling so we can return more specific errors
	err := dec.Decode(dst)
	if err != nil {
		return err
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only contain a single JSON value")
	}

	return nil
}
