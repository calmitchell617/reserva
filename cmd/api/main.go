package main

import (
	"context"
	"database/sql"
	"expvar"
	"flag"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"runtime"
	"sync/atomic"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/calmitchell617/reserva/internal/data"

	_ "github.com/lib/pq"
)

type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  time.Duration
	}
	iters            int
	concurrencyLimit int
}

type application struct {
	models data.Models
}

func main() {
	var cfg config

	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")

	flag.StringVar(&cfg.db.dsn, "db-dsn", "", "PostgreSQL DSN")

	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 50, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 50, "PostgreSQL max idle connections")
	flag.DurationVar(&cfg.db.maxIdleTime, "db-max-idle-time", 15*time.Minute, "PostgreSQL max connection idle time")

	flag.IntVar(&cfg.iters, "iters", 10000, "Number of iterations")
	flag.IntVar(&cfg.concurrencyLimit, "concurrency-limit", 10, "Concurrency limit")

	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	db, err := openDB(cfg)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	defer db.Close()

	logger.Info("database connection pool established")

	expvar.Publish("goroutines", expvar.Func(func() any {
		return runtime.NumGoroutine()
	}))

	expvar.Publish("database", expvar.Func(func() any {
		return db.Stats()
	}))

	expvar.Publish("timestamp", expvar.Func(func() any {
		return time.Now().Unix()
	}))

	app := &application{
		models: data.NewModels(db),
	}

	users, err := app.models.Users.GetAll()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	numUsers := len(users)

	start := time.Now()

	eg := errgroup.Group{}

	// set limit
	eg.SetLimit(cfg.concurrencyLimit)

	// initialize atomic counter
	var counter int32

	for int(atomic.LoadInt32(&counter)) < cfg.iters {

		eg.Go(func() error {

			// get two random users
			acquiringUser := users[rand.Intn(numUsers)]
			issuingUser := users[rand.Intn(numUsers)]

			// ensure the users are different
			if acquiringUser.ID == issuingUser.ID {
				return nil
			}

			// get acquiring user from token
			_, err := app.models.Users.GetForToken(acquiringUser.TokenHash)
			if err != nil {
				return err
			}

			// get the issuing account id
			issuingAccountId, err := app.models.Cards.GetAccountIDFromCard(issuingUser.CardID)
			if err != nil {
				return err
			}

			// create a transfer request
			transferRequest := &data.TransferRequest{
				CardId:             acquiringUser.CardID,
				AcquiringAccountID: acquiringUser.AccountID,
				IssuingAccountID:   issuingAccountId,
				Amount:             rand.Int63n(1000),
				CreatedAt:          time.Now(),
			}

			// insert the transfer request
			transferRequest, err = app.models.TransferRequests.Insert(transferRequest)
			if err != nil {
				return err
			}

			transfer := data.Transfer{
				TransferRequestID: transferRequest.ID,
				FromAccountID:     transferRequest.IssuingAccountID,
				ToAccountID:       transferRequest.AcquiringAccountID,
				Amount:            transferRequest.Amount,
				CreatedAt:         time.Now(),
			}

			err = app.models.Transfers.Insert(&transfer)
			if err != nil {
				return err
			}

			atomic.AddInt32(&counter, 1)

			return nil
		})
	}

	err = eg.Wait()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	logger.Info(fmt.Sprintf("%v transfer requests created in %v, rate of %.0f per second", cfg.iters, time.Since(start), float64(cfg.iters)/time.Since(start).Seconds()))
}

func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)
	db.SetConnMaxIdleTime(cfg.db.maxIdleTime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
