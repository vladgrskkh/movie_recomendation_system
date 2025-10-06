package main

import (
	"net/http"
)

// TO DO: JSON responses, data base connection

func (app *application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"status":  "avaliable",
		"env":     app.config.env,
		"version": version,
	}

	err := app.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		app.logger.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (app *application) getDataHandler(w http.ResponseWriter, r *http.Request) {

	w.WriteHeader(http.StatusOK)
}

func (app *application) postDataHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Work in progress..."))
	w.WriteHeader(http.StatusNotImplemented)
}

func (app *application) deleteDataHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Work in progress..."))
	w.WriteHeader(http.StatusNotImplemented)
}
