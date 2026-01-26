package main

import (
	"errors"
	"net/http"

	"github.com/vladgrskkh/movie_recomendation_system/internal/data"
)

// likeMovieHandler godoc
//
// @Summary Like a movie
// @Description Functionality for liking movie
// @Tags movies
// @Accept json
// @Produce json
// @Success 201 {object} map[string]string "OK | Example {"message": "successfully liked movie"}
// @Failure 400 {object} map[string]string "Bad Request | Example {"error": "body contains badly-formated JSON"}"
// @Failure 403 {object} map[string]string "Forbidden | Example {"error": "your account must be activated to access this resourse"}"
// @Failure 500 {object} map[string]string "Internal Server Error | Example {"error": "server encountered a problem and could not process your request"}"
// @Security BearerAuth
// @Router /movies/{movieID}/like [put]
func (app *application) likeMovieHandler(w http.ResponseWriter, r *http.Request) {
	movieID, err := app.readIDParam(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := app.contextGetUser(r)

	// TODO: implement movies method for inserting liked movie
	err = app.models.Movies.InsertFavourite(movieID, user.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	message := "successfully liked movie with id %d"
	err = app.writeJSON(w, http.StatusOK, envelope{"message": message}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// unlikeMovieHandler godoc
//
// @Summary Like a movie
// @Description Functionality for liking movie
// @Tags movies
// @Accept json
// @Produce json
// @Success 201 {object} map[string]string "OK | Example {"message": "successfully unliked movie"}
// @Failure 400 {object} map[string]string "Bad Request | Example {"error": "body contains badly-formated JSON"}"
// @Failure 403 {object} map[string]string "Forbidden | Example {"error": "your account must be activated to access this resourse"}"
// @Failure 404 {object} map[string]string "Not Found | Example {"error": "requested resource could not be found"}"
// @Failure 500 {object} map[string]string "Internal Server Error | Example {"error": "server encountered a problem and could not process your request"}"
// @Security BearerAuth
// @Router /movies/{movieID}/unlike [put]
func (app *application) unlikeMovieHandler(w http.ResponseWriter, r *http.Request) {
	movieID, err := app.readIDParam(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := app.contextGetUser(r)

	// TODO: implement movies method for deleting liked movie
	err = app.models.Movies.DeleteFavourite(movieID, user.ID)
	if err != nil {
		switch {
		// NOTE: maybe use here validation error or bad request idk
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}

		return
	}

	message := "successfully unliked movie"
	err = app.writeJSON(w, http.StatusOK, envelope{"message": message}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
