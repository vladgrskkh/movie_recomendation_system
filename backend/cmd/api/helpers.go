package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

var (
	ErrKeyNotInteger = errors.New("must be an integer")
)

// const (
// 	numberOfKeys = 20 // for generating kafka keys
// )

type envelope map[string]interface{}

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
func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
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
	_, err = w.Write(js)

	return err
}

// readJSON is a helper method for reading JSON requests
func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	maxBytes := 1_048_576 // 1 MB

	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formated JSON(at character %d)", syntaxError.Offset)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formated JSON")

		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type(at character %d)", unmarshalTypeError.Offset)

		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)

		case err.Error() == "http: request body too large":
			return fmt.Errorf("body must not be larger than %d bytes", maxBytes)

		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		default:
			return err
		}
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only contain a single JSON value")
	}

	return nil
}

func (app *application) background(fn func()) {
	app.wg.Add(1)
	go func() {
		defer app.wg.Done()

		defer func() {
			if err := recover(); err != nil {
				app.logger.Error(fmt.Sprint(err))
			}
		}()

		fn()
	}()
}

// readString is a helper method for retrieving a string value from a url.Values object.
// If the value is not present, it returns the defaultValue.
func (app *application) readString(qs url.Values, key string, defaultValue string) string {
	s := qs.Get(key)

	if s == "" {
		return defaultValue
	}

	return s
}

// readCSV is a helper method for retrieving a comma-separated string value from a url.Values object.
// If the value is not present, it returns the defaultValue.
// The function splits the returned string by commas and returns a slice of strings.
func (app *application) readCSV(qs url.Values, key string, defaultValue []string) []string {
	csv := qs.Get(key)

	if csv == "" {
		return defaultValue
	}

	return strings.Split(csv, ",")
}

// readInt is a helper method for retrieving an integer value from a url.Values object.
// If the value is not present, it returns the defaultValue.
// The function parses the returned string as an integer. If the parsing fails, it returns the defaultValue.
func (app *application) readInt(qs url.Values, key string, defaultValue int) (int, error) {
	s := qs.Get(key)

	if s == "" {
		return defaultValue, nil
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		return defaultValue, ErrKeyNotInteger
	}

	return i, nil
}

// readValidation is a helper method for reading validation message and converting it to a map for json response
// This method should only be used for reading validationMessage when user is registering/logging in
// Can change it in future, but best options is to change validation package or implement our own
// func (app *application) readValidation(validationMessage string) map[string]string {
// 	validationErrors := make(map[string]string)
// 	if strings.TrimSpace(validationMessage) == "" {
// 		return validationErrors
// 	}

// 	re := regexp.MustCompile(`\s*([^:;]+?)\s*:\s*([^;]+)`)
// 	matches := re.FindAllStringSubmatch(validationMessage[:len(validationMessage)-1], -1)

// 	for _, match := range matches {
// 		if len(match) < 3 {
// 			continue
// 		}

// 		field := strings.TrimSpace(match[1])
// 		msg := strings.TrimSpace(match[2])

// 		if field == "" || msg == "" {
// 			continue
// 		}

// 		// checking if field exists
// 		if _, ok := validationErrors[field]; ok {
// 			// panic here because we only use this function for reading validation when we expect unique fields
// 			panic("duplicate field in validation message")
// 		}

// 		validationErrors[field] = msg
// 	}

// 	return validationErrors
// }

// dont need this for now so comment it
// func generateUUIDString() [numberOfKeys]string {
// 	var uuids [numberOfKeys]string
// 	for i := 0; i < numberOfKeys; i++ {
// 		uuids[i] = uuid.NewString()
// 	}

// 	return uuids
// }
