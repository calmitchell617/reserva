package main

import (
	"errors"
	"net/http"

	"github.com/calmitchell617/reserva/internal/data"
	"github.com/calmitchell617/reserva/internal/validator"
)

func (app *application) registerBankHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Username string `json:"username"`
		Admin    bool   `json:"admin"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	bank := &data.Bank{
		Username: input.Username,
		Admin:    input.Admin,
	}

	err = bank.Password.Set(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	v := validator.New()

	if data.ValidateBank(v, bank); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Banks.Insert(bank)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateUsername):
			v.AddError("username", "a bank with this username already exists")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusAccepted, envelope{"bank": bank}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showBankHandler(w http.ResponseWriter, r *http.Request) {
	username, err := app.readUsernameParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	requestingBank := app.contextGetBank(r)

	if !requestingBank.Admin && requestingBank.Username != username {
		app.notPermittedResponse(w, r)
		return
	}

	bank, err := app.models.Banks.GetByUsername(username)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"bank": bank}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
