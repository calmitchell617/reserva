include .envrc

.PHONY: vendor
vendor:
	go mod tidy
	go mod verify
	go mod vendor

## run/api: run the cmd/api application
.PHONY: run/api
run/api: build/api
	sudo bin/api -port 80 -write-db-dsn=${DB_DSN} -read-db-dsn=${DB_DSN} -smtp-host=${SMTP_HOST} -smtp-port=${SMTP_PORT} -smtp-username=${SMTP_USERNAME} -smtp-password=${SMTP_PASSWORD} -smtp-sender=${SMTP_SENDER}

## delve: run the server
.PHONY: delve
delve: build/delve
	sudo ~/go/bin/dlv exec ./bin/api -- -port 80 -db-dsn=${DB_DSN} -smtp-host=${SMTP_HOST} -smtp-port=${SMTP_PORT} -smtp-username=${SMTP_USERNAME} -smtp-password=${SMTP_PASSWORD} -smtp-sender=${SMTP_SENDER}

## build/api: build the cmd/api application
.PHONY: build/api
build/api:
	@echo 'Building cmd/api...'
	go build -ldflags="-s" -o=./bin/api ./cmd/api

## build/delve: build the cmd/api application with delve friendly flags
.PHONY: build/delve
build/delve:
	@echo 'Building cmd/api...'
	go build -o=./bin/api ./cmd/api