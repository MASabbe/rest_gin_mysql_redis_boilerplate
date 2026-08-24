.PHONY: all run build test test-cover test-race fmt vet lint docker-up docker-down docker-logs clean

# Default binary name and directory
BINARY_NAME=bin/api
MAIN_SRC=./cmd/api

all: fmt vet test build

## run: Run the application locally
run:
	go run $(MAIN_SRC)

## build: Compile the production binary
build:
	go build -ldflags="-w -s" -o $(BINARY_NAME) $(MAIN_SRC)

## test: Run unit tests
test:
	go test -v ./...

## test-cover: Run unit tests with coverage summary
test-cover:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

## test-race: Run unit tests with race condition detector
test-race:
	go test -v -race ./...

## fmt: Format Go source code
fmt:
	go fmt ./...

## vet: Run Go static analysis
vet:
	go vet ./...

## lint: Run linter if installed
lint:
	golangci-lint run ./...

## docker-up: Start full stack in Docker Compose
docker-up:
	docker-compose up -d --build

## docker-down: Stop all containers and remove volumes
docker-down:
	docker-compose down -v

## docker-logs: Follow logs of all containers
docker-logs:
	docker-compose logs -f

## feature: Scaffold a new canonical feature module (e.g. make feature NAME=product)
feature:
	go run ./cmd/scaffold -name=$(NAME)

## clean: Remove build artifacts and coverage files
clean:
	rm -rf bin coverage.out

