BUILD_CLI=go build -ldflags="-s -w" -o ./bin/${BIN} cmd/oinit/oinit.go
BUILD_CA=go build -ldflags="-s -w" -o ./bin/${BIN} cmd/oinit-ca/oinit-ca.go
BUILD_SHELL=go build -ldflags="-s -w" -o ./bin/${BIN} cmd/oinit-shell/oinit-shell.go
BUILD_SWITCH=go build -ldflags="-s -w" -o ./bin/${BIN} cmd/oinit-switch/oinit-switch.go

.PHONY: all oinit oinit-ca oinit-shell oinit-switch oinit-ca-docker swagger clean

all: oinit oinit-ca oinit-switch

oinit:
	$(BUILD_CLI)

oinit-ca:
	$(BUILD_CA)

oinit-shell:
	$(BUILD_SHELL)

oinit-switch:
	$(BUILD_SWITCH)

oinit-ca-docker:
	docker build -f build/Dockerfile -t oinit-ca .

swagger:
	swag init --parseInternal -g cmd/oinit-ca/oinit-ca.go -o api/docs/
	swag fmt -d internal/api/

clean:
	rm -rf ./bin

# Cross-platform compilation

.PHONY: all-cross all-oinit

all-cross: all-oinit all-oinit-ca all-oinit-shell all-oinit-switch

all-oinit: oinit-linux-amd64 oinit-linux-arm64 oinit-darwin-amd64 oinit-darwin-arm64 oinit-freebsd-amd64 oinit-freebsd-arm64
.PHONY: oinit-linux-amd64 oinit-linux-arm64 oinit-darwin-amd64 oinit-darwin-arm64 oinit-freebsd-amd64 oinit-freebsd-arm64

oinit-linux-amd64: export GOOS=linux
oinit-linux-amd64: export GOARCH=amd64
oinit-linux-amd64: export BIN=oinit-linux-amd64
oinit-linux-amd64:
	$(BUILD_CLI)

oinit-linux-arm64: export GOOS=linux
oinit-linux-arm64: export GOARCH=arm64
oinit-linux-arm64: export BIN=oinit-linux-arm64
oinit-linux-arm64:
	$(BUILD_CLI)

oinit-darwin-amd64: export GOOS=darwin
oinit-darwin-amd64: export GOARCH=amd64
oinit-darwin-amd64: export BIN=oinit-darwin-amd64
oinit-darwin-amd64:
	$(BUILD_CLI)

oinit-darwin-arm64: export GOOS=darwin
oinit-darwin-arm64: export GOARCH=arm64
oinit-darwin-arm64: export BIN=oinit-darwin-arm64
oinit-darwin-arm64:
	$(BUILD_CLI)

oinit-freebsd-amd64: export GOOS=freebsd
oinit-freebsd-amd64: export GOARCH=amd64
oinit-freebsd-amd64: export BIN=oinit-freebsd-amd64
oinit-freebsd-amd64:
	$(BUILD_CLI)

oinit-freebsd-arm64: export GOOS=freebsd
oinit-freebsd-arm64: export GOARCH=arm64
oinit-freebsd-arm64: export BIN=oinit-freebsd-arm64
oinit-freebsd-arm64:
	$(BUILD_CLI)

all-oinit-ca: oinit-ca-linux-amd64 oinit-ca-linux-arm64 oinit-ca-darwin-amd64 oinit-ca-darwin-arm64 oinit-ca-freebsd-amd64 oinit-ca-freebsd-arm64
.PHONY: oinit-ca-linux-amd64 oinit-ca-linux-arm64 oinit-ca-darwin-amd64 oinit-ca-darwin-arm64 oinit-ca-freebsd-amd64 oinit-ca-freebsd-arm64

oinit-ca-linux-amd64: export GOOS=linux
oinit-ca-linux-amd64: export GOARCH=amd64
oinit-ca-linux-amd64: export BIN=oinit-ca-linux-amd64
oinit-ca-linux-amd64:
	$(BUILD_CA)

oinit-ca-linux-arm64: export GOOS=linux
oinit-ca-linux-arm64: export GOARCH=arm64
oinit-ca-linux-arm64: export BIN=oinit-ca-linux-arm64
oinit-ca-linux-arm64:
	$(BUILD_CA)

oinit-ca-darwin-amd64: export GOOS=darwin
oinit-ca-darwin-amd64: export GOARCH=amd64
oinit-ca-darwin-amd64: export BIN=oinit-ca-darwin-amd64
oinit-ca-darwin-amd64:
	$(BUILD_CA)

oinit-ca-darwin-arm64: export GOOS=darwin
oinit-ca-darwin-arm64: export GOARCH=arm64
oinit-ca-darwin-arm64: export BIN=oinit-ca-darwin-arm64
oinit-ca-darwin-arm64:
	$(BUILD_CA)

oinit-ca-freebsd-amd64: export GOOS=freebsd
oinit-ca-freebsd-amd64: export GOARCH=amd64
oinit-ca-freebsd-amd64: export BIN=oinit-ca-freebsd-amd64
oinit-ca-freebsd-amd64:
	$(BUILD_CA)

oinit-ca-freebsd-arm64: export GOOS=freebsd
oinit-ca-freebsd-arm64: export GOARCH=arm64
oinit-ca-freebsd-arm64: export BIN=oinit-ca-freebsd-arm64
oinit-ca-freebsd-arm64:
	$(BUILD_CA)

all-oinit-shell: oinit-shell-linux-amd64 oinit-shell-linux-arm64 oinit-shell-darwin-amd64 oinit-shell-darwin-arm64 oinit-shell-freebsd-amd64 oinit-shell-freebsd-arm64
.PHONY: oinit-shell-linux-amd64 oinit-shell-linux-arm64 oinit-shell-darwin-amd64 oinit-shell-darwin-arm64 oinit-shell-freebsd-amd64 oinit-shell-freebsd-arm64

oinit-shell-linux-amd64: export GOOS=linux
oinit-shell-linux-amd64: export GOARCH=amd64
oinit-shell-linux-amd64: export BIN=oinit-shell-linux-amd64
oinit-shell-linux-amd64:
	$(BUILD_SHELL)

oinit-shell-linux-arm64: export GOOS=linux
oinit-shell-linux-arm64: export GOARCH=arm64
oinit-shell-linux-arm64: export BIN=oinit-shell-linux-arm64
oinit-shell-linux-arm64:
	$(BUILD_SHELL)

oinit-shell-darwin-amd64: export GOOS=darwin
oinit-shell-darwin-amd64: export GOARCH=amd64
oinit-shell-darwin-amd64: export BIN=oinit-shell-darwin-amd64
oinit-shell-darwin-amd64:
	$(BUILD_SHELL)

oinit-shell-darwin-arm64: export GOOS=darwin
oinit-shell-darwin-arm64: export GOARCH=arm64
oinit-shell-darwin-arm64: export BIN=oinit-shell-darwin-arm64
oinit-shell-darwin-arm64:
	$(BUILD_SHELL)

oinit-shell-freebsd-amd64: export GOOS=freebsd
oinit-shell-freebsd-amd64: export GOARCH=amd64
oinit-shell-freebsd-amd64: export BIN=oinit-shell-freebsd-amd64
oinit-shell-freebsd-amd64:
	$(BUILD_SHELL)

oinit-shell-freebsd-arm64: export GOOS=freebsd
oinit-shell-freebsd-arm64: export GOARCH=arm64
oinit-shell-freebsd-arm64: export BIN=oinit-shell-freebsd-arm64
oinit-shell-freebsd-arm64:
	$(BUILD_SHELL)

all-oinit-switch: oinit-switch-linux-amd64 oinit-switch-linux-arm64 oinit-switch-darwin-amd64 oinit-switch-darwin-arm64 oinit-switch-freebsd-amd64 oinit-switch-freebsd-arm64
.PHONY: oinit-switch-linux-amd64 oinit-switch-linux-arm64 oinit-switch-darwin-amd64 oinit-switch-darwin-arm64 oinit-switch-freebsd-amd64 oinit-switch-freebsd-arm64

oinit-switch-linux-amd64: export GOOS=linux
oinit-switch-linux-amd64: export GOARCH=amd64
oinit-switch-linux-amd64: export BIN=oinit-switch-linux-amd64
oinit-switch-linux-amd64:
	$(BUILD_SWITCH)

oinit-switch-linux-arm64: export GOOS=linux
oinit-switch-linux-arm64: export GOARCH=arm64
oinit-switch-linux-arm64: export BIN=oinit-switch-linux-arm64
oinit-switch-linux-arm64:
	$(BUILD_SWITCH)

oinit-switch-darwin-amd64: export GOOS=darwin
oinit-switch-darwin-amd64: export GOARCH=amd64
oinit-switch-darwin-amd64: export BIN=oinit-switch-darwin-amd64
oinit-switch-darwin-amd64:
	$(BUILD_SWITCH)

oinit-switch-darwin-arm64: export GOOS=darwin
oinit-switch-darwin-arm64: export GOARCH=arm64
oinit-switch-darwin-arm64: export BIN=oinit-switch-darwin-arm64
oinit-switch-darwin-arm64:
	$(BUILD_SWITCH)

oinit-switch-freebsd-amd64: export GOOS=freebsd
oinit-switch-freebsd-amd64: export GOARCH=amd64
oinit-switch-freebsd-amd64: export BIN=oinit-switch-freebsd-amd64
oinit-switch-freebsd-amd64:
	$(BUILD_SWITCH)

oinit-switch-freebsd-arm64: export GOOS=freebsd
oinit-switch-freebsd-arm64: export GOARCH=arm64
oinit-switch-freebsd-arm64: export BIN=oinit-switch-freebsd-arm64
oinit-switch-freebsd-arm64:
	$(BUILD_SWITCH)
