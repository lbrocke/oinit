OUT=./bin

.PHONY: all oinit oinit-ca oinit-shell oinit-switch oinit-ca-docker swagger clean all-checks fmt fmt-check test vet staticcheck

all: oinit oinit-ca oinit-shell oinit-switch

oinit:
	go build -ldflags="-s -w" -o ${OUT}/oinit cmd/oinit/oinit.go

oinit-ca:
	go build -ldflags="-s -w" -o ${OUT}/oinit-ca cmd/oinit-ca/oinit-ca.go

oinit-shell:
	go build -ldflags="-s -w" -o ${OUT}/oinit-shell cmd/oinit-shell/oinit-shell.go

oinit-switch:
	go build -ldflags="-s -w" -o ${OUT}/oinit-switch cmd/oinit-switch/oinit-switch.go

oinit-ca-docker:
	docker build -f build/Dockerfile -t oinit-ca .

swagger:
	swag init --parseInternal -g cmd/oinit-ca/oinit-ca.go -o api/docs/
	swag fmt -d internal/api/

clean:
	rm -rf ./bin

all-checks: fmt test vet staticcheck

fmt:
	go fmt ./...

fmt-check:
	test -z $$(gofmt -l .)

test:
	go test -v ./...

vet:
	go vet ./...

staticcheck:
	staticcheck ./...
