BINARY_NAME := cc-queue

.PHONY: deps
deps:
	go mod tidy

.PHONY: build
build:
	go build -o $(BINARY_NAME) .

.PHONY: install
install:
	go install .

.PHONY: test
test:
	go test -race ./...

.PHONY: cover
cover:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
