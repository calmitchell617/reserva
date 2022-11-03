package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/calmitchell617/reserva/internal/validator"

	"github.com/calmitchell617/reserva/internal/data"
)

func (app *application) createTransferHandler(w http.ResponseWriter, r *http.Request) {
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

	v := validator.New()

	if data.ValidateTransfer(v, transfer); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// sourceAccount, err := app.models.Accounts.Get(input.SourceAccountId)
	// if err != nil {
	// 	switch {
	// 	case errors.Is(err, data.ErrRecordNotFound):
	// 		app.notFoundResponse(w, r)
	// 	default:
	// 		app.serverErrorResponse(w, r, err)
	// 	}
	// 	return
	// }

	// if sourceAccount.ControllingBank != requestingBank.Username && !requestingBank.Admin {
	// 	app.notPermittedResponse(w, r)
	// 	return
	// }

	// if sourceAccount.BalanceInCents < transfer.AmountInCents {
	// 	app.insufficentFundsResponse(w, r)
	// 	return
	// }

	transfer, err = app.models.Transfers.Insert(transfer, *requestingBank)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/transfers/%d", transfer.Id))

	err = app.writeJSON(w, http.StatusCreated, envelope{"transfer": transfer}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
