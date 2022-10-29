FROM golang:1.19-bullseye

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o /usr/local/bin/api ./cmd/api
# RUN go build -ldflags="-s" -o=./bin/api ./cmd/api

CMD ["api"]