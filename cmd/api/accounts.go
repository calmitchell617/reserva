package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/calmitchell617/reserva/internal/data"
	"github.com/calmitchell617/reserva/internal/validator"
)

func (app *application) createAccountHandler(w http.ResponseWriter, r *http.Request) {

	requestingBank := app.contextGetBank(r)

	account := &data.Account{
		BankId: requestingBank.Id,
	}

	v := validator.New()

	if data.ValidateAccount(v, account); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err := app.models.Accounts.Insert(account)
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

	account, err := app.models.Accounts.Get(id, requestingBank.Id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
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

func (app *application) updateAccountHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	requestingBank := app.contextGetBank(r)

	account, err := app.models.Accounts.Get(id, requestingBank.Id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	var input struct {
		Frozen         *bool  `json:"frozen"`
		BalanceInCents *int64 `json:"balance_in_cents"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.Frozen != nil {
		account.Frozen = *input.Frozen
	}

	if input.BalanceInCents != nil {
		account.BalanceInCents = *input.BalanceInCents
	}

	v := validator.New()

	if data.ValidateAccount(v, account); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Accounts.Update(account, requestingBank.Id)
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

func (app *application) deleteAccountHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	requestingBank := app.contextGetBank(r)

	err = app.models.Accounts.Delete(id, requestingBank.Id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "account successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listAccountsHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Frozen bool
		data.Filters
	}

	requestingBank := app.contextGetBank(r)

	v := validator.New()

	qs := r.URL.Query()

	input.Frozen = app.readBool(qs, "frozen", false, v)

	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)

	input.Filters.Sort = app.readString(qs, "sort", "id")
	input.Filters.SortSafelist = []string{"id", "frozen", "-id", "-frozen"}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	accounts, metadata, err := app.models.Accounts.GetAll(requestingBank.Id, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"accounts": accounts, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
