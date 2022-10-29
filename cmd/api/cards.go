package main

import (
	"crypto/ed25519"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/calmitchell617/reserva/internal/data"
	"github.com/calmitchell617/reserva/internal/validator"
)

func (app *application) createCardHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		AccountId    int64  `json:"account_id"`
		Password     string `json:"password"`
		ExpiryInDays int    `json:"expiry_in_days"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	expiry := time.Now().AddDate(0, 0, input.ExpiryInDays)

	card := &data.Card{
		AccountId: input.AccountId,
		Expiry:    expiry,
	}

	err = card.Password.Set(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	v := validator.New()

	if data.ValidateCard(v, card); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	_, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

	card.PrivateKey = privateKey

	requestingBank := app.contextGetBank(r)

	card.Id, err = generateCardNumber()
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

	for i := 0; i < 5; i++ {
		err = app.models.Cards.Insert(card, requestingBank.Id)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrDuplicateCardId):
				if i == 4 {
					app.serverErrorResponse(w, r, errors.New("couldn't generate a unique card number"))
					return
				}
				continue
			default:
				app.serverErrorResponse(w, r, err)
				return
			}
		}
		break
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/cards/%d", card.Id))

	err = app.writeJSON(w, http.StatusCreated, envelope{"card": card}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showCardHandler(w http.ResponseWriter, r *http.Request) {
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

// func (app *application) updateAccountHandler(w http.ResponseWriter, r *http.Request) {
// 	id, err := app.readIDParam(r)
// 	if err != nil {
// 		app.notFoundResponse(w, r)
// 		return
// 	}

// 	requestingBank := app.contextGetBank(r)

// 	account, err := app.models.Accounts.Get(id, requestingBank.Id)
// 	if err != nil {
// 		switch {
// 		case errors.Is(err, data.ErrRecordNotFound):
// 			app.notFoundResponse(w, r)
// 		default:
// 			app.serverErrorResponse(w, r, err)
// 		}
// 		return
// 	}

// 	var input struct {
// 		Frozen         *bool  `json:"frozen"`
// 		BalanceInCents *int64 `json:"balance_in_cents"`
// 	}

// 	err = app.readJSON(w, r, &input)
// 	if err != nil {
// 		app.badRequestResponse(w, r, err)
// 		return
// 	}

// 	if input.Frozen != nil {
// 		account.Frozen = *input.Frozen
// 	}

// 	if input.BalanceInCents != nil {
// 		account.BalanceInCents = *input.BalanceInCents
// 	}

// 	v := validator.New()

// 	if data.ValidateAccount(v, account); !v.Valid() {
// 		app.failedValidationResponse(w, r, v.Errors)
// 		return
// 	}

// 	err = app.models.Accounts.Update(account, requestingBank.Id)
// 	if err != nil {
// 		switch {
// 		case errors.Is(err, data.ErrEditConflict):
// 			app.editConflictResponse(w, r)
// 		default:
// 			app.serverErrorResponse(w, r, err)
// 		}
// 		return
// 	}

// 	err = app.writeJSON(w, http.StatusOK, envelope{"account": account}, nil)
// 	if err != nil {
// 		app.serverErrorResponse(w, r, err)
// 	}
// }

// func (app *application) deleteAccountHandler(w http.ResponseWriter, r *http.Request) {
// 	id, err := app.readIDParam(r)
// 	if err != nil {
// 		app.notFoundResponse(w, r)
// 		return
// 	}

// 	requestingBank := app.contextGetBank(r)

// 	err = app.models.Accounts.Delete(id, requestingBank.Id)
// 	if err != nil {
// 		switch {
// 		case errors.Is(err, data.ErrRecordNotFound):
// 			app.notFoundResponse(w, r)
// 		default:
// 			app.serverErrorResponse(w, r, err)
// 		}
// 		return
// 	}

// 	err = app.writeJSON(w, http.StatusOK, envelope{"message": "account successfully deleted"}, nil)
// 	if err != nil {
// 		app.serverErrorResponse(w, r, err)
// 	}
// }

// func (app *application) listAccountsHandler(w http.ResponseWriter, r *http.Request) {
// 	var input struct {
// 		Frozen bool
// 		data.Filters
// 	}

// 	requestingBank := app.contextGetBank(r)

// 	v := validator.New()

// 	qs := r.URL.Query()

// 	input.Frozen = app.readBool(qs, "frozen", false, v)

// 	input.Filters.Page = app.readInt(qs, "page", 1, v)
// 	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)

// 	input.Filters.Sort = app.readString(qs, "sort", "id")
// 	input.Filters.SortSafelist = []string{"id", "frozen", "-id", "-frozen"}

// 	if data.ValidateFilters(v, input.Filters); !v.Valid() {
// 		app.failedValidationResponse(w, r, v.Errors)
// 		return
// 	}

// 	accounts, metadata, err := app.models.Accounts.GetAll(requestingBank.Id, input.Filters)
// 	if err != nil {
// 		app.serverErrorResponse(w, r, err)
// 		return
// 	}

// 	err = app.writeJSON(w, http.StatusOK, envelope{"accounts": accounts, "metadata": metadata}, nil)
// 	if err != nil {
// 		app.serverErrorResponse(w, r, err)
// 	}
// }
