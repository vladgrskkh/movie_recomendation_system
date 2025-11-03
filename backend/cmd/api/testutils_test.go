package main

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"
)

func newTestApplication(t *testing.T) *application {
	return &application{
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
}

type testServer struct {
	*httptest.Server
}

func newTestServer(t *testing.T, h http.Handler) *testServer {
	ts := httptest.NewServer(h)
	return &testServer{ts}
}

func (ts *testServer) get(t *testing.T, urlPath string) (int, http.Header, []byte) {
	rs, err := ts.Client().Get(ts.URL + urlPath)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		e := rs.Body.Close()
		if err != nil {
			err = fmt.Errorf("previous error: %w; close error: %w", err, e)
		} else if e != nil {
			t.Fatal(e)
		}
	}()

	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}

	return rs.StatusCode, rs.Header, body
}

func testRoutes(app *application) http.Handler {
	r := chi.NewRouter()

	r.Route("/v1", func(r chi.Router) {
		r.Get("/healthcheck", app.healthCheckHandler)
		r.Get("/swagger/*", httpSwagger.Handler())

		r.Route("/movie", func(r chi.Router) {
			r.Get("/", app.listMoviesHandler)
			r.Post("/", app.postMovieHandler)
			r.Post("/predict", app.predictHandler)

			r.Route("/{movieID}", func(r chi.Router) {
				r.Get("/", app.getMovieHandler)
				r.Patch("/", app.updateMovieHandler)
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
			r.Post("/activation", app.createActivationTokenHandler)
		})
	})

	r.Method(http.MethodGet, "/metrics", promhttp.Handler())

	return r
}
