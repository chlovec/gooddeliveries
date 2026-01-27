# syntax=docker/dockerfile:1
FROM golang:1.25-alpine@sha256:aee43c3ccbf24fdffb7295693b6e33b21e01baec1b2a55acc351fde345e9ec34
WORKDIR /app

COPY go.mod .
RUN go mod download

COPY . .
RUN go build -o challenge .

ENTRYPOINT ["/app/challenge"]
