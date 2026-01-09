.PHONY: help dev build test migrate migrate-up migrate-down migrate-create sqlboiler clean docker-up docker-down

# Load environment variables from .env file if it exists
ifneq (,$(wildcard .env))
    include .env
    export
endif

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

dev: ## Run API server in development mode
	go run cmd/api/main.go

worker: ## Run worker in development mode
	go run cmd/worker/main.go

build: ## Build all binaries
	go build -o bin/api cmd/api/main.go
	go build -o bin/worker cmd/worker/main.go
	go build -o bin/seed cmd/seed/main.go

test: ## Run tests
	go test -v -race -coverprofile=coverage.out ./...

test-coverage: test ## Run tests with coverage report
	go tool cover -html=coverage.out

migrate: ## Run all pending migrations
	migrate -path migrations -database "$${DATABASE_URL}" up

migrate-up: ## Run migrations up by N steps (e.g., make migrate-up N=1)
	migrate -path migrations -database "$${DATABASE_URL}" up $(N)

migrate-down: ## Run migrations down by N steps (e.g., make migrate-down N=1)
	migrate -path migrations -database "$${DATABASE_URL}" down $(N)

migrate-create: ## Create a new migration file (e.g., make migrate-create NAME=create_users_table)
	migrate create -ext sql -dir migrations -seq $(NAME)

sqlboiler: ## Generate SQLBoiler models
	sqlboiler psql --wipe

seed: ## Seed database with test data
	go run cmd/seed/main.go

docker-up: ## Start docker-compose services
	docker-compose up -d

docker-down: ## Stop docker-compose services
	docker-compose down

docker-logs: ## Show docker-compose logs
	docker-compose logs -f

clean: ## Clean build artifacts
	rm -rf bin/
	rm -f coverage.out

install-tools: ## Install development tools
	go install github.com/volatiletech/sqlboiler/v4@latest
	go install github.com/volatiletech/sqlboiler/v4/drivers/sqlboiler-psql@latest
	brew install golang-migrate
