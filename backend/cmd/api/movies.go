package main

import (
	"net/http"
)

// getMoviesPopular godoc
//
// @Summary List popular movies
// @Description Retrieve a list of popular movies (table in db with manual updates)
// @Tags movies
// @Produce json
// @Security BearerAuth
// @Success 200 {object} []data.Movie
// @Failure 401 {object} map[string]string "Unauthorized | Example {"error": "this resourse avaliable only for authenticated users"}"
// @Failure 404 {object} map[string]string "Not Found | Example {"error": "requested resource could not be found"}"
// @Failure 500 {object} map[string]string "Internal Server Error | Example {"error": "server encountered a problem and could not process your request"}"
// @Router /movies/popular [get]
func (app *application) getMoviesPopular(w http.ResponseWriter, r *http.Request) {
	movies, err := app.models.Movies.GetPopular()
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// just in case, this should not trigger
	if len(movies) == 0 {
		app.notFoundResponse(w, r)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"movies": movies}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// getMoviesNew godoc
//
// @Summary List new movies
// @Description Retrieve a list of new movies (20 movies by certain year)
// @Tags movies
// @Produce json
// @Security BearerAuth
// @Success 200 {object} []data.Movie
// @Failure 401 {object} map[string]string "Unauthorized | Example {"error": "this resourse avaliable only for authenticated users"}"
// @Failure 404 {object} map[string]string "Not Found | Example {"error": "requested resource could not be found"}"
// @Failure 500 {object} map[string]string "Internal Server Error | Example {"error": "server encountered a problem and could not process your request"}"
// @Router /movies/new [get]
func (app *application) getMoviesNew(w http.ResponseWriter, r *http.Request) {
	// change year to something smarter, mb some flag, or env
	movies, err := app.models.Movies.GetNew(2025)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	if len(movies) == 0 {
		app.notFoundResponse(w, r)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"movies": movies}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// getMoviesRecommended godoc
//
// @Summary List recommended movies
// @Description Retrieve a list of recommended movies (table in db with manual updates)
// @Tags movies
// @Produce json
// @Security BearerAuth
// @Success 200 {object} []data.Movie
// @Failure 401 {object} map[string]string "Unauthorized | Example {"error": "this resourse avaliable only for authenticated users"}"
// @Failure 404 {object} map[string]string "Not Found | Example {"error": "requested resource could not be found"}"
// @Failure 500 {object} map[string]string "Internal Server Error | Example {"error": "server encountered a problem and could not process your request"}"
// @Router /movies/recommended [get]
func (app *application) getMoviesRecommended(w http.ResponseWriter, r *http.Request) {
	movies, err := app.models.Movies.GetRecommended()
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// just in case, this should not trigger
	if len(movies) == 0 {
		app.notFoundResponse(w, r)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"movies": movies}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
