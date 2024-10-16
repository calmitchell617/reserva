include .envrc

# ----------------------------------------------
# building and running
# ----------------------------------------------

## build/reserva: build the cmd/reserva application
.PHONY: build/reserva
build/reserva:
	go build -ldflags="-s -w" -o=./bin/reserva ./cmd/reserva

## benchmark/postgresql: benchmark a postgresql db
.PHONY: benchmark/postgresql
benchmark/postgresql: build/reserva
	go run ./cmd/reserva -db-dsn=${POSTGRESQL_BENCHMARK_DSN} -engine=postgresql

## benchmark/mariadb: benchmark a mariadb db
.PHONY: benchmark/mariadb
benchmark/mariadb: build/reserva
	go run ./cmd/reserva -db-dsn=${MARIADB_BENCHMARK_DSN} -engine=mariadb

# ----------------------------------------------
# postgresql
# ----------------------------------------------

## deploy/postgresql: create a postgresql docker container
.PHONY: deploy/postgresql
deploy/postgresql:
	docker rm -f postgresql || true
	docker run --name postgresql -e POSTGRES_PASSWORD=${POSTGRESQL_PASSWORD} --platform linux/arm64 -p 5432:5432 -d postgres:17.0-bookworm

## prepare/postgresql: prepare a postgresql db for benchmarking
.PHONY: prepare/postgresql
prepare/postgresql:
	psql ${POSTGRESQL_SETUP_DSN} -f migrations/postgresql_init.sql

# ----------------------------------------------
# mariadb
# ----------------------------------------------

## deploy/mariadb: create a mariadb docker container
.PHONY: deploy/mariadb
deploy/mariadb:
	docker rm -f mariadb || true
	docker run --name mariadb -e MYSQL_ROOT_PASSWORD=${MARIADB_PASSWORD} --platform linux/arm64 -p 3306:3306 -d mariadb:11.5.2-noble

## prepare/mariadb: prepare a mariadb db for benchmarking
.PHONY: prepare/mariadb
prepare/mariadb:
	mariadb -h 127.0.0.1 -P 3306 -u root -p${MARIADB_PASSWORD} mysql < migrations/mariadb_init.sql