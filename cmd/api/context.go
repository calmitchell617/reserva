package main

import (
	"context"
	"net/http"

	"github.com/calmitchell617/reserva/internal/data"
)

type contextKey string

const bankContextKey = contextKey("bank")

// adds the requesting bank to the request context for retreival by handlers
func (app *application) contextSetBank(r *http.Request, bank *data.Bank) *http.Request {
	ctx := context.WithValue(r.Context(), bankContextKey, bank)
	return r.WithContext(ctx)
}

func (app *application) contextGetBank(r *http.Request) *data.Bank {
	// gets requesting bank from request context (for usage in handlers)
	bank, ok := r.Context().Value(bankContextKey).(*data.Bank)
	if !ok {
		// this shouldn't ever happen. If it did, then authentication was skipped for a route
		panic("missing bank value in request context")
	}

	return bank
}
