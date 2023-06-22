BIN_CLI=oinit
BIN_CA=oinit-ca
BIN_SWITCH=oinit-switch

.PHONY: all oinit oinit-ca swagger oinit-switch clean

all: oinit oinit-ca oinit-switch

oinit:
	go build -o ./bin/${BIN_CLI} cmd/oinit/oinit.go

oinit-ca:
	go build -o ./bin/${BIN_CA} cmd/oinit-ca/oinit-ca.go

oinit-ca-docker:
	docker build -f build/Dockerfile -t oinit-ca .

swagger:
	swag init --parseInternal -g cmd/oinit-ca/oinit-ca.go -o api/docs/
	swag fmt -d internal/api/

oinit-switch:
	go build -o ./bin/${BIN_SWITCH} cmd/oinit-switch/oinit-switch.go

clean:
	rm -rf ./bin
