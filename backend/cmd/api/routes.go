package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"
	_ "github.com/vladgrskkh/movie_recomendation_system/cmd/api/docs"
)

func (app *application) routes() http.Handler {
	r := chi.NewRouter()

	r.Get("/v1/healthcheck", app.healthCheckHandler)
	r.Get("/v1/movie/{movieID}", app.requireAuthenticatedUser(app.getMovieHandler))
	r.Get("/v1/movie", app.requireAuthenticatedUser(app.getAllMoviesHandler))
	r.Get("/v1/swagger/*", httpSwagger.Handler())
	r.Post("/v1/movie", app.requireActivatedUser(app.postMovieHandler))
	r.Post("/v1/users", app.registerUserHandler)
	r.Put("/v1/tokens/refresh", app.refreshTokenHandler)
	r.Post("/v1/tokens/authentication", app.createAuthenticationTokenHandler)
	r.Post("/v1/movie/predict", app.predictHandler)

	r.Put("/v1/users/activate", app.activateUserHandler)

	r.Patch("/v1/movie/{movieID}", app.updateMovieHandler)

	r.Delete("/v1/movie/{movieID}", app.deleteMovieHandler)

	return app.recoverPanic(app.authentication(r))
}
