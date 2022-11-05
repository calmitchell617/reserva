package main

import (
	"context"
	"database/sql"
	"expvar"
	"flag"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/calmitchell617/reserva/internal/data"
	"github.com/calmitchell617/reserva/internal/jsonlog"
	"github.com/calmitchell617/reserva/internal/vcs"

	_ "github.com/go-sql-driver/mysql"
)

var (
	version = vcs.Version()
)

type config struct {
	port int
	env  string
	db   struct {
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
	limiter struct {
		enabled bool
		rps     float64
		burst   int
	}

	cors struct {
		trustedOrigins []string
	}
}

// the application struct is a dependency injection technique that
// allows you to add functionality to handlers without having global
// variables all over the place
type application struct {
	config config
	logger *jsonlog.Logger
	models data.Models
	wg     sync.WaitGroup
}

func main() {
	var cfg config

	// get and parse server command line flags

	flag.IntVar(&cfg.port, "port", 80, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")

	flag.StringVar(&cfg.db.dsn, "db-dsn", "", "DB DSN")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 100, "DB max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 100, "DB max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "DB max connection idle time")

	flag.StringVar(&cfg.cache.host, "cache-host", "", "Cache host")
	flag.IntVar(&cfg.cache.port, "cache-port", 0, "Cache port")
	flag.StringVar(&cfg.cache.password, "cache-dsn", "", "Cache password")
	flag.IntVar(&cfg.cache.db, "cache-db", 0, "Cache db")

	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", false, "Enable rate limiter")
	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")

	displayVersion := flag.Bool("version", false, "Display version and exit")

	flag.Parse()

	if *displayVersion {
		fmt.Printf("Version:\t%s\n", version)
		os.Exit(0)
	}

	if cfg.db.dsn == "" {
		fmt.Println("You must enter a DB DSN to start the server.")
		os.Exit(1)
	}

	if cfg.cache.host == "" {
		fmt.Println("You must enter a cache host to start the server.")
		os.Exit(1)
	}

	if cfg.cache.port == 0 {
		fmt.Println("You must enter a cache port to start the server.")
		os.Exit(1)
	}

	// we will use a custom JSON logger
	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	// open Singlestore and Redis connections
	db, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}
	defer db.Close()

	cache, err := openCache(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}
	defer cache.Close()

	logger.PrintInfo("database connection pool established", nil)

	expvar.NewString("version").Set(version)

	// Publish some stats for telemetry / monitoring
	expvar.Publish("goroutines", expvar.Func(func() interface{} {
		return runtime.NumGoroutine()
	}))

	expvar.Publish("database", expvar.Func(func() interface{} {
		return db.Stats()
	}))

	expvar.Publish("timestamp", expvar.Func(func() interface{} {
		return time.Now().Unix()
	}))

	// initialize application struct
	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db, cache),
	}

	// start it up!
	err = app.serve()
	if err != nil {
		logger.PrintFatal(err, nil)
	}
}

func openDB(cfg config) (*sql.DB, error) {
	// opens singlestore connection
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

	// check that we can actually connect to the DB
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func openCache(cfg config) (*redis.Client, error) {
	// opens Redis connection

	address := fmt.Sprintf("%v:%v", cfg.cache.host, cfg.cache.port)

	client := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: cfg.cache.password,
		DB:       cfg.cache.db,
		PoolSize: 100,
	})

	// check that we can actually connect to Redis
	err := client.Ping(context.Background()).Err()
	if err != nil {
		return nil, err
	}
	return client, nil
}
