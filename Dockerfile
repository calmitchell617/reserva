FROM golang:1.19-bullseye

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o /usr/local/bin/api ./cmd/api
RUN go build -v -o /usr/local/bin/loadTest ./cmd/loadTest
RUN go build -v -o /usr/local/bin/addBank ./cmd/addBank

CMD tail -f /dev/null