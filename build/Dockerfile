# syntax=docker/dockerfile:1

FROM golang:1.20-alpine AS builder

WORKDIR /build
COPY . /build

RUN go build -ldflags="-s -w" -o oinit-ca cmd/oinit-ca/oinit-ca.go

FROM alpine

WORKDIR /app
COPY --from=builder /build/oinit-ca /app/oinit-ca

RUN mkdir -p /etc/oinit-ca

ENTRYPOINT /app/oinit-ca 0.0.0.0:80 /etc/oinit-ca/config.ini
