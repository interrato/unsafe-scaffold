FROM golang:1.26.3-alpine3.21 AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
ENV CGO_ENABLED=0
RUN go build -trimpath -ldflags="-s -w" -o /usr/local/bin/interrato.fly.dev

FROM alpine:3.21

COPY --from=builder /usr/local/bin/interrato.fly.dev /usr/local/bin/interrato.fly.dev

ENTRYPOINT ["/usr/local/bin/interrato.fly.dev"]
