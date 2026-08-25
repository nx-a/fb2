APP_NAME := fb2-reader
BUILD_DIR := bin

.PHONY: all build test vet fmt check clean run

all: check build

build:
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(APP_NAME) ./cmd/fb2

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -w cmd internal

check: fmt test vet

run:
	go run ./cmd/fb2

clean:
	rm -rf $(BUILD_DIR)
