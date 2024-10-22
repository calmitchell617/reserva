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

## benchmark/mysql: benchmark a mysql db
.PHONY: benchmark/mysql
benchmark/mysql: build/reserva
	go run ./cmd/reserva -db-dsn=${MYSQL_BENCHMARK_DSN} -engine=mysql

## benchmark/all: benchmark all dem docker dbs
.PHONY: benchmark/all
benchmark/all: build/reserva
	bash -c "go run ./cmd/reserva -db-dsn=${MYSQL_BENCHMARK_DSN} -engine=mysql & go run ./cmd/reserva -db-dsn=${POSTGRESQL_BENCHMARK_DSN} -engine=postgresql & go run ./cmd/reserva -db-dsn=${MARIADB_BENCHMARK_DSN} -engine=mariadb & wait"

# ----------------------------------------------
# postgresql
# ----------------------------------------------

## deploy/postgresql: create a postgresql docker container
.PHONY: deploy/postgresql
deploy/postgresql:
	docker rm -f postgresql || true
	docker run --name postgresql -v ./config/postgresql/postgresql.conf:/etc/postgresql/postgresql.conf -e POSTGRES_PASSWORD=${POSTGRESQL_PASSWORD} --platform linux/amd64 -p 5432:5432 -d postgres:17.0-bookworm  -c 'config_file=/etc/postgresql/postgresql.conf'

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
	docker run --name mariadb -v ./config/mariadb/1.cnf:/etc/mysql/conf.d/1.cnf -e MYSQL_ROOT_PASSWORD=${MARIADB_PASSWORD} --platform linux/amd64 -p 3306:3306 -d mariadb:11.5.2-noble

## prepare/mariadb: prepare a mariadb db for benchmarking
.PHONY: prepare/mariadb
prepare/mariadb:
	mysql -h ${MARIADB_HOSTNAME} -P 3306 -u root -p${MARIADB_PASSWORD} mysql < migrations/mysql_init.sql

# ----------------------------------------------
# mysql
# ----------------------------------------------

## deploy/mysql: create a mysql docker container
.PHONY: deploy/mysql
deploy/mysql:
	docker rm -f mysql || true
	docker run --name mysql -v ./config/mysql/my.cnf:/etc/mysql/conf.d/my.cnf -e MYSQL_ROOT_PASSWORD=${MYSQL_PASSWORD} --platform linux/amd64 -p 3306:3306 -d mysql:9.1.0-oraclelinux9

## prepare/mysql: prepare a mysql db for benchmarking
.PHONY: prepare/mysql
prepare/mysql:
	mysql -h ${MYSQL_HOSTNAME} -P 3306 -u root -p${MYSQL_PASSWORD} mysql < migrations/mysql_init.sql

# ALL

## prepare/all: prepare all dbs for benchmarking
.PHONY: prepare/all
prepare/all:
	time bash -c "psql ${POSTGRESQL_SETUP_DSN} -q -f migrations/postgresql_init.sql & mysql -h ${MARIADB_HOSTNAME} -P 3306 -u root -p${MARIADB_PASSWORD} mysql < migrations/mysql_init.sql & mysql -h ${MYSQL_HOSTNAME} -P 3306 -u root -p${MYSQL_PASSWORD} mysql < migrations/mysql_init.sql & wait"
