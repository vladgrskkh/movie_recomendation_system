package main

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"
)

func newTestApplication(t *testing.T) *application {
	return &application{
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		config: config{
			jwt: struct {
				secretKey      string
				secretKeyBytes []byte
			}{
				secretKeyBytes: []byte("my_secret_key"),
			},
		},
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
	req, err := http.NewRequest(http.MethodGet, ts.URL+urlPath, nil)
	if err != nil {
		t.Fatal(err)
	}

	token, err := testAuth(1, true, newTestApplication(t))
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	rs, err := ts.Client().Do(req)
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

func (ts *testServer) post(t *testing.T, urlPath string, requestBody io.Reader) (int, http.Header, []byte) {
	req, err := http.NewRequest(http.MethodPost, ts.URL+urlPath, requestBody)
	if err != nil {
		t.Fatal(err)
	}

	token, err := testAuth(1, true, newTestApplication(t))
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	rs, err := ts.Client().Do(req)
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

func (ts *testServer) delete(t *testing.T, urlPath string) (int, http.Header, []byte) {
	req, err := http.NewRequest(http.MethodDelete, ts.URL+urlPath, nil)
	if err != nil {
		t.Fatal(err)
	}

	token, err := testAuth(1, true, newTestApplication(t))
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	rs, err := ts.Client().Do(req)
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

func testAuth(userID int64, activation bool, app *application) (string, error) {
	token, err := createToken(userID, activation, app)
	if err != nil {
		return "", err
	}

	return token, nil
}

func testRoutes(app *application) http.Handler {
	r := chi.NewRouter()

	r.Use(app.recoverPanic)
	r.Use(app.authentication)

	// Rate-limit all routes
	// Think about adding rate limmiter for specific routes(registaration, login)
	if app.config.limiter.enable {
		r.Use(httprate.LimitByIP(app.config.limiter.rps, time.Second))
	}

	r.Route("/v1", func(r chi.Router) {
		r.Get("/healthcheck", app.healthCheckHandler)

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
			r.Post("/activation", app.createActivationTokenHandler)
		})
	})

	return r
}
