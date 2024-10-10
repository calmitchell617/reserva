package main

import (
	"context"
	"net/http"

	"github.com/calmitchell617/reserva/internal/data"
)

type contextKey string

const caretakerContextKey = contextKey("caretaker")

func (app *application) contextSetCaretaker(r *http.Request, caretaker *data.Caretaker) *http.Request {
	ctx := context.WithValue(r.Context(), caretakerContextKey, caretaker)
	return r.WithContext(ctx)
}

func (app *application) contextGetCaretaker(r *http.Request) *data.Caretaker {
	caretaker, ok := r.Context().Value(caretakerContextKey).(*data.Caretaker)
	if !ok {
		panic("missing caretaker value in request context")
	}

	return caretaker
}
