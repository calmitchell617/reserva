package main

import (
	"context"
	"net/http"

	"github.com/calmitchell617/reserva/internal/data"
)

type contextKey string

const bankContextKey = contextKey("bank")

func (app *application) contextSetBank(r *http.Request, bank *data.Bank) *http.Request {
	ctx := context.WithValue(r.Context(), bankContextKey, bank)
	return r.WithContext(ctx)
}

func (app *application) contextGetBank(r *http.Request) *data.Bank {
	bank, ok := r.Context().Value(bankContextKey).(*data.Bank)
	if !ok {
		panic("missing bank value in request context")
	}

	return bank
}
