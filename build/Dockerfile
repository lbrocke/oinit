# syntax=docker/dockerfile:1

FROM golang:1.20-alpine AS builder

RUN mkdir /build

WORKDIR /build
COPY . /build

RUN go mod download
RUN go build -ldflags="-s -w" -o ./oinit-ca cmd/oinit-ca/oinit-ca.go

FROM alpine

RUN mkdir /app
RUN mkdir -p /etc/oinit-ca

WORKDIR /app
COPY --from=builder /build/oinit-ca /app/oinit-ca

ENTRYPOINT /app/oinit-ca 0.0.0.0:80 /etc/oinit-ca/config.ini
