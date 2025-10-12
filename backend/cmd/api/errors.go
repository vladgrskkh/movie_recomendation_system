package main

import (
	"log/slog"
	"net/http"
)

func (app *application) logError(r *http.Request, err error) {
	app.logger.Error(err.Error(), slog.String("request_method", r.Method),
		slog.String("request_url", r.URL.String()))
}

func (app *application) errorResponse(w http.ResponseWriter, r *http.Request, status int, message interface{}) {
	data := map[string]interface{}{
		"error": message,
	}

	err := app.writeJSON(w, status, data, nil)
	if err != nil {
		app.logError(r, err)
		w.WriteHeader(500)
	}
}

func (app *application) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logError(r, err)

	message := "server encountered a problem and could not process your request"
	app.errorResponse(w, r, http.StatusInternalServerError, message)
}

func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	message := "requested resource could not be found"
	app.errorResponse(w, r, http.StatusNotFound, message)
}

func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.errorResponse(w, r, http.StatusBadRequest, err.Error())
}

func (app *application) failedValidationResponse(w http.ResponseWriter, r *http.Request, errors error) {
	app.errorResponse(w, r, http.StatusUnprocessableEntity, errors)
}
