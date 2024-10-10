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

	router.HandlerFunc(http.MethodGet, "/api/v1/healthcheck", app.healthcheckHandler)

	router.HandlerFunc(http.MethodPost, "/api/v1/users", app.registerUserHandler) // admins only
	// show user. admins can see all, standard can see self
	// list users. admins only
	// update user. admins can update all, standard can update self
	router.HandlerFunc(http.MethodPut, "/api/v1/users/activate", app.activateUserHandler)       // all can activate self
	router.HandlerFunc(http.MethodPut, "/api/v1/users/password", app.updateUserPasswordHandler) // all can update their own password
	// delete user. admins only

	router.HandlerFunc(http.MethodPost, "/api/v1/accounts", app.requireActiveUser(app.createAccountHandler))  // all
	router.HandlerFunc(http.MethodGet, "/api/v1/accounts/:id", app.requireActiveUser(app.showAccountHandler)) // admins can see all, standard can see own
	// list accounts. admins can see all, standard can see own
	// update account. admins only
	// delete account. admins can delete all, standard can delete own

	router.HandlerFunc(http.MethodPost, "/api/v1/account-holders", app.requireActiveUser(app.createAccountHolderHandler))
	// delete account holder

	// associate account with account holder(s)
	// dissociate account with account holder(s)

	// create card
	// show card
	// list cards
	// update card
	// delete card

	// create transfer request
	// create transfer

	// create hold request
	// create hold

	// delete account

	router.HandlerFunc(http.MethodPost, "/api/v1/tokens/authentication", app.createAuthenticationTokenHandler)
	router.HandlerFunc(http.MethodPost, "/api/v1/tokens/activation", app.createActivationTokenHandler)
	router.HandlerFunc(http.MethodPost, "/api/v1/tokens/password-reset", app.createPasswordResetTokenHandler)

	router.Handler(http.MethodGet, "/debug/vars", expvar.Handler())

	return app.metrics(app.recoverPanic(app.enableCORS(app.rateLimit(app.authenticate(router)))))

	// router.HandlerFunc(http.MethodGet, "/api/v1/movies", app.requireActiveUser(app.listMoviesHandler))
	// router.HandlerFunc(http.MethodPost, "/api/v1/movies", app.requireActiveUser(app.createMovieHandler))
	// router.HandlerFunc(http.MethodGet, "/api/v1/movies/:id", app.requireActiveUser(app.showMovieHandler))
	// router.HandlerFunc(http.MethodPatch, "/api/v1/movies/:id", app.requireActiveUser(app.updateMovieHandler))
	// router.HandlerFunc(http.MethodDelete, "/api/v1/movies/:id", app.requireActiveUser(app.deleteMovieHandler))

}
