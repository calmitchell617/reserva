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
	go run ./cmd/reserva -write-dsn=${POSTGRESQL_BENCHMARK_DSN} -engine=postgresql

## benchmark/mariadb: benchmark a mariadb db
.PHONY: benchmark/mariadb
benchmark/mariadb: build/reserva
	go run ./cmd/reserva -write-dsn=${MARIADB_BENCHMARK_DSN} -engine=mariadb

## benchmark/mysql: benchmark a mysql db
.PHONY: benchmark/mysql
benchmark/mysql: build/reserva
	go run ./cmd/reserva -write-dsn=${MYSQL_BENCHMARK_DSN} -engine=mysql

## benchmark/all: benchmark all dem docker dbs
.PHONY: benchmark/all
benchmark/all: build/reserva
	bash -c "go run ./cmd/reserva -write-dsn=${MYSQL_BENCHMARK_DSN} -engine=mysql & go run ./cmd/reserva -write-dsn=${POSTGRESQL_BENCHMARK_DSN} -engine=postgresql & go run ./cmd/reserva -write-dsn=${MARIADB_BENCHMARK_DSN} -engine=mariadb & wait"

# ----------------------------------------------
# postgresql
# ----------------------------------------------

## deploy/postgresql: create a postgresql docker container
.PHONY: deploy/postgresql
deploy/postgresql:
	docker rm -f postgresql || true
	docker run --name postgresql -v ./config/postgresql.conf:/etc/postgresql/postgresql.conf -e POSTGRES_PASSWORD=${POSTGRESQL_PASSWORD} --platform linux/amd64 -p 5432:5432 -d postgres:17.0-bookworm  -c 'config_file=/etc/postgresql/postgresql.conf'

## docker run --device-read-iops /dev/nvme0n1p2:1000 --device-write-iops /dev/nvme0n1p2:1000 --name postgresql -v ./config/postgresql/postgresql.conf:/etc/postgresql/postgresql.conf -e POSTGRES_PASSWORD=${POSTGRESQL_PASSWORD} --platform linux/amd64 -p 5432:5432 -d postgres:17.0-bookworm  -c 'config_file=/etc/postgresql/postgresql.conf'


## prepare/postgresql: prepare a postgresql db for benchmarking
.PHONY: prepare/postgresql
prepare/postgresql:
	time psql ${POSTGRESQL_SETUP_DSN} -f migrations/postgresql_init.sql

## prepare/optimized-postgresql: prepare a postgresql db for benchmarking
.PHONY: prepare/optimized-postgresql
prepare/optimized-postgresql:
	time psql ${POSTGRESQL_SETUP_DSN} -f migrations/postgresql_init_optimized.sql

## prepare/alloydb: prepare a postgresql db for benchmarking
.PHONY: prepare/alloydb
prepare/alloydb:
	psql ${ALLOYDB_SETUP_DSN} -f migrations/postgresql_init.sql

# ----------------------------------------------
# mariadb
# ----------------------------------------------

## deploy/mariadb: create a mariadb docker container
.PHONY: deploy/mariadb
deploy/mariadb:
	docker rm -f mariadb || true
	docker run --name mariadb -v ./config/mariadb.cnf:/etc/mysql/conf.d/mariadb.cnf -e MYSQL_ROOT_PASSWORD=${MARIADB_PASSWORD} --platform linux/amd64 -p 3306:3306 -d mariadb:11.5.2-noble

## prepare/mariadb: prepare a mariadb db for benchmarking
.PHONY: prepare/mariadb
prepare/mariadb:
	time mariadb -h ${MARIADB_HOSTNAME} -P 3306 -u root -p${MARIADB_PASSWORD} mysql < migrations/mysql_init.sql

# ----------------------------------------------
# mysql
# ----------------------------------------------

## deploy/mysql: create a mysql docker container
.PHONY: deploy/mysql
deploy/mysql:
	docker rm -f mysql || true
	docker run --name mysql -v ./config/mysql.cnf:/etc/mysql/conf.d/mysql.cnf -e MYSQL_ROOT_PASSWORD=${MYSQL_PASSWORD} --platform linux/amd64 -p 3306:3306 -d mysql:9.1.0-oraclelinux9

## prepare/mysql: prepare a mysql db for benchmarking
.PHONY: prepare/mysql
prepare/mysql:
	mysql -h ${MYSQL_HOSTNAME} -P 3306 -u root -p${MYSQL_PASSWORD} mysql < migrations/mysql_init.sql

# ALL

## prepare/all: prepare all dbs for benchmarking
.PHONY: prepare/all
prepare/all:
	time bash -c "psql ${POSTGRESQL_SETUP_DSN} -q -f migrations/postgresql_init.sql & mysql -h ${MARIADB_HOSTNAME} -P 3306 -u root -p${MARIADB_PASSWORD} mysql < migrations/mysql_init.sql & mysql -h ${MYSQL_HOSTNAME} -P 3306 -u root -p${MYSQL_PASSWORD} mysql < migrations/mysql_init.sql & wait"

## prepare/pg-and-alloy: 
.PHONY: prepare/pg-and-alloy
prepare/pg-and-alloy:
	time bash -c "psql ${POSTGRESQL_SETUP_DSN} -f migrations/postgresql_init.sql & psql ${ALLOYDB_SETUP_DSN} -f migrations/postgresql_init.sql & wait"

## benchmark/pg-and-alloy: benchmark pg and alloy
.PHONY: benchmark/pg-and-alloy
benchmark/pg-and-alloy: build/reserva
	bash -c "go run ./cmd/reserva -write-dsn=${POSTGRESQL_BENCHMARK_DSN} -read-dsn=${POSTGRESQL_READ_DSN} -engine=postgresql & go run ./cmd/reserva -write-dsn=${ALLOYDB_BENCHMARK_DSN} -read-dsn=${ALLOYDB_READ_DSN} -engine=postgresql -name=alloydb & wait"