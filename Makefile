.PHONY: help test test-cover lint build run clean mockgen

help:
	@echo "Available targets:"
	@echo "  test        - Run all tests"
	@echo "  test-cover  - Run tests with coverage report"
	@echo "  lint        - Run golangci-lint"
	@echo "  build       - Build the project"
	@echo "  run         - Run the API locally (loads .env from project root)"
	@echo "  clean       - Remove temporary files"
	@echo "  mockgen     - Regenerate mocks"

test:
	go test ./... -v

test-cover:
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out
	@echo "Coverage HTML report: coverage.html"
	go tool cover -html=coverage.out -o coverage.html

lint:
	$(shell which golangci-lint || echo "golangci-lint not found, installing...") run ./...

build:
	go build ./...

run:
	go run ./cmd/app

clean:
	rm -f coverage.out coverage.html

mockgen:
	mockgen -source=internal/core/interfaces/primary/lesson_plan_manager.go -destination=internal/api/handlers/mocks/mock_lesson_plan_manager.go -package=mocks
	mockgen -source=internal/core/interfaces/primary/user_manager.go -destination=internal/api/handlers/mocks/mock_user_manager.go -package=mocks
