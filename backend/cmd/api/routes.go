package main

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"

	_ "github.com/vladgrskkh/movie_recomendation_system/cmd/api/docs"
)

func (app *application) routes() http.Handler {
	r := chi.NewRouter()

	r.Use(app.metrics)
	r.Use(app.recoverPanic)
	r.Use(app.authentication)

	// Rate-limit all routes
	// Think about adding rate limmiter for specific routes(registaration, login)
	if app.config.limiter.enable {
		r.Use(httprate.LimitByIP(app.config.limiter.rps, time.Second))
	}

	r.Route("/v1", func(r chi.Router) {
		r.Get("/healthcheck", app.healthCheckHandler)
		r.Get("/swagger/*", httpSwagger.Handler())

		r.Route("/movie", func(r chi.Router) {
			r.Use(app.requireAuthenticatedUser)
			r.Get("/", app.listMoviesHandler)
			r.With(app.requireActivatedUser).Post("/", app.postMovieHandler)
			r.Post("/predict", app.predictHandler)

			r.Route("/{movieID}", func(r chi.Router) {
				r.Get("/", app.getMovieHandler)
				r.With(app.requireActivatedUser).Patch("/", app.updateMovieHandler)
				r.Delete("/", app.deleteMovieHandler)
			})
		})

		r.Route("/users", func(r chi.Router) {
			r.Post("/", app.registerUserHandler)
			r.Put("/activate", app.activateUserHandler)
			r.Put("/password", app.updateUserPasswordHandler)
		})

		r.Route("/tokens", func(r chi.Router) {
			r.Post("/authentication", app.createAuthenticationTokenHandler)
			r.Post("/refresh", app.refreshTokenHandler)
			r.Post("/password-reset", app.createPasswordResetCodeHandler)
		})
	})

	r.Method(http.MethodGet, "/metrics", promhttp.Handler())

	return r
}
