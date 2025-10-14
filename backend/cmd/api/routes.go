package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (app *application) routes() http.Handler {
	r := chi.NewRouter()

	r.Get("/", app.simpleHandler)
	r.Get("/v1/healthcheck", app.healthCheckHandler)
	r.Get("/v1/movie/{movieID}", app.getMovieHandler)
	r.Get("/v1/movie", app.getAllMoviesHandler)

	r.Post("/v1/movie", app.postMovieHandler)
	r.Post("/v1/users", app.registerUserHandler)

	r.Patch("/v1/movie/{movieID}", app.updateMovieHandler)

	r.Delete("/v1/movie/{movieID}", app.deleteMovieHandler)
	return r
}
