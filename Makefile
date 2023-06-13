OINIT_OUT=oinit
OINITCA_OUT=oinit-ca

.PHONY: all oinit oinit-ca swag clean

all: oinit oinit-ca

oinit:
	go build -o ./bin/${OINIT_OUT} cmd/oinit/oinit.go

oinit-ca:
	go build -o ./bin/${OINITCA_OUT} cmd/oinit-ca/oinit-ca.go

swagger:
	swag init --parseInternal -g cmd/oinit-ca/oinit-ca.go -o api/docs/
	swag fmt -d internal/api/

clean:
	rm -rf ./bin
