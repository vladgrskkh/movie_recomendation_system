package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/vladgrskkh/movie_recomendation_system/internal/data"
	"github.com/vladgrskkh/movie_recomendation_system/internal/data/mocks"
)

func TestHealthCheckHandler(t *testing.T) {
	app := newTestApplication(t)

	app.config.env = "development"

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

func TestPostMovieHandler(t *testing.T) {
	app := newTestApplication(t)

	ts := newTestServer(t, testRoutes(app))
	defer ts.Close()

	mockMovies := mocks.NewMoviesInterface(t)

	movieReq := movieInput{
		Title:   "Test Movie",
		Year:    2024,
		Runtime: 125,
		Genres:  []string{"Drama", "Action"},
	}

	movie := data.Movie{
		Title:   movieReq.Title,
		Year:    movieReq.Year,
		Runtime: movieReq.Runtime,
		Genres:  movieReq.Genres,
	}

	mockMovies.On("Insert", &movie).Return(nil).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*data.Movie)
		arg.ID = 1
		arg.Version = 1
	})

	app.models.Movies = mockMovies

	tests := []struct {
		name     string
		reqBody  interface{}
		wantCode int
		wantBody *data.Movie
	}{
		{
			name:     "Valid Movie",
			reqBody:  movieReq,
			wantCode: http.StatusCreated,
			wantBody: &data.Movie{
				ID:      1,
				Title:   movieReq.Title,
				Year:    movieReq.Year,
				Runtime: movieReq.Runtime,
				Genres:  movieReq.Genres,
				Version: 1,
			},
		},
		{
			name: "Invalid Movie (missing title)",
			reqBody: movieInput{
				Year:    2024,
				Runtime: 125,
				Genres:  []string{"Drama", "Action"},
			},
			wantCode: http.StatusUnprocessableEntity,
		},
		{
			name: "Invalid Movie (year out of range)",
			reqBody: movieInput{
				Title:   "Test Movie",
				Year:    3,
				Runtime: 125,
				Genres:  []string{"Drama", "Action"},
			},
			wantCode: http.StatusUnprocessableEntity,
		},
		{
			name: "Invalid Movie (negative runtime)",
			reqBody: movieInput{
				Title:   "Test Movie",
				Year:    2024,
				Runtime: -125,
				Genres:  []string{"Drama", "Action"},
			},
			wantCode: http.StatusUnprocessableEntity,
		},
		{
			name: "Invalid Movie (empty genres)",
			reqBody: movieInput{
				Title:   "Test Movie",
				Year:    2024,
				Runtime: 125,
				Genres:  []string{},
			},
			wantCode: http.StatusUnprocessableEntity,
		},
		{
			name: "Invalid Movie (too many genres)",
			reqBody: movieInput{
				Title:   "Test Movie",
				Year:    2024,
				Runtime: 125,
				Genres:  []string{"Drama", "Action", "Comedy", "Horror", "Sci-Fi", "Romance"},
			},
			wantCode: http.StatusUnprocessableEntity,
		},
		{
			name: "Invalid Movie (duplicate genres)",
			reqBody: movieInput{
				Title:   "Test Movie",
				Year:    2024,
				Runtime: 125,
				Genres:  []string{"Drama", "Action", "Drama"},
			},
			wantCode: http.StatusUnprocessableEntity,
		},
		{
			name:     "Invalid JSON",
			reqBody:  map[string]string{"invalid_json": "invalid_json"},
			wantCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestBody, err := json.MarshalIndent(&tt.reqBody, "", "\t")
			assert.NoError(t, err)

			buffer := bytes.NewBuffer(requestBody)
			code, _, body := ts.post(t, "/v1/movie", buffer)

			assert.Equal(t, tt.wantCode, code, fmt.Sprintf("status code should be %d", tt.wantCode))
			if tt.wantBody != nil {
				var movieResp map[string]data.Movie

				err = json.Unmarshal(body, &movieResp)
				assert.NoError(t, err)

				assert.Equal(t, int64(1), movieResp["movie"].ID, "movie ID should be 1")
				assert.Equal(t, int32(1), movieResp["movie"].Version, "movie version should be 1")
				assert.Equal(t, movie.Title, movieResp["movie"].Title, "movie title should be equal")
				assert.Equal(t, movie.Year, movieResp["movie"].Year, "movie year should be equal")
				assert.Equal(t, movie.Runtime, movieResp["movie"].Runtime, "movie runtime should be equal")
				assert.Equal(t, movie.Genres, movieResp["movie"].Genres, "movie genres should be equal")
			}
		})
	}
}

func TestDeleteMovieHandler(t *testing.T) {
	app := newTestApplication(t)

	ts := newTestServer(t, testRoutes(app))
	defer ts.Close()

	mockMovies := mocks.NewMoviesInterface(t)

	mockMovies.On("Delete", int64(1)).Return(nil)
	mockMovies.On("Delete", int64(2)).Return(data.ErrRecordNotFound)

	app.models.Movies = mockMovies

	tests := []struct {
		name     string
		urlPath  string
		wantCode int
	}{
		{
			name:     "Valid ID",
			urlPath:  "/v1/movie/1",
			wantCode: http.StatusOK,
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
			code, _, _ := ts.delete(t, tt.urlPath)

			assert.Equal(t, tt.wantCode, code, fmt.Sprintf("status code should be %d", tt.wantCode))
		})
	}
}
