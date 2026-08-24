.PHONY: all run build test test-unit test-integration test-race test-cover fmt vet lint security verify docker-up docker-down docker-logs feature clean

# Default binary name and directory
BINARY_NAME=bin/api
MAIN_SRC=./cmd/api

all: verify

## run: Run the application locally
run:
	go run $(MAIN_SRC)

## build: Compile the production binary
build:
	go build -ldflags="-w -s" -o $(BINARY_NAME) $(MAIN_SRC)

## test: Run all tests
test:
	go test -v ./...

## test-unit: Run unit tests (short mode)
test-unit:
	go test -v -short ./...

## test-integration: Run module integration tests
test-integration:
	go test -v ./internal/modules/...

## test-race: Run tests with race condition detector
test-race:
	go test -v -race ./...

## test-cover: Run tests with coverage summary and HTML report
test-cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	go tool cover -func=coverage.out

## fmt: Format Go source code
fmt:
	go fmt ./...

## vet: Run Go static analysis
vet:
	go vet ./...

## lint: Run golangci-lint (falls back to go vet if not installed)
lint:
	@which golangci-lint > /dev/null 2>&1 && golangci-lint run ./... || go vet ./...

## security: Run automated secret scanning and gosec security analysis
security:
	go run ./scripts/security_check.go
	go run github.com/securego/gosec/v2/cmd/gosec@latest -quiet ./...

## verify: Run complete quality gate (fmt, vet, security, race detection, coverage, build)
verify: fmt vet security test-race test-cover build
	@echo "========================================="
	@echo "  All Quality Gates Passed Successfully! "
	@echo "========================================="

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
	rm -rf bin coverage.out coverage.html
