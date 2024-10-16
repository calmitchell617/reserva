package main

import (
	"context"
	"database/sql"
	"errors"
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

	_ "github.com/go-sql-driver/mysql"
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
		engine       string
	}
	duration         time.Duration
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

	flag.StringVar(&cfg.db.engine, "engine", "", "Database engine")

	flag.DurationVar(&cfg.duration, "duration", 12*time.Hour, "Test duration")
	flag.IntVar(&cfg.concurrencyLimit, "concurrency-limit", 50, "Concurrency limit")

	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	if cfg.db.engine == "" {
		logger.Error("engine is required")
		os.Exit(1)
	}

	db, err := openDB(cfg)
	if err != nil {
		logger.Error(fmt.Errorf("error opening database connection: %w", err).Error())
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

	users, err := app.models.Users.GetAll(cfg.db.engine)
	if err != nil {
		logger.Error(fmt.Errorf("error getting users: %w", err).Error())
		os.Exit(1)
	}

	numUsers := len(users)

	start := time.Now()

	eg := errgroup.Group{}

	// set limit
	eg.SetLimit(cfg.concurrencyLimit)

	// initialize atomic counter
	var counter int32

	fmt.Printf("Starting test\n\n")

	for time.Since(start) < cfg.duration {

		eg.Go(func() error {

			// get a random amount
			amount := rand.Int63n(1000)

			// get two random users
			acquiringUserChoice := users[rand.Intn(numUsers)]
			acquiringAccountID := acquiringUserChoice.AccountID
			issuingUserChoice := users[rand.Intn(numUsers)]

			// ensure the users are different
			if acquiringUserChoice.ID == issuingUserChoice.ID {
				return nil
			}

			// get acquiring user and check permission with token
			users, err := app.models.Users.GetForToken(acquiringUserChoice.Token.Hash, cfg.db.engine)
			if err != nil {
				fmt.Printf("error getting user -> %v\n", err)
				return err
			}

			// make sure the acquiring user has permission to request payments
			var acquiringUser *data.User

			for _, user := range users {
				if user.Token.PermissionID == 1 {
					acquiringUser = user
					break
				}
			}

			if acquiringUser == nil {
				fmt.Printf("acquiring user not found or does not have permission\n")
				return errors.New("acquiring user not found or does not have permission")
			}

			acquiringUser.AccountID = acquiringAccountID

			// get the issuing account info from the card.. this is the info that would come from a POS terminal or payment gateway
			issuingAccount, card, err := app.models.Accounts.GetFromCard(&issuingUserChoice.Card, cfg.db.engine)
			if err != nil {
				fmt.Printf("error getting account from card -> %v\n", err)
				return err
			}

			// check account balance
			if issuingAccount.Balance < amount {
				fmt.Printf("issuing account has insufficient funds\n")
				return errors.New("issuing account has insufficient funds")
			}

			// check account frozen status
			if issuingAccount.Frozen {
				fmt.Printf("issuing account is frozen\n")
				return errors.New("issuing account is frozen")
			}

			// check card frozen status
			if card.Frozen {
				fmt.Printf("issuing card is frozen\n")
				return errors.New("issuing card is frozen")
			}

			// check card expiration date
			if card.ExpirationDate.Before(time.Now()) {
				fmt.Printf("issuing card is expired\n")
				return errors.New("issuing card is expired")
			}

			// check card security code
			if card.SecurityCode != issuingUserChoice.Card.SecurityCode {
				fmt.Printf("issuing card security code does not match\n")
				return errors.New("issuing card security code does not match")
			}

			// at this point, the issuing org will have to approve the transfer request.
			// They would be sent data about the account to a known endpoint, and would respond with a token if they approve the transfer.

			// get issuing user and check permission with token
			users, err = app.models.Users.GetForToken(issuingUserChoice.Token.Hash, cfg.db.engine)
			if err != nil {
				fmt.Printf("error getting user -> %v\n", err)
				return err
			}

			var issuingUser *data.User

			// make sure they have permission to approve transfer requests
			for _, user := range users {
				if user.Token.PermissionID == 1 {
					issuingUser = user
					break
				}
			}

			if issuingUser == nil {
				fmt.Printf("issuing user not found or does not have permission\n")
				return errors.New("issuing user not found or does not have permission")
			}

			// check if the issuing user is in the same organization as the issuing account
			if issuingUser.OrganizationID != issuingAccount.OrganizationID {
				fmt.Printf("issuing user does not have permission\n")
				return errors.New("issuing user does not have permission")
			}

			// create a transfer
			transfer := data.Transfer{
				CardID:         card.ID,
				FromAccountID:  issuingAccount.ID,
				ToAccountID:    acquiringUser.AccountID,
				RequestingUser: *acquiringUser,
				Amount:         amount,
				CreatedAt:      time.Now(),
			}

			err = app.models.Transfers.TransferFunds(&transfer, cfg.db.engine)
			if err != nil {
				fmt.Printf("error transferring funds -> %v\n", err)
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

	logger.Info(fmt.Sprintf("%v transfer requests created in %v, rate of %.0f per second", counter, cfg.duration, float64(counter)/time.Since(start).Seconds()))

	// query transfers table to get the number of transfers per minute

}

func openDB(cfg config) (*sql.DB, error) {

	var driver string

	switch cfg.db.engine {
	case "postgresql":
		driver = "postgres"
	case "mariadb":
		driver = "mysql"
	default:
		return nil, fmt.Errorf("unsupported database engine")
	}

	db, err := sql.Open(driver, cfg.db.dsn)
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
