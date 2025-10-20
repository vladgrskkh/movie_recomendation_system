package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (app *application) routes() http.Handler {
	r := chi.NewRouter()

	r.Get("/", app.simpleHandler)
	r.Get("/v1/healthcheck", app.healthCheckHandler)
	r.Get("/v1/movie/{movieID}", app.requireAuthenticatedUser(app.getMovieHandler))
	r.Get("/v1/movie", app.requireAuthenticatedUser(app.getAllMoviesHandler))

	r.Post("/v1/movie", app.requireActivatedUser(app.postMovieHandler))
	r.Post("/v1/users", app.registerUserHandler)
	r.Put("/v1/tokens/refresh", app.refreshTokenHandler)
	r.Post("/v1/tokens/authentication", app.createAuthenticationTokenHandler)

	r.Put("/v1/users/activate", app.activateUserHandler)

	r.Patch("/v1/movie/{movieID}", app.updateMovieHandler)

	r.Delete("/v1/movie/{movieID}", app.deleteMovieHandler)

	return app.recoverPanic(app.authentication(r))
}
