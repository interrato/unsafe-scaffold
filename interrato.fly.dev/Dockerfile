FROM golang:1.26.3-alpine3.23 AS builder

COPY . src
WORKDIR src

ENV CGO_ENABLED=0
RUN go build -trimpath -o /usr/local/bin/interrato.fly.dev

ENTRYPOINT ["/usr/local/bin/interrato.fly.dev"]
