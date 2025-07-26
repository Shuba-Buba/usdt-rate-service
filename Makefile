.PHONY: proto

GO_PATH := $(shell go env GOPATH)
PROTOC_GEN_GO := $(GO_PATH)/bin/protoc-gen-go
PROTOC_GEN_GO_GRPC := $(GO_PATH)/bin/protoc-gen-go-grpc


.PHONY: build test docker-build run lint proto clean

SERVICE_NAME := usdt-rate-service
MAIN_PATH := ./cmd
BINARY_NAME := app

build:
	go build -o $(BINARY_NAME) $(MAIN_PATH)

docker-build:
	docker-compose build

run:
	docker-compose up

lint:
	golangci-lint run

proto:
	protoc --plugin=protoc-gen-go=$(PROTOC_GEN_GO) \
	       --plugin=protoc-gen-go-grpc=$(PROTOC_GEN_GO_GRPC) \
	       --go_out=. --go_opt=paths=source_relative \
	       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
	       proto/rate/v1/rate.proto