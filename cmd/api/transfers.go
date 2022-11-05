package main

// contains handler to create money transfers

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/calmitchell617/reserva/internal/validator"

	"github.com/calmitchell617/reserva/internal/data"
)

func (app *application) createTransferHandler(w http.ResponseWriter, r *http.Request) {
	// only controlling bank can initiate a transfer from a given account
	requestingBank := app.contextGetBank(r)

	var input struct {
		SourceAccountId int64 `json:"source_account_id"`
		TargetAccountId int64 `json:"target_account_id"`
		AmountInCents   int64 `json:"amount_in_cents"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	transfer := &data.Transfer{
		SourceAccountId: input.SourceAccountId,
		TargetAccountId: input.TargetAccountId,
		AmountInCents:   input.AmountInCents,
		CreatedAt:       time.Now(),
	}

	// validate transfer object
	v := validator.New()

	if data.ValidateTransfer(v, transfer); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// insert the transfer
	transfer, err = app.models.Transfers.Insert(transfer, *requestingBank)
	if err != nil {
		switch {
		case errors.Is(err, data.NoPermission):
			app.notPermittedResponse(w, r)
		case errors.Is(err, data.InsufficentFunds):
			app.insufficentFundsResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/transfers/%d", transfer.Id))

	err = app.writeJSON(w, http.StatusCreated, envelope{"transfer": transfer}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
