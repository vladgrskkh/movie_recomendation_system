package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/invopop/validation"
	"github.com/invopop/validation/is"

	pb "github.com/vladgrskkh/movie_recomendation_system/genproto/v1/predict"

	"github.com/vladgrskkh/movie_recomendation_system/internal/data"
	"github.com/vladgrskkh/movie_recomendation_system/internal/validate"
)

// HealthCheck godoc
//
// @Summary Health check
// @Description Returns service health and metadata
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string "OK | Example {"status": "avaliable", "env": "production", "version": "1.0.0"}"
// @Failure 500 {object} map[string]string "Internal Server Error | Example {"error": "server encountered a problem and could not process your request"}"
// @Router /healthcheck [get]
func (app *application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	data := envelope{
		"status":  "avaliable",
		"env":     app.config.env,
		"version": version,
	}

	err := app.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// GetMovie godoc
//
// @Summary Get a movie by ID
// @Description Returns a single movie by numeric ID
// @Tags movies
// @Produce json
// @Param movieID path int true "Movie ID"
// @Success 200 {object} data.Movie
// @Failure 401 {object} map[string]string "Unauthorized | Example {"error": "this resourse avaliable only for authenticated users"}"
// @Failure 404 {object} map[string]string "Not Found | Example {"error": "requested resource could not be found"}"
// @Failure 500 {object} map[string]string "Internal Server Error | Example {"error": "server encountered a problem and could not process your request"}"
// @Security BearerAuth
// @Router /movie/{movieID} [get]
func (app *application) getMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	movie, err := app.models.Movies.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}

		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

type movieInput struct {
	Title   string   `json:"title" example:"The Shawshank Redemption"`
	Year    int32    `json:"year" example:"1994"`
	Runtime int32    `json:"runtime" example:"142"`
	Genres  []string `json:"genres" example:"Drama,Crime"`
}

// CreateMovie godoc
//
// @Summary Create a movie
// @Description Create a new movie (admin only)
// @Tags movies
// @Accept json
// @Produce json
// @Param movie body movieInput true "Movie payload"
// @Success 201 {object} data.Movie
// @Failure 400 {object} map[string]string "Bad Request | Example {"error": "body contains badly-formated JSON"}"
// @Failure 403 {object} map[string]string "Forbidden | Example {"error": "your account must be activated to access this resourse"}"
// @Failure 422 {object} map[string]string "Unprocessable Entity | Example {"error": "validation error"}"
// @Failure 500 {object} map[string]string "Internal Server Error | Example {"error": "server encountered a problem and could not process your request"}"
// @Security BearerAuth
// @Router /movie [post]
func (app *application) postMovieHandler(w http.ResponseWriter, r *http.Request) {
	var input movieInput

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	movie := &data.Movie{
		Title:   input.Title,
		Year:    input.Year,
		Runtime: input.Runtime,
		Genres:  input.Genres,
	}

	// Validation of the movie input
	err = validation.ValidateStruct(movie,
		validation.Field(&movie.Title, validation.Required, validation.Length(1, 500)),
		validation.Field(&movie.Year, validation.Required, validation.Min(1888), validation.Max(int32(time.Now().Year()))),
		validation.Field(&movie.Runtime, validation.Required, validation.Min(1)),
		validation.Field(&movie.Genres, validation.Required, validation.Length(1, 5), validation.By(validate.Unique(movie.Genres))),
	)
	if err != nil {
		app.failedValidationResponse(w, r, err)
		return
	}

	err = app.models.Movies.Insert(movie)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/movie/%d", movie.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"movie": movie}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

// admin only
// DeleteMovie godoc
//
// @Summary Delete a movie
// @Description Delete movie by ID (admin only)
// @Tags movies
// @Produce json
// @Param movieID path int true "Movie ID"
// @Success 200 {object} map[string]string "OK | Example {"message": "movie successfully deleted"}"
// @Failure 404 {object} map[string]string "Not Found | Example {"error": "requested resource could not be found"}"
// @Failure 500 {object} map[string]string "Internal Server Error | Example {"error": "server encountered a problem and could not process your request"}"
// @Security BearerAuth
// @Router /movie/{movieID} [delete]
func (app *application) deleteMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.Movies.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
			return
		default:
			app.serverErrorResponse(w, r, err)
			return
		}
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "movie successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// UpdateMovie godoc
//
// @Summary Update a movie
// @Description Patch movie by ID
// @Tags movies
// @Accept json
// @Produce json
// @Param movieID path int true "Movie ID"
// @Param movie body movieInput true "Partial movie payload"
// @Success 200 {object} data.Movie
// @Failure 400 {object} map[string]string "Bad Request | Example {"error": "body contains badly-formated JSON"}"
// @Failure 404 {object} map[string]string "Not Found | Example {"error": "requested resource could not be found"}"
// @Failure 409 {object} map[string]string "Conflict | Example {"error": "unable to update the record due to an edit conflict, please try again"}"
// @Failure 422 {object} map[string]string "Unprocessable Entity | Example {"error": "validation error"}"
// @Failure 500 {object} map[string]string "Internal Server Error | Example {"error": "server encountered a problem and could not process your request"}"
// @Security BearerAuth
// @Router /movie/{movieID} [patch]
func (app *application) updateMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	movie, err := app.models.Movies.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}

		return
	}

	// using pointers here to be able to compare which field was empty
	var input struct {
		Title   *string  `json:"title"`
		Year    *int32   `json:"year"`
		Runtime *int32   `json:"runtime"`
		Genres  []string `json:"genres"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.Title != nil {
		movie.Title = *input.Title
	}

	if input.Year != nil {
		movie.Year = *input.Year
	}

	if input.Runtime != nil {
		movie.Runtime = *input.Runtime
	}

	if input.Genres != nil {
		movie.Genres = input.Genres
	}

	err = validation.ValidateStruct(movie,
		validation.Field(&movie.Title, validation.Required, validation.Length(1, 500)),
		validation.Field(&movie.Year, validation.Required, validation.Min(1888), validation.Max(int32(time.Now().Year()))),
		validation.Field(&movie.Runtime, validation.Required, validation.Min(1)),
		validation.Field(&movie.Genres, validation.Required, validation.Length(1, 5), validation.By(validate.Unique(movie.Genres))),
	)
	if err != nil {
		app.failedValidationResponse(w, r, err)
		return
	}

	err = app.models.Movies.Update(movie)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

type MoviesListResponse struct {
	Movies   []data.Movie  `json:"movies"`
	Metadata data.Metadata `json:"metadata"`
}

// ListMovies godoc
//
// @Summary List movies
// @Description Retrieve a list of movies filtered by title and genres with pagination and sorting
// @Tags movies
// @Produce json
// @Param title query string false "Full-text search by title"
// @Param genres query []string false "Comma-separated list of genres" collectionFormat(csv)
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(20)
// @Param sort query string false "Sort by: one of id,title,year,runtime,-id,-title,-year,-runtime" default(id)
// @Security BearerAuth
// @Success 200 {object} MoviesListResponse
// @Failure 401 {object} map[string]string "Unauthorized | Example {"error": "this resourse avaliable only for authenticated users"}"
// @Failure 422 {object} map[string]string "Unprocessable Entity | Example {"error": "validation error"}"
// @Failure 500 {object} map[string]string "Internal Server Error | Example {"error": "server encountered a problem and could not process your request"}"
// @Router /movie [get]
func (app *application) listMoviesHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title  string
		Genres []string
	}

	var filters data.Filters

	qs := r.URL.Query()

	var err error

	input.Title = app.readString(qs, "title", "")
	input.Genres = app.readCSV(qs, "genres", []string{})

	filters.Page, err = app.readInt(qs, "page", 1)
	if err != nil {
		app.failedValidationResponse(w, r, err)
		return
	}

	filters.PageSize, err = app.readInt(qs, "page_size", 20)
	if err != nil {
		app.failedValidationResponse(w, r, err)
		return
	}

	filters.Sort = app.readString(qs, "sort", "id")
	filters.SortSafeList = []string{"id", "title", "year", "runtime", "-id", "-title", "-year", "-runtime"}

	err = validation.ValidateStruct(&filters,
		validation.Field(&filters.Page, validation.Required, validation.Min(1), validation.Max(10_000_000)),
		validation.Field(&filters.PageSize, validation.Required, validation.Min(1)),
		validation.Field(&filters.Sort, validation.Required, validation.In(filters.SortSafeList...)),
	)

	if err != nil {
		app.failedValidationResponse(w, r, err)
		return
	}

	movies, metadata, err := app.models.Movies.GetAll(input.Title, input.Genres, filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"movies": movies, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

type registerInput struct {
	Name     string `json:"name" example:"John Doe"`
	Email    string `json:"email" example:"something@example.com"`
	Password string `json:"password" example:"s1mplepA$$word"`
}

// kafkaMessage struct hold info about user activation/reset password
// need to work on naming
type kafkaMessage struct {
	UserID       int64  `json:"user_id"`
	Email        string `json:"email"`
	Name         string `json:"name"`
	Token        string `json:"token"`
	TemplateName string `json:"template_name,omitempty"`
	Task         string `json:"task"`
}

// RegisterUser godoc
//
// @Summary Register a new user
// @Description Creates a user account and sends activation email
// @Tags users
// @Accept json
// @Produce json
// @Param user body registerInput true "Registration payload"
// @Success 202 {object} data.User
// @Failure 400 {object} map[string]string "Bad Request | Example {"error": "body contains badly-formated JSON"}"
// @Failure 422 {object} map[string]string "Unprocessable Entity | Example {"error": "validation error"}"
// @Failure 500 {object} map[string]string "Internal Server Error | Example {"error": "server encountered a problem and could not process your request"}"
// @Router /users [post]
func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var input registerInput

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := &data.User{
		Name:      input.Name,
		Email:     input.Email,
		Activated: false,
	}

	err = user.Password.Set(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = validation.ValidateStruct(&input,
		validation.Field(&input.Name, validation.Required, validation.Length(1, 500)),
		validation.Field(&input.Email, validation.Required, is.Email),
		validation.Field(&input.Password, validation.Required, validation.Length(8, 72)))

	if err != nil {
		app.failedValidationResponse(w, r, err)
		return
	}

	err = app.models.Users.Insert(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			app.failedValidationResponse(w, r, err)
		default:
			app.serverErrorResponse(w, r, err)
		}

		return
	}

	token, err := app.models.Tokens.New(user.ID, 5*time.Minute, data.ScopeActivation)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// SMTP

	// kafka producer handling
	message := &kafkaMessage{
		UserID:       user.ID,
		Email:        user.Email,
		Name:         user.Name,
		Token:        token.Plaintext,
		TemplateName: "user_welcome.html",
		Task:         "send welcome email with activation token to user's email",
	}

	// change key later (need to test how it works)
	err = app.producer.Produce(message, app.config.kafka.topic, nil, time.Now())
	if err != nil {
		// log or return if cannot produce msg (either bad json format or some problem with brokers)
		app.logger.Error(err.Error())
	}

	data := envelope{
		"activationToken": token.Plaintext,
		"name":            user.Name,
	}

	app.background(func() {
		err = app.mailer.Send(user.Email, "user_welcome.html", data)
		if err != nil {
			app.logger.Error(err.Error())
		}
	})

	err = app.writeJSON(w, http.StatusAccepted, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

type activateInput struct {
	Token string `json:"token"`
}

// ActivateUser godoc
//
// @Summary Activate user
// @Description Activates a user account using activation token
// @Tags users
// @Accept json
// @Produce json
// @Param activation body activateInput true "Activation payload"
// @Success 200 {object} activateInput
// @Failure 400 {object} map[string]string "Bad Request | Example {"error": "body contains badly-formated JSON"}"
// @Failure 409 {object} map[string]string "Conflict | Example {"error": "unable to update the record due to an edit conflict, please try again"}"
// @Failure 422 {object} map[string]string "Unprocessable Entity | Example {"error": "validation error"}"
// @Failure 500 {object} map[string]string "Internal Server Error | Example {"error": "server encountered a problem and could not process your request"}"
// @Router /users/activate [put]
func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	var input activateInput

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	err = validation.Validate(input.Token, validation.Required, validation.Length(5, 5))
	if err != nil {
		app.failedValidationResponse(w, r, err)
		return
	}

	user, err := app.models.Users.GetForToken(data.ScopeActivation, input.Token)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.invalidActicationTokenResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}

		return
	}

	user.Activated = true

	err = app.models.Users.Update(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}

		return
	}

	err = app.models.Tokens.DeleteAllForUser(data.ScopeActivation, user.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// auth token with updated payload (user.Actavated field)
	authToken, err := createToken(user.ID, user.Activated, app)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"token": authToken}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

type refreshInput struct {
	RefreshToken string `json:"refresh_token"`
}

type tokenPair struct {
	AuthenticationToken string `json:"authentication_token"`
	RefreshToken        string `json:"refresh_token"`
}

// RefreshToken godoc
//
// refreshTokenHandler wants refresh token to create auth token and new refresh token
// auth token is jwt and refresh token is high entropy string
//
// @Summary Refresh tokens
// @Description Exchange refresh token for new auth and refresh tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param token body refreshInput true "Refresh token payload"
// @Success 201 {object} tokenPair
// @Failure 400 {object} map[string]string "Bad Request | Example {"error": "body contains badly-formated JSON"}"
// @Failure 422 {object} map[string]string "Unprocessable Entity | Example {"error": "validation error"}"
// @Failure 500 {object} map[string]string "Internal Server Error | Example {"error": "server encountered a problem and could not process your request"}"
// @Router /tokens/refresh [post]
func (app *application) refreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	var input refreshInput

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user, err := app.models.Users.GetForToken(data.ScopeRefresh, input.RefreshToken)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.invalidRefreshTokenResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}

		return
	}

	// maybe create helper function for all the code down there
	err = app.models.Tokens.DeleteAllForUser(data.ScopeRefresh, user.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// refresh token
	refreshToken, err := app.models.Tokens.New(user.ID, 30*24*time.Hour, data.ScopeRefresh)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// auth token
	authToken, err := createToken(user.ID, user.Activated, app)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	tokenPair := tokenPair{
		AuthenticationToken: authToken,
		RefreshToken:        refreshToken.Plaintext,
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"token_pair": tokenPair}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

type loginInput struct {
	Email    string `json:"email" example:"something@example.com"`
	Password string `json:"password" example:"s1mplepA$$word"`
}

// createAuthenticationTokenHandler godoc
//
// createAuthenticationTokenHandler is log in for app
// every time user log in we will create new auth token and refresh token(deleting prev refresh token if exists)
//
// @Summary Log in and get tokens
// @Description Creates authentication and refresh tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body loginInput true "Login payload"
// @Success 201 {object} tokenPair
// @Failure 400 {object} map[string]string "Bad Request | Example {"error": "body contains badly-formated JSON"}"
// @Failure 401 {object} map[string]string "Unauthorized | Example {"error": "invalid authentication credentials"}
// @Failure 422 {object} map[string]string "Unprocessable Entity | Example {"error": "validation error"}"
// @Failure 500 {object} map[string]string "Internal Server Error | Example {"error": "server encountered a problem and could not process your request"}"
// @Router /tokens/authentication [post]
func (app *application) createAuthenticationTokenHandler(w http.ResponseWriter, r *http.Request) {
	var input loginInput

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	err = validation.ValidateStruct(&input,
		validation.Field(&input.Email, validation.Required, is.Email),
		validation.Field(&input.Password, validation.Required, validation.Length(8, 72)))
	if err != nil {
		app.failedValidationResponse(w, r, err)
		return
	}

	user, err := app.models.Users.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.invalidCredentialResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}

		return
	}

	match, err := user.Password.Matches(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

	if !match {
		app.invalidCredentialResponse(w, r)
		return
	}

	// maybe create helper function for all the code down there
	err = app.models.Tokens.DeleteAllForUser(data.ScopeRefresh, user.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// refresh token
	refreshToken, err := app.models.Tokens.New(user.ID, 30*24*time.Hour, data.ScopeRefresh)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// auth token
	authToken, err := createToken(user.ID, user.Activated, app)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	tokenPair := tokenPair{
		AuthenticationToken: authToken,
		RefreshToken:        refreshToken.Plaintext,
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"token_pair": tokenPair}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

type predictionInput struct {
	Title string `json:"title" example:"The Shawshank Redemption"`
}

// Predict Handler godoc
//
// @Summary Get predict movie
// @Description Validates movie input and predict movie
// @Tags movies
// @Accept json
// @Produce json
// @Param credentials body predictionInput true "Moive payload"
// @Security BearerAuth
// @Success 200 {object} predictionInput
// @Failure 400 {object} map[string]string "Bad Request | Example {"error": "body contains badly-formated JSON"}"
// @Failure 401 {object} map[string]string "Unauthorized | Example {"error": "this resourse avaliable only for authenticated users"}"
// @Failure 422 {object} map[string]string "Unprocessable Entity | Example {"error": "validation error"}"
// @Failure 500 {object} map[string]string "Internal Server Error | Example {"error": "server encountered a problem and could not process your request"}"
// @Router /movie/predict [post]
func (app *application) predictHandler(w http.ResponseWriter, r *http.Request) {
	var input predictionInput

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	err = validation.ValidateStruct(&input,
		validation.Field(&input.Title, validation.Required, validation.Length(1, 500)),
	)
	if err != nil {
		app.failedValidationResponse(w, r, err)
		return
	}

	client := pb.NewRecommendationClient(app.grpcConn)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	recommendation, err := client.Recommend(ctx, &pb.RecommendRequest{
		MovieTitle: input.Title,
	})
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	data := envelope{"recommendations": recommendation.GetRecommendations()}

	err = app.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

type inputChangePassword struct {
	Email string `json:"email" example:"something@example.com"`
}

// createPasswordResetCodeHandler godoc
//
// @Summary Post create password reset
// @Description Validates email and checks if user exists and activated than sends email with code
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body inputChangePassword true "Change password payload"
// @Success 202 {object} map[string]string "Accepted | Exmaple {"message": "check your email for reset code"}"
// @Failure 400 {object} map[string]string "Bad Request | Example {"error": "body contains badly-formated JSON"}"
// @Failure 422 {object} map[string]string "Unprocessable Entity | Example {"error": "validation error"}"
// @Failure 500 {object} map[string]string "Internal Server Error | Example {"error": "server encountered a problem and could not process your request"}"
// @Router /tokens/password-reset [post]
func (app *application) createPasswordResetCodeHandler(w http.ResponseWriter, r *http.Request) {
	var input inputChangePassword

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	err = validation.ValidateStruct(&input,
		validation.Field(&input.Email, validation.Required, is.Email),
	)
	if err != nil {
		app.failedValidationResponse(w, r, err)
		return
	}

	user, err := app.models.Users.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.invalidCredentialResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}

		return
	}

	if !user.Activated {
		app.failedValidationResponse(w, r, errors.New("user account must be activated"))
		return
	}

	resetCode, err := app.models.Tokens.New(user.ID, 2*time.Minute, data.ScopePasswordReset)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// kafka producer handling
	message := &kafkaMessage{
		UserID:       user.ID,
		Email:        user.Email,
		Name:         user.Name,
		Token:        resetCode.Plaintext,
		TemplateName: "user_reset_password.html",
		Task:         "send reset password token to user email",
	}

	// change key later (need to test how it works)
	err = app.producer.Produce(message, app.config.kafka.topic, nil, time.Now())
	if err != nil {
		// log or return if cannot produce msg (either bad json format or some problem with brokers)
		app.logger.Error(err.Error())
	}

	data := envelope{
		"resetCode": resetCode.Plaintext,
		"name":      user.Name,
	}

	app.background(func() {
		err = app.mailer.Send(user.Email, "user_reset_password.html", data)
		if err != nil {
			app.logger.Error(err.Error())
		}
	})

	err = app.writeJSON(w, http.StatusAccepted, envelope{"message": "check your email for reset code"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

type inputUpdatePassword struct {
	Code        string `json:"code" example:"123454"`
	NewPassword string `json:"new_password" example:"n3wP@ssw0rd!"`
}

// updateUserPasswordHandler godoc
//
// @Summary Put update user password
// @Description Validates new password and code, sets new password for user
// @Tags users
// @Accept json
// @Produce json
// @Param credentials body inputUpdatePassword true "Update password payload"
// @Success 200 {object} map[string]string "OK | Exmaple {"message": "your password was successfully reset"}"
// @Failure 400 {object} map[string]string "Bad Request | Example {"error": "body contains badly-formated JSON"}"
// @Failure 409 {object} map[string]string "Conflict | Example {"error": "unable to update the record due to an edit conflict, please try again"}"
// @Failure 422 {object} map[string]string "Unprocessable Entity | Example {"error": "validation error"}"
// @Failure 500 {object} map[string]string "Internal Server Error | Example {"error": "server encountered a problem and could not process your request"}"
// @Router /users/password [put]
func (app *application) updateUserPasswordHandler(w http.ResponseWriter, r *http.Request) {
	var input inputUpdatePassword

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	err = validation.ValidateStruct(&input,
		validation.Field(&input.NewPassword, validation.Required, validation.Length(8, 72)),
		validation.Field(&input.Code, validation.Required, validation.Length(5, 5)),
	)
	if err != nil {
		app.failedValidationResponse(w, r, err)
		return
	}

	user, err := app.models.Users.GetForToken(data.ScopePasswordReset, input.Code)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.failedValidationResponse(w, r, errors.New("invalid or expired password code"))
		default:
			app.serverErrorResponse(w, r, err)
		}

		return
	}

	err = user.Password.Set(input.NewPassword)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.models.Users.Update(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}

		return
	}

	err = app.models.Tokens.DeleteAllForUser(data.ScopePasswordReset, user.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	msg := envelope{"message": "your password was successfully reset"}

	err = app.writeJSON(w, http.StatusOK, msg, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// createActivationTokenHandler godoc
//
// @Summary Create new activation token
// @Description Validates email and checks if user exists, if not activated we send mail with activation code
// @Tags tokens
// @Accept json
// @Produce json
// @Param credentials body inputChangePassword true "Create activation token payload"
// @Success 202 {object} map[string]string "Accepted | Exmaple {"message": "check your email for activation code"}"
// @Failure 400 {object} map[string]string "Bad Request | Example {"error": "body contains badly-formated JSON"}"
// @Failure 422 {object} map[string]string "Unprocessable Entity | Example {"error": "validation error"}"
// @Failure 500 {object} map[string]string "Internal Server Error | Example {"error": "server encountered a problem and could not process your request"}"
// @Router /tokens/activation [post]
func (app *application) createActivationTokenHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email string `json:"email" example:"something@example.com"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	err = validation.ValidateStruct(&input,
		validation.Field(&input.Email, validation.Required, is.Email),
	)
	if err != nil {
		app.failedValidationResponse(w, r, err)
		return
	}

	user, err := app.models.Users.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.failedValidationResponse(w, r, errors.New("no matching email address found"))
		default:
			app.serverErrorResponse(w, r, err)
		}

		return
	}

	if user.Activated {
		app.failedValidationResponse(w, r, errors.New("user has already been activated"))
		return
	}

	token, err := app.models.Tokens.New(user.ID, 5*time.Minute, data.ScopeActivation)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	data := envelope{
		"activationToken": token.Plaintext,
		"name":            user.Name,
	}

	// kafka producer handling
	message := &kafkaMessage{
		UserID:       user.ID,
		Email:        user.Email,
		Name:         user.Name,
		Token:        token.Plaintext,
		TemplateName: "user_activation_token.html",
		Task:         "send activation token to user email",
	}

	// change key later (need to test how it works)
	err = app.producer.Produce(message, app.config.kafka.topic, nil, time.Now())
	if err != nil {
		// log or return if cannot produce msg (either bad json format or some problem with brokers)
		app.logger.Error(err.Error())
	}

	app.background(func() {
		err = app.mailer.Send(user.Email, "user_activation_token.html", data)
		if err != nil {
			app.logger.Error(err.Error())
		}
	})

	msg := envelope{"message": "check your email for activation code"}

	err = app.writeJSON(w, http.StatusAccepted, msg, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
