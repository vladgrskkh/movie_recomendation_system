package main

import (
	"errors"
	"net/http"

	"github.com/invopop/validation"
	"github.com/invopop/validation/is"
	"github.com/vladgrskkh/movie_recomendation_system/internal/data"
)

type deleteUserInput struct {
	Email    string `json:"email" example:"something@example.com"`
	Password string `json:"password" example:"s1mplepA$$word"`
}

// DeleteUser godoc
//
// @Summary Delete a user account
// @Description Deletes a user account after validating email and password
// @Tags users
// @Accept json
// @Produce json
// @Param user body deleteUserInput true "User deletion payload"
// @Success 200 {object} map[string]string "OK | Example {\"message\": \"user successfully deleted\"}"
// @Failure 400 {object} map[string]string "Bad Request | Example {\"error\": \"body contains badly-formated JSON\"}"
// @Failure 401 {object} map[string]string "Unauthorized | Example {\"error\": \"invalid authentication credentials\"}"
// @Failure 403 {object} map[string]string "Forbidden | Example {\"error\": \"this resourse doesn't belong to authenticated user\"}"
// @Failure 422 {object} map[string]string "Unprocessable Entity | Example {\"error\": \"validation error\"}"
// @Failure 500 {object} map[string]string "Internal Server Error | Example {\"error\": \"server encountered a problem and could not process your request\"}"
// @Security BearerAuth
// @Router /users [delete]
func (app *application) deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	var input deleteUserInput

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

	// check if deletion performs by the same user
	reqUser := app.contextGetUser(r)
	if reqUser.ID != user.ID {
		app.forbiddenUserResponse(w, r)
		return
	}

	match, err := user.Password.Matches(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	if !match {
		app.invalidCredentialResponse(w, r)
		return
	}

	err = app.models.Users.Delete(user)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "user successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
