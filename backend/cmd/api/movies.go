package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	pb "github.com/vladgrskkh/movie-recommender-contracts/v1/predict"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/vladgrskkh/movie_recomendation_system/internal/data"
)

const (
	// Param that regulates number of movies from recommendation service
	topK = 10
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
	// FIXME: 2005 here is bc ml model is trained only on movies older than 2008
	movies, err := app.models.Movies.GetNew(2005)
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

type moviesResponse struct {
	Movies []*data.Movie `json:"movies"`
}

// getMoviesRecommended godoc
//
// @Summary List recommended movies
// @Description Retrieve a list of recommended movies for user
// @Tags movies
// @Produce json
// @Security BearerAuth
// @Success 200 {object} moviesResponse
// @Failure 401 {object} map[string]string "Unauthorized | Example {"error": "this resourse avaliable only for authenticated users"}"
// @Failure 404 {object} map[string]string "Not Found | Example {"error": "requested resource could not be found"}"
// @Failure 500 {object} map[string]string "Internal Server Error | Example {"error": "server encountered a problem and could not process your request"}"
// @Router /movies/recommended [get]
func (app *application) getMoviesRecommended(w http.ResponseWriter, r *http.Request) {
	// So I can do this two ways:
	// 1. check for recommended movies and if nothing found i than go and search for movies user watched
	// than I go to predict service for recommendations, returning it to user and store it in db
	// 2. same as 1 but I can generate recommendations not just when recommnedations is empty
	// problem with first approach is that I need to somehow retain old recommendations
	// problem with second is in his nature and I also need to mechanism for retaining old recommendations
	// Maybe for recommended movies I just use redis (solves problem with old recommendations also fast)

	// fetch from redis here
	user := app.contextGetUser(r)

	// this should not trigger
	if user.IsAnonymous() {
		app.authenticationRequiredResponse(w, r)
		return
	}

	movies, err := app.models.RecommendedMovies.Get(user.ID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.logger.Info("creating recommendations for user", slog.Int64("userID", user.ID))
		default:
			app.serverErrorResponse(w, r, err)
			return
		}
	}

	if len(movies) != 0 {
		err = app.writeJSON(w, http.StatusOK, moviesResponse{Movies: movies}, nil)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		return
	}

	watchedMovies, err := app.models.Movies.GetWatched(user.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	resp, err := app.predictClient.Recommend(context.Background(), &pb.RecommendRequest{MovieID: watchedMovies, TopK: topK})
	if err != nil {
		switch status.Code(err) {
		case codes.InvalidArgument:
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}

		return
	}

	if len(resp.GetRecommendations()) == 0 {
		app.serverErrorResponse(w, r, err)
		return
	}

	movieIDs := make([]int64, 0, 10)
	for _, movie := range resp.GetRecommendations() {
		movieIDs = append(movieIDs, movie.MovieID)
	}

	movies, err = app.models.Movies.GetByIDs(movieIDs)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	app.background(func() {
		err := app.models.RecommendedMovies.Set(user.ID, movies)
		if err != nil {
			app.logError(r, err)
		}
	})

	err = app.writeJSON(w, http.StatusOK, moviesResponse{Movies: movies}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
