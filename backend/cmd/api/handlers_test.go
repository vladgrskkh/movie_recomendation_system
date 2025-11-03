package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vladgrskkh/movie_recomendation_system/internal/data"
	"github.com/vladgrskkh/movie_recomendation_system/internal/data/mocks"
)

func TestHealthCheckHandler(t *testing.T) {
	app := newTestApplication(t)

	cfg := config{
		env: "development",
	}

	app.config = cfg

	ts := newTestServer(t, testRoutes(app))
	defer ts.Close()

	code, _, body := ts.get(t, "/v1/healthcheck")

	var data struct {
		Status  string `json:"status"`
		Env     string `json:"env"`
		Version string `json:"version"`
	}

	err := json.Unmarshal(body, &data)
	assert.NoError(t, err)

	assert.Equal(t, 200, code, "status code should be 200")
	assert.Equal(t, data.Status, "avaliable", "status should be 'available'")
	assert.Contains(t, []string{"development", "staging", "production"}, data.Env, "env should be valid")
}

func TestGetMovieHandler(t *testing.T) {
	app := newTestApplication(t)

	ts := newTestServer(t, testRoutes(app))
	defer ts.Close()

	mockMovies := mocks.NewMoviesInterface(t)

	movie := data.Movie{
		ID:      1,
		Title:   "Test Movie",
		Year:    2024,
		Runtime: 125,
		Genres:  []string{"Drama", "Action"},
		Version: 1,
	}

	mockMovies.On("Get", int64(1)).Return(&movie, nil)
	mockMovies.On("Get", int64(2)).Return(nil, data.ErrRecordNotFound)

	app.models.Movies = mockMovies

	tests := []struct {
		name     string
		urlPath  string
		wantCode int
		wantBody *data.Movie
	}{
		{
			name:     "Valid ID",
			urlPath:  "/v1/movie/1",
			wantCode: http.StatusOK,
			wantBody: &movie,
		},
		{
			name:     "Non-existent ID",
			urlPath:  "/v1/movie/2",
			wantCode: http.StatusNotFound,
		},
		{
			name:     "Negative ID",
			urlPath:  "/v1/movie/-1",
			wantCode: http.StatusNotFound,
		},
		{
			name:     "Decimal ID",
			urlPath:  "/v1/movie/1.23",
			wantCode: http.StatusNotFound,
		},
		{
			name:     "String ID",
			urlPath:  "/v1/movie/smth",
			wantCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, _, body := ts.get(t, tt.urlPath)

			var data map[string]data.Movie

			assert.Equal(t, tt.wantCode, code, fmt.Sprintf("status code should be %d", tt.wantCode))
			if tt.wantBody != nil {
				err := json.Unmarshal(body, &data)
				assert.NoError(t, err)

				assert.Equal(t, *tt.wantBody, data["movie"], "movie should be equal")
			}
		})
	}
}
