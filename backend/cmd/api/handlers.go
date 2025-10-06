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
	var input struct {
		Id   int    `json:"id"`
		Name string `json:"name"`
		Text string `json:"text"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	err = app.writeJSON(w, http.StatusCreated, map[string]interface{}{"input": input}, make(http.Header))
	if err != nil {
		app.logger.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

}

func (app *application) deleteDataHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Work in progress..."))
	w.WriteHeader(http.StatusNotImplemented)
}
