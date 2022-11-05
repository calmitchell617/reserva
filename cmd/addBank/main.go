package main

// This program adds a bank to the Singlestore backend DB

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/calmitchell617/reserva/internal/data"
	"github.com/calmitchell617/reserva/internal/jsonlog"
	"github.com/calmitchell617/reserva/internal/validator"
	"github.com/calmitchell617/reserva/internal/vcs" // New import

	_ "github.com/go-sql-driver/mysql"
)

var (
	version = vcs.Version()
)

type config struct {
	db struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
	cache struct {
		host     string
		port     int
		password string
		db       int
	}
	bank struct {
		username string
		password string
		admin    bool
	}
}

func main() {
	var cfg config

	// Get command line flags
	flag.StringVar(&cfg.db.dsn, "db-dsn", "", "DB DSN")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "DB max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "DB max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "DB max connection idle time")

	flag.StringVar(&cfg.cache.host, "cache-host", "", "Cache host")
	flag.IntVar(&cfg.cache.port, "cache-port", 0, "Cache port")
	flag.StringVar(&cfg.cache.password, "cache-dsn", "", "Cache password")
	flag.IntVar(&cfg.cache.db, "cache-db", 0, "Cache db")

	flag.StringVar(&cfg.bank.username, "bank-username", "", "Bank username")
	flag.StringVar(&cfg.bank.password, "bank-password", "", "Bank password")
	flag.BoolVar(&cfg.bank.admin, "bank-admin", false, "Bank admin status")

	flag.Parse()

	if cfg.db.dsn == "" {
		fmt.Println("You must enter a DB DSN to add a user.")
		os.Exit(1)
	}

	if cfg.cache.host == "" {
		fmt.Println("You must enter a cache host to add a user.")
		os.Exit(1)
	}

	if cfg.bank.username == "" {
		fmt.Println("You must enter a bank username to add a user.")
		os.Exit(1)
	}

	if cfg.bank.password == "" {
		fmt.Println("You must enter a bank password to add a user.")
		os.Exit(1)
	}

	if cfg.cache.port == 0 {
		fmt.Println("You must enter a cache port to add a user.")
		os.Exit(1)
	}

	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	// Open DB connection
	db, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}
	defer db.Close()

	// Open cache connection
	cache, err := openCache(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}
	defer cache.Close()

	// Create the new bank object
	bank := &data.Bank{
		Username: cfg.bank.username,
		Admin:    cfg.bank.admin,
	}

	err = bank.Password.Set(cfg.bank.password)
	if err != nil {
		logger.PrintFatal(fmt.Errorf("unable to set bank password, err: %v", err), nil)
	}

	// Validate the bank
	v := validator.New()

	if data.ValidateBank(v, bank); !v.Valid() {
		logger.PrintFatal(fmt.Errorf("unable to validate bank, err: %v", err), nil)
	}

	models := data.NewModels(db, cache)

	// Insert the bank then exit the program
	err = models.Banks.Insert(bank)
	if err != nil {
		logger.PrintFatal(fmt.Errorf("unable to insert bank into DB, err: %v", err), nil)
		return
	}

	logger.PrintInfo("bank created", nil)
}

func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("mysql", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)

	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}

	db.SetConnMaxIdleTime(duration)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func openCache(cfg config) (*redis.Client, error) {

	address := fmt.Sprintf("%v:%v", cfg.cache.host, cfg.cache.port)

	client := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: cfg.cache.password,
		DB:       cfg.cache.db,
	})

	err := client.Ping(context.Background()).Err()
	if err != nil {
		return nil, err
	}
	return client, nil
}
