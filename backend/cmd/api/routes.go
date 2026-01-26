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

		r.Route("/movies", func(r chi.Router) {
			r.Use(app.requireAuthenticatedUser)
			r.Get("/", app.listMoviesHandler)
			r.With(app.requireActivatedUser).Post("/", app.postMovieHandler)
			r.Post("/predict", app.predictHandler)
			r.Get("/new", app.getMoviesNew)
			r.Get("/recommended", app.getMoviesRecommended)
			r.Get("/popular", app.getMoviesPopular)

			r.Route("/{movieID}", func(r chi.Router) {
				r.Get("/", app.getMovieHandler)
				r.With(app.requireActivatedUser).Patch("/", app.updateMovieHandler)
				r.Delete("/", app.deleteMovieHandler)
				r.Put("/like", app.likeMovieHandler)
				r.Put("/unlike", app.unlikeMovieHandler)
			})
		})

		r.Route("/users", func(r chi.Router) {
			r.Post("/", app.registerUserHandler)
			r.Put("/activate", app.activateUserHandler)
			r.Put("/password", app.updateUserPasswordHandler)
			r.With(app.requireAuthenticatedUser).Delete("/", app.deleteUserHandler)
		})

		r.Route("/tokens", func(r chi.Router) {
			r.Post("/authentication", app.createAuthenticationTokenHandler)
			r.Post("/refresh", app.refreshTokenHandler)
			r.Post("/password-reset", app.createPasswordResetCodeHandler)
			r.Post("/activation", app.createActivationTokenHandler)
		})

		r.Route("/kafka", func(r chi.Router) {
			r.Post("/dummy", app.createKafkaMessage)
		})

		r.Route("/images", func(r chi.Router) {
			r.Use(app.requireAuthenticatedUser)
			r.Post("/", app.UploadImageHandler)
			r.Get("/{imageID}", app.GetImageHandler)
			r.Delete("/{imageID}", app.DeleteImageHandler)
		})
	})

	r.Method(http.MethodGet, "/metrics", promhttp.Handler())

	return r
}
