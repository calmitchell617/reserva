package main

import (
	"expvar"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)

	// in prod, banks will be created by a backend UI
	router.HandlerFunc(http.MethodPost, "/v1/banks", app.requireAdminBank(app.registerBankHandler))
	router.HandlerFunc(http.MethodGet, "/v1/banks", app.requireAuthenticatedBank(app.showBankHandler))

	router.HandlerFunc(http.MethodGet, "/v1/accounts", app.requireUnfrozenBank(app.listAccountsHandler))
	router.HandlerFunc(http.MethodPost, "/v1/accounts", app.requireUnfrozenBank(app.createAccountHandler))
	router.HandlerFunc(http.MethodGet, "/v1/accounts/:id", app.requireUnfrozenBank(app.showAccountHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/accounts/:id", app.requireUnfrozenBank(app.updateAccountHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/accounts/:id", app.requireUnfrozenBank(app.deleteAccountHandler))

	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandler)

	router.Handler(http.MethodGet, "/debug/vars", expvar.Handler())

	return app.metrics(app.recoverPanic(app.enableCORS(app.rateLimit(app.authenticate(router)))))
}
