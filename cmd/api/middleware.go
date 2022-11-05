package main

import (
	"errors"
	"expvar"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/calmitchell617/reserva/internal/data"
	"github.com/calmitchell617/reserva/internal/validator"

	"github.com/felixge/httpsnoop"
	"github.com/tomasen/realip"
	"golang.org/x/time/rate"
)

func (app *application) recoverPanic(next http.Handler) http.Handler {
	// Don't panic
	// 42
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (app *application) rateLimit(next http.Handler) http.Handler {
	// Someone will definitely try to DDOS any central bank digital currency
	// This implements token bucket rate limiting technique
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	go func() {
		for {
			time.Sleep(time.Minute)

			mu.Lock()

			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}

			mu.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if app.config.limiter.enabled {
			ip := realip.FromRequest(r)

			mu.Lock()

			if _, found := clients[ip]; !found {
				clients[ip] = &client{
					limiter: rate.NewLimiter(rate.Limit(app.config.limiter.rps), app.config.limiter.burst),
				}
			}

			clients[ip].lastSeen = time.Now()

			if !clients[ip].limiter.Allow() {
				mu.Unlock()
				app.rateLimitExceededResponse(w, r)
				return
			}

			mu.Unlock()
		}

		next.ServeHTTP(w, r)
	})
}

func (app *application) authenticate(next http.Handler) http.Handler {
	// Check what bank is making this request, add that bank object to the request context

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Authorization")

		authorizationHeader := r.Header.Get("Authorization")

		if authorizationHeader == "" {
			// if there is no bearer token, then it is an anonymous request
			r = app.contextSetBank(r, data.AnonymousBank)
			next.ServeHTTP(w, r)
			return
		}

		// get the bearer token and split them up
		headerParts := strings.Split(authorizationHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		token := headerParts[1]

		// validate the token
		v := validator.New()

		if data.ValidateTokenPlaintext(v, token); !v.Valid() {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		// get the bank from that token
		bank, err := app.models.Banks.GetByTokenForAuthentication(data.ScopeAuthentication, token)
		if err != nil {
			switch {
			// if there is no bank that matches, send invalid credentials response
			case errors.Is(err, data.ErrRecordNotFound):
				app.invalidAuthenticationTokenResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}

		// set the context's bank object and pass it along
		r = app.contextSetBank(r, bank)

		next.ServeHTTP(w, r)
	})
}

func (app *application) requireAuthenticatedBank(next http.HandlerFunc) http.HandlerFunc {
	// This requires an authenticated (non anonymous) bank
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bank := app.contextGetBank(r)

		if bank.IsAnonymous() {
			app.authenticationRequiredResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (app *application) requireAdminBank(next http.HandlerFunc) http.HandlerFunc {
	// requies a bank that has admin priviledges
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bank := app.contextGetBank(r)

		if !bank.Admin {
			app.notPermittedResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})

	return app.requireAuthenticatedBank(fn)
}

func (app *application) metrics(next http.Handler) http.Handler {
	// publish some stats for telemtry
	totalRequestsReceived := expvar.NewInt("total_requests_received")
	totalResponsesSent := expvar.NewInt("total_responses_sent")
	totalProcessingTimeMicroseconds := expvar.NewInt("total_processing_time_μs")

	totalResponsesSentByStatus := expvar.NewMap("total_responses_sent_by_status")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		totalRequestsReceived.Add(1)

		metrics := httpsnoop.CaptureMetrics(next, w, r)

		totalResponsesSent.Add(1)

		totalProcessingTimeMicroseconds.Add(metrics.Duration.Microseconds())

		totalResponsesSentByStatus.Add(strconv.Itoa(metrics.Code), 1)
	})
}
