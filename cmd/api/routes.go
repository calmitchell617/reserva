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

	// Banks
	router.HandlerFunc(http.MethodPost, "/v1/banks", app.requireAdminBank(app.registerBankHandler))
	router.HandlerFunc(http.MethodGet, "/v1/banks/:username", app.requireAuthenticatedBank(app.showBankHandler))

	// Accounts
	router.HandlerFunc(http.MethodPost, "/v1/accounts", app.requireAuthenticatedBank(app.createAccountHandler))
	router.HandlerFunc(http.MethodGet, "/v1/accounts/:id", app.requireAuthenticatedBank(app.showAccountHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/accounts/frozen", app.requireAuthenticatedBank(app.updateAccountFrozenHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/accounts/metadata", app.requireAuthenticatedBank(app.updateAccountMetadataHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/accounts/change_money_supply", app.requireAdminBank(app.updateAccountBalanceInCentsHandler))

	// Transfers
	router.HandlerFunc(http.MethodPost, "/v1/transfers", app.requireAuthenticatedBank(app.createTransferHandler))

	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandler)

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)
	router.Handler(http.MethodGet, "/debug/vars", expvar.Handler())

	return app.metrics(app.recoverPanic(app.enableCORS(app.rateLimit(app.authenticate(router)))))
}
