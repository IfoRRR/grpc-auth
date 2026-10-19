.PHONY: gen run client test tidy

export PATH := $(HOME)/sdk/go/bin:$(HOME)/go/bin:$(HOME)/sdk/protoc/bin:$(PATH)
export GOPATH := $(HOME)/go

gen:
	bash scripts/gen.sh

tidy:
	go mod tidy

test:
	go test ./...

run:
	go run ./cmd/server

client:
	go run ./cmd/client
