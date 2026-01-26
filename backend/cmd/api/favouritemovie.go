package main

import (
	"errors"
	"net/http"

	"github.com/vladgrskkh/movie_recomendation_system/internal/data"
)

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

func (app *application) unlikeMovieHandler(w http.ResponseWriter, r *http.Request) {
	movieID, err := app.readIDParam(r)
	if err != nil {
		app.serverErrorResponse(w, r, err)
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
