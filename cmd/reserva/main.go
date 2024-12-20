package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"sync/atomic"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/calmitchell617/reserva/internal/data"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

type config struct {
	name string
	db   struct {
		readDsn        string
		writeDsn       string
		hasReadReplica bool
		maxIdleTime    time.Duration
		engine         string
		queryTimeout   time.Duration
	}
	duration         time.Duration
	concurrencyLimit int
	deletes          bool
	kindaRandom      bool
}

type application struct {
	models       data.Models
	queryTimeout time.Duration
}

func main() {
	var cfg config

	flag.StringVar(&cfg.name, "name", "", "Name of system")

	flag.StringVar(&cfg.db.writeDsn, "write-dsn", "", "Write DSN")
	flag.StringVar(&cfg.db.readDsn, "read-dsn", "", "Read DSN")

	flag.DurationVar(&cfg.db.maxIdleTime, "db-max-idle-time", 15*time.Minute, "Max DB connection idle time")

	flag.DurationVar(&cfg.db.queryTimeout, "queryTimeout", 1*time.Minute, "Max DB query time")

	flag.StringVar(&cfg.db.engine, "engine", "", "Database engine")

	flag.DurationVar(&cfg.duration, "duration", 123*time.Minute, "Test duration")
	flag.IntVar(&cfg.concurrencyLimit, "concurrency-limit", 64, "Concurrency limit")
	flag.BoolVar(&cfg.deletes, "deletes", true, "Perform deletes during benchmark")
	flag.BoolVar(&cfg.kindaRandom, "kinda-random", false, "make user selection kinda random")

	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	if cfg.db.engine == "" {
		logger.Error("engine is required")
		os.Exit(1)
	}

	if cfg.name == "" {
		cfg.name = cfg.db.engine
	}

	cfg.db.hasReadReplica = true

	if cfg.db.readDsn == "" {
		cfg.db.hasReadReplica = false
	}

	writeDb, readDb, err := openDB(cfg)
	if err != nil {
		logger.Error(fmt.Errorf("error opening database connection: %w", err).Error())
		os.Exit(1)
	}
	defer writeDb.Close()
	if cfg.db.hasReadReplica {
		defer readDb.Close()
	}

	logger.Info("database connection pool established")

	app := &application{
		models:       data.NewModels(writeDb, readDb, cfg.db.queryTimeout),
		queryTimeout: cfg.db.queryTimeout,
	}

	users, err := app.models.Users.GetAll(cfg.db.engine)
	if err != nil {
		logger.Error(fmt.Errorf("error getting users: %w", err).Error())
		os.Exit(1)
	}

	start := time.Now()

	eg := errgroup.Group{}

	// set limit
	eg.SetLimit(cfg.concurrencyLimit)

	// initialize atomic counter
	var transferCounter int32
	var deleteCounter int32

	transferIds := &SafeInt64Map{
		valMap: make(map[int64]bool, 0),
	}

	logger.Info(fmt.Sprintf("Starting test of %v", cfg.name))

	lastTransferCheckTime := time.Now()
	var lastTransferPlusDeletes int32 = 0

	for time.Since(start) < cfg.duration {

		if time.Since(lastTransferCheckTime) > 3*time.Second {
			transferPlusDeletes := atomic.LoadInt32(&transferCounter) + atomic.LoadInt32(&deleteCounter)
			logger.Info(fmt.Sprintf("%v completing %.0f actions per second", cfg.name, float64(transferPlusDeletes-lastTransferPlusDeletes)/time.Since(lastTransferCheckTime).Seconds()))
			lastTransferCheckTime = time.Now()
			lastTransferPlusDeletes = transferPlusDeletes
		}

		eg.Go(func() error {

			// get a random amount
			amount := rand.Int63n(1000)

			var acquiringUserChoice data.User
			var issuingUserChoice data.User

			// get two random users
			if cfg.kindaRandom {
				_, acquiringUserChoice = users.GetKindaRandom()
				_, issuingUserChoice = users.GetKindaRandom()
			} else {
				_, acquiringUserChoice = users.GetRandom()
				_, issuingUserChoice = users.GetRandom()
			}

			acquiringAccountID := acquiringUserChoice.AccountID

			// ensure the users are different
			if acquiringUserChoice.ID == issuingUserChoice.ID {
				return nil
			}

			// get acquiring user and check permission with token
			users, err := app.models.Users.GetForToken(acquiringUserChoice.Token.Hash, cfg.db.engine)
			if err != nil {
				// err = fmt.Errorf("error getting user -> %w", err)
				// logger.Error(err.Error())
				time.Sleep(time.Second)
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
				err = fmt.Errorf("acquiring user not found or does not have permission")
				logger.Error(err.Error())
				return err
			}

			acquiringUser.AccountID = acquiringAccountID

			// get the issuing account info from the card.. this is the info that would come from a POS terminal or payment gateway
			issuingAccount, card, err := app.models.Accounts.GetFromCard(&issuingUserChoice.Card, cfg.db.engine)
			if err != nil {
				err = fmt.Errorf("error getting account from card -> %w", err)
				logger.Error(err.Error())
				return err
			}

			// check account balance
			if issuingAccount.Balance < amount {
				err = errors.New("issuing account has insufficient funds")
				logger.Error(err.Error())
				return err
			}

			// check account frozen status
			if issuingAccount.Frozen {
				err = errors.New("issuing account is frozen")
				logger.Error(err.Error())
				return err
			}

			// check card frozen status
			if card.Frozen {
				err = errors.New("issuing card is frozen")
				logger.Error(err.Error())
				return err
			}

			// check card expiration date
			if card.ExpirationDate.Before(time.Now()) {
				err = errors.New("issuing card is expired")
				logger.Error(err.Error())
				return err
			}

			// check card security code
			if card.SecurityCode != issuingUserChoice.Card.SecurityCode {
				err = errors.New("issuing card security code does not match")
				logger.Error(err.Error())
				return err
			}

			// at this point, the issuing org will have to approve the transfer request.
			// They would be sent data about the account to a known endpoint, and would respond with a token if they approve the transfer.

			// get issuing user and check permission with token
			users, err = app.models.Users.GetForToken(issuingUserChoice.Token.Hash, cfg.db.engine)
			if err != nil {
				err = fmt.Errorf("error getting user -> %w", err)
				logger.Error(err.Error())
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
				err = errors.New("issuing user not found or does not have permission")
				logger.Error(err.Error())
				return err
			}

			// check if the issuing user is in the same organization as the issuing account
			if issuingUser.OrganizationID != issuingAccount.OrganizationID {
				err = errors.New("issuing user is not in the same organization as the issuing account")
				logger.Error(err.Error())
				return err
			}

			// create a transfer
			transfer := &data.Transfer{
				CardID:         card.ID,
				FromAccountID:  issuingAccount.ID,
				ToAccountID:    acquiringUser.AccountID,
				RequestingUser: *acquiringUser,
				Amount:         amount,
				CreatedAt:      time.Now(),
			}

			_, err = app.models.Transfers.TransferFunds(transfer, cfg.db.engine)
			if err != nil {
				err = fmt.Errorf("error transferring funds -> %w", err)
				logger.Error(err.Error())
				return err
			}

			atomic.AddInt32(&transferCounter, 1)

			if cfg.deletes {
				transferIds.Add(transfer.ID)

				if transferCounter != 0 && transferCounter%20 == 0 {
					toDeleteElement, err := transferIds.GetRandom()
					if err != nil {
						err = fmt.Errorf("error getting random transfer -> %w", err)
						logger.Error(err.Error())
						return err
					}
					err = app.models.Transfers.Delete(toDeleteElement, cfg.db.engine)
					if err != nil {
						err = fmt.Errorf("error deleting transfer -> %w", err)
						logger.Error(err.Error())
						return err
					}

					transferIds.Remove(toDeleteElement)

					atomic.AddInt32(&deleteCounter, 1)
				}
			}

			return nil
		})
	}

	err = eg.Wait()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	totalActions := transferCounter + deleteCounter

	logger.Info(fmt.Sprintf("%v completed %v actions in %v, rate of %.0f per second", cfg.name, totalActions, cfg.duration, float64(totalActions)/time.Since(start).Seconds()))
}

func openDB(cfg config) (writeDb *sql.DB, readDb *sql.DB, err error) {

	var driver string

	switch cfg.db.engine {
	case "postgresql":
		driver = "postgres"
	case "mariadb", "mysql":
		driver = "mysql"
	default:
		return nil, nil, fmt.Errorf("unsupported database engine")
	}

	writeDb, err = sql.Open(driver, cfg.db.writeDsn)
	if err != nil {
		return nil, nil, err
	}

	writeDb.SetMaxOpenConns(cfg.concurrencyLimit)
	writeDb.SetMaxIdleConns(cfg.concurrencyLimit)
	writeDb.SetConnMaxIdleTime(cfg.db.maxIdleTime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = writeDb.PingContext(ctx)
	if err != nil {
		writeDb.Close()
		return nil, nil, err
	}

	readDb = writeDb

	if cfg.db.hasReadReplica {
		readDb, err = sql.Open(driver, cfg.db.readDsn)
		if err != nil {
			return nil, nil, err
		}

		readDb.SetMaxOpenConns(cfg.concurrencyLimit)
		readDb.SetMaxIdleConns(cfg.concurrencyLimit)
		readDb.SetConnMaxIdleTime(cfg.db.maxIdleTime)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = readDb.PingContext(ctx)
		if err != nil {
			readDb.Close()
			return nil, nil, err
		}
	}

	return writeDb, readDb, nil
}
