package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/calmitchell617/reserva/internal/validator"

	"github.com/calmitchell617/reserva/internal/data"
)

func (app *application) createAccountHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		MetaData string `json:"metadata"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	requestingBank := app.contextGetBank(r)
	account := &data.Account{
		Metadata:        input.MetaData,
		ControllingBank: requestingBank.Username,
	}

	v := validator.New()

	if data.ValidateAccount(v, account); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	account.Id, err = app.models.Accounts.Insert(account)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/accounts/%d", account.Id))

	err = app.writeJSON(w, http.StatusCreated, envelope{"account": account}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showAccountHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	requestingBank := app.contextGetBank(r)

	account, err := app.models.Accounts.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	if account.ControllingBank != requestingBank.Username && !requestingBank.Admin {
		app.notFoundResponse(w, r)
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"account": account}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateAccountFrozenHandler(w http.ResponseWriter, r *http.Request) {
	requestingBank := app.contextGetBank(r)

	var input struct {
		Id     int64 `json:"id"`
		Frozen bool  `json:"frozen"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	account, err := app.models.Accounts.Get(input.Id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	if account.ControllingBank != requestingBank.Username && !requestingBank.Admin {
		app.notFoundResponse(w, r)
	}

	account.Frozen = input.Frozen

	err = app.models.Accounts.Update(account)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"account": account}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateAccountMetadataHandler(w http.ResponseWriter, r *http.Request) {
	requestingBank := app.contextGetBank(r)

	var input struct {
		Id       int64  `json:"id"`
		Metadata string `json:"metadata"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()

	if data.ValidateAccountMetadata(v, input.Metadata); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	account, err := app.models.Accounts.Get(input.Id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	if account.ControllingBank != requestingBank.Username && !requestingBank.Admin {
		app.notFoundResponse(w, r)
	}

	account.Metadata = input.Metadata

	err = app.models.Accounts.Update(account)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"account": account}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateAccountBalanceInCentsHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Id            int64 `json:"id"`
		ChangeInCents int64 `json:"change_in_cents"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	account, err := app.models.Accounts.Get(input.Id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	account.BalanceInCents += input.ChangeInCents

	err = app.models.Accounts.Update(account)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"account": account}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
