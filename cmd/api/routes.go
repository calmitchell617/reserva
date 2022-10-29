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
	if app.config.env == "development" {
		router.HandlerFunc(http.MethodPost, "/v1/banks", app.registerBankHandler)
	}
	router.HandlerFunc(http.MethodGet, "/v1/banks", app.requireActivatedBank(app.showBankHandler))
	router.HandlerFunc(http.MethodPut, "/v1/banks/activate", app.activateBankHandler)
	router.HandlerFunc(http.MethodPut, "/v1/banks/update-password", app.updateBankPasswordHandler)

	router.HandlerFunc(http.MethodGet, "/v1/accounts", app.requireActivatedBank(app.listAccountsHandler))
	router.HandlerFunc(http.MethodPost, "/v1/accounts", app.requireActivatedBank(app.createAccountHandler))
	router.HandlerFunc(http.MethodGet, "/v1/accounts/:id", app.requireActivatedBank(app.showAccountHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/accounts/:id", app.requireActivatedBank(app.updateAccountHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/accounts/:id", app.requireActivatedBank(app.deleteAccountHandler))

	router.HandlerFunc(http.MethodPost, "/v1/cards", app.requireActivatedBank(app.createCardHandler))

	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandler)
	router.HandlerFunc(http.MethodPost, "/v1/tokens/activation", app.createActivationTokenHandler)
	router.HandlerFunc(http.MethodPost, "/v1/tokens/reset-password", app.createPasswordResetTokenHandler)

	router.Handler(http.MethodGet, "/debug/vars", expvar.Handler())

	return app.metrics(app.recoverPanic(app.enableCORS(app.rateLimit(app.authenticate(router)))))
}
