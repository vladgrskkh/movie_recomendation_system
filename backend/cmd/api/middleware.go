package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/vladgrskkh/movie_recomendation_system/internal/data"
)

func (app *application) authentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Authorization")

		authorizationHeader := r.Header.Get("Authorization")

		if authorizationHeader == "" {
			r = app.contextSetUser(r, data.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}

		headerParts := strings.Split(authorizationHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			app.invalidAuthenticationResponse(w, r)
			return
		}

		token := headerParts[1]

		claims, err := validateToken(token)
		if err != nil {
			switch {
			case errors.Is(err, ErrInvalidToken):
				app.failedValidationResponse(w, r, err)
			default:
				app.serverErrorResponse(w, r, err)
			}

			return
		}

		user := data.User{
			ID:        claims.UserID,
			Activated: claims.Activated,
		}

		r = app.contextSetUser(r, &user)

		next.ServeHTTP(w, r)
	})
}

func (app *application) requireAuthenticatedUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := app.contextGetUser(r)

		if user.IsAnonymous() {
			app.authenticationRequiredResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (app *application) requireActivatedUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := app.contextGetUser(r)

		if !user.Activated {
			app.inactiveAccountResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "Close")

				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (app *application) metrics(next http.Handler) http.Handler {
	// TODO: think how to add avg stats to prometheus
	totalRequestsReceived := promauto.NewCounter(prometheus.CounterOpts{
		Name: "total_requests_received",
		Help: "The total number of requests received",
	})
	totalResponsesSent := promauto.NewCounter(prometheus.CounterOpts{
		Name: "total_responses_sent",
		Help: "The total number of responses sent",
	})
	totalProcessingTimeMicroseconds := promauto.NewCounter(prometheus.CounterOpts{
		Name: "total_processing_time_microseconds",
		Help: "The total (cumulative) time taken to process all requests in microseconds",
	})
	activeRequests := promauto.NewGauge(prometheus.GaugeOpts{
		Name: "in_flight_requests",
		Help: "The number of 'active' in-flight requests",
	})
	// avgRequestReceivedPerRequest := promauto.NewGauge(prometheus.GaugeOpts{
	// Name: "avg_request_received_per_request",
	// Help: "The average amount of requests received per second (between scrapes)",
	// })
	// avgProcessingTimePerRequest := promauto.NewGauge(prometheus.GaugeOpts{
	// Name: "avg_processing_time_per_request",
	// Help: "The average processing time per request in microseconds (between scrapes)",
	// })

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		totalRequestsReceived.Inc()
		activeRequests.Inc()

		next.ServeHTTP(w, r)

		totalResponsesSent.Inc()
		activeRequests.Dec()

		duration := time.Since(start).Microseconds()
		totalProcessingTimeMicroseconds.Add(float64(duration))
	})
}
