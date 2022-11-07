include .envrc

.PHONY: vendor
vendor:
	go mod tidy
	go mod verify
	go mod vendor

## run/api: run the cmd/api application
.PHONY: run/api
run/api: build/api
	sudo bin/api -port 80 -db-dsn=${DB_DSN} -cache-host=${CACHE_HOST} -cache-port=6379

## run/addBank: run the cmd/addBank application
.PHONY: run/addBank
run/addBank: build/addBank
	sudo bin/addBank -db-dsn=${DB_DSN} -cache-host=${CACHE_HOST} -cache-port=6379 -bank-username="adminBank" -bank-password="Mypass123" -bank-admin=true

## run/loadTest: run the cmd/loadTest application
.PHONY: run/loadTest
run/loadTest: build/loadTest
	sudo bin/loadTest -host="http://localhost"

## delve: run the server
.PHONY: delve
delve: build/delve
	sudo ~/go/bin/dlv exec ./bin/api -- -port 80 -db-dsn=${DB_DSN} -smtp-host=${SMTP_HOST} -smtp-port=${SMTP_PORT} -smtp-username=${SMTP_USERNAME} -smtp-password=${SMTP_PASSWORD} -smtp-sender=${SMTP_SENDER}

## build/api: build the cmd/api application
.PHONY: build/api
build/api:
	@echo 'Building cmd/api...'
	go build -ldflags="-s" -o=./bin/api ./cmd/api

## build/addBank: build the cmd/addBank application
.PHONY: build/addBank
build/addBank:
	@echo 'Building cmd/addBank...'
	go build -ldflags="-s" -o=./bin/addBank ./cmd/addBank

## build/loadTest: build the cmd/loadTest application
.PHONY: build/loadTest
build/loadTest:
	@echo 'Building cmd/loadTest...'
	go build -ldflags="-s" -o=./bin/loadTest ./cmd/loadTest

## build/delve: build the cmd/api application with delve friendly flags
.PHONY: build/delve
build/delve:
	@echo 'Building cmd/api...'
	go build -o=./bin/api ./cmd/api