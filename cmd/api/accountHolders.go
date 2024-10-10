package main

import (
	"fmt"
	"net/http"

	"github.com/calmitchell617/reserva/internal/data"
	"github.com/calmitchell617/reserva/internal/validator"
)

func (app *application) createAccountHolderHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ExternalID string `json:"external_id"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	accountHolder := &data.AccountHolder{
		ExternalID: input.ExternalID,
	}

	v := validator.New()

	if data.ValidateAccountHolder(v, accountHolder); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.AccountHolders.Insert(accountHolder)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/account-holder/%d", accountHolder.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"account-holder": accountHolder}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
