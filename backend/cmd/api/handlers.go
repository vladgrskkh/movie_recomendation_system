package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/invopop/validation"
	"github.com/vladgrskkh/movie_recomendation_system/internal/data"
	"github.com/vladgrskkh/movie_recomendation_system/internal/validate"
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

func (app *application) getMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.logger.Println(err)
		return
	}

	movie, err := app.db.Get(id)
	if err != nil {
		app.logger.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = app.writeJSON(w, http.StatusOK, map[string]interface{}{"movie": movie}, nil)

	if err != nil {
		app.logger.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

}

// admin only
func (app *application) postMovieHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title   string   `json:"title"`
		Year    int32    `json:"year"`
		Runtime int32    `json:"runtime"`
		Genres  []string `json:"genres"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	movie := &data.Movie{
		Title:   input.Title,
		Year:    input.Year,
		Runtime: input.Runtime,
		Genres:  input.Genres,
	}

	// Validation of the movie input
	err = validation.ValidateStruct(movie,
		validation.Field(&movie.Title, validation.Required, validation.Length(1, 500)),
		validation.Field(&movie.Year, validation.Required, validation.Min(1888), validation.Max(int32(time.Now().Year()))),
		validation.Field(&movie.Runtime, validation.Required, validation.Min(1)),
		validation.Field(&movie.Genres, validation.Required, validation.Length(1, 5), validation.By(validate.Unique(movie.Genres))),
	)
	if err != nil {
		// TODO: improve error handling
		app.logger.Println(err)
		http.Error(w, "Wrong json input", http.StatusBadRequest)
	}

	err = app.db.Insert(movie)
	if err != nil {
		app.logger.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/movie/%d", movie.ID))

	err = app.writeJSON(w, http.StatusCreated, map[string]interface{}{"input": input}, headers)
	if err != nil {
		app.logger.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

}

// admin only
func (app *application) deleteMovieHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Work in progress..."))
	w.WriteHeader(http.StatusNotImplemented)
}
