package main

// This file contains the /v1/account API handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/calmitchell617/reserva/internal/validator"

	"github.com/calmitchell617/reserva/internal/data"
)

func (app *application) createAccountHandler(w http.ResponseWriter, r *http.Request) {
	// Creates an account
	var input struct {
		MetaData string `json:"metadata"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// What bank is requesting this?
	requestingBank := app.contextGetBank(r)
	account := &data.Account{
		Metadata:        input.MetaData,
		ControllingBank: requestingBank.Username,
	}

	v := validator.New()

	// Validate the account input data
	if data.ValidateAccount(v, account); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Insert the account into the DB
	account.Id, err = app.models.Accounts.Insert(account)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Respond
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/accounts/%d", account.Id))

	err = app.writeJSON(w, http.StatusCreated, envelope{"account": account}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showAccountHandler(w http.ResponseWriter, r *http.Request) {
	// gets an account's info

	// Get account ID
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Get the bank that requested this
	requestingBank := app.contextGetBank(r)

	// Get the account
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

	// If the bank is not the account's controlling bank, exit
	if account.ControllingBank != requestingBank.Username && !requestingBank.Admin {
		app.notFoundResponse(w, r)
	}

	// write response
	err = app.writeJSON(w, http.StatusOK, envelope{"account": account}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateAccountFrozenHandler(w http.ResponseWriter, r *http.Request) {
	// freezes / unfreezes an account

	// get requesting bank
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

	// get the requested account
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

	// if the banks doesn't control this account, exit
	if account.ControllingBank != requestingBank.Username && !requestingBank.Admin {
		app.notFoundResponse(w, r)
	}

	// change frozen value
	account.Frozen = input.Frozen

	// update in DB
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

	// write response
	err = app.writeJSON(w, http.StatusOK, envelope{"account": account}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateAccountMetadataHandler(w http.ResponseWriter, r *http.Request) {
	// updates accounts metadata (JSON) field

	// get requesting bank
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

	// validate the account metadata JSON syntax
	v := validator.New()

	if data.ValidateAccountMetadata(v, input.Metadata); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// get the account
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

	// check that the requesting bank controls the requested account
	if account.ControllingBank != requestingBank.Username && !requestingBank.Admin {
		app.notFoundResponse(w, r)
	}

	// update metadata
	account.Metadata = input.Metadata

	// update account in DB
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

	// write response
	err = app.writeJSON(w, http.StatusOK, envelope{"account": account}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateAccountBalanceInCentsHandler(w http.ResponseWriter, r *http.Request) {
	// updates an account's balance without an associated transfer - this alters the total money supply
	// only central banks can access this route

	var input struct {
		Id            int64 `json:"id"`
		ChangeInCents int64 `json:"change_in_cents"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// get account
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

	// change balance value
	account.BalanceInCents = account.BalanceInCents + input.ChangeInCents

	// update in DB
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

	// respond
	err = app.writeJSON(w, http.StatusOK, envelope{"account": account}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
