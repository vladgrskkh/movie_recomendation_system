package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/invopop/validation"
	"github.com/invopop/validation/is"

	"github.com/vladgrskkh/movie_recomendation_system/internal/data"
	"github.com/vladgrskkh/movie_recomendation_system/internal/validate"
)

// HealthCheck godoc
// @Summary Health check
// @Description Returns service health and metadata
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string
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
// @Summary Get a movie by ID
// @Description Returns a single movie by numeric ID
// @Tags movies
// @Produce json
// @Param movieID path int true "Movie ID"
// @Success 200 {object} data.Movie
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Security BearerAuth
// @Router /movie/{movieID} [get]
func (app *application) getMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
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
// @Summary Create a movie
// @Description Create a new movie (admin only)
// @Tags movies
// @Accept json
// @Produce json
// @Param movie body movieInput true "Movie payload"
// @Success 201 {object} data.Movie
// @Failure 400 {object} map[string]string
// @Failure 422 {object} map[string]string
// @Failure 500 {object} map[string]string
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
// @Summary Delete a movie
// @Description Delete movie by ID (admin only)
// @Tags movies
// @Produce json
// @Param movieID path int true "Movie ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /movie/{movieID} [delete]
func (app *application) deleteMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	err = app.models.Movies.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.badRequestResponse(w, r, err)
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
// @Summary Update a movie
// @Description Patch movie by ID
// @Tags movies
// @Accept json
// @Produce json
// @Param movieID path int true "Movie ID"
// @Param movie body movieInput true "Partial movie payload"
// @Success 200 {object} data.Movie
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 422 {object} map[string]string
// @Failure 500 {object} map[string]string
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

// ListMovies godoc
// @Summary List movies
// @Description Get all movies (may support pagination)
// @Tags movies
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string][]object
// @Router /movie [get]
func (app *application) getAllMoviesHandler(w http.ResponseWriter, r *http.Request) {

}

type registerInput struct {
	Name     string `json:"name" example:"John Doe"`
	Email    string `json:"email" example:"something@example.com"`
	Password string `json:"password" example:"s1mplepA$$word"`
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
// @Failure 400 {object} map[string]string
// @Failure 422 {object} map[string]string
// @Failure 500 {object} map[string]string
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
			app.badRequestResponse(w, r, err)
		default:
			app.serverErrorResponse(w, r, err)
		}

		return
	}

	token, err := app.models.Tokens.New(user.ID, 72*time.Hour, data.ScopeActivation)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// SMTP

	data := envelope{
		"activationToken": token.Plaintext,
		"userID":          user.ID,
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
// @Summary Activate user
// @Description Activates a user account using activation token
// @Tags users
// @Accept json
// @Produce json
// @Param activation body activateInput true "Activation payload"
// @Success 200 {object} data.User
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 422 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /users/activate [put]
func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	var input activateInput

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	err = validation.Validate(input.Token, validation.Required, validation.Length(26, 26))
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

	err = app.writeJSON(w, http.StatusOK, envelope{"user": user}, nil)
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

// refreshTokenHandler wants refresh token to create auth token and new refresh token
// auth token is jwt and refresh token is high entropy string
// RefreshToken godoc
// @Summary Refresh tokens
// @Description Exchange refresh token for new auth and refresh tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param token body refreshInput true "Refresh token payload"
// @Success 201 {object} tokenPair
// @Failure 400 {object} map[string]string
// @Failure 422 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /tokens/refresh [put]
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
	authToken, err := createToken(user.ID)
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

// createAuthenticationTokenHandler is log in for app
// every time user log in we will create new auth token and refresh token(deleting prev refresh token if exists)
// CreateAuthToken godoc
// @Summary Log in and get tokens
// @Description Creates authentication and refresh tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body loginInput true "Login payload"
// @Success 201 {object} tokenPair
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 422 {object} map[string]string
// @Failure 500 {object} map[string]string
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
	authToken, err := createToken(user.ID)
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

// Predict Handler godoc
// @Summary Get predict movie
// @Description Validates movie input and predict movie
// @Tags movies
// @Accept json
// @Produce json
// @Param credentials body object true "Moive payload"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /movie/predict [post]
func (app *application) predictHandler(w http.ResponseWriter, r *http.Request) {
	mock := envelope{
		"predict": "movie_id number",
	}

	err := app.writeJSON(w, http.StatusOK, mock, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
