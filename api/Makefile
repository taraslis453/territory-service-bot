.PHONY: help docker-up docker-down docker-restart seed clean build run test docker-full-up docker-full-down cloud-setup cloud-build cloud-deploy cloud-logs cloud-status cloud-all 

help: ## Display this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

docker-up: ## Start PostgreSQL container
	@echo "Starting PostgreSQL container..."
	docker compose up -d
	@echo "Waiting for PostgreSQL to be ready..."
	@sleep 5
	@echo "PostgreSQL is ready!"

docker-down: ## Stop and remove PostgreSQL container
	@echo "Stopping PostgreSQL container..."
	docker compose down

docker-restart: docker-down docker-up ## Restart PostgreSQL container

docker-clean: ## Stop containers and remove volumes (WARNING: This will delete all data!)
	@echo "Stopping containers and removing volumes..."
	docker compose down -v
	@echo "All data removed!"

docker-full-up: ## Start full stack (PostgreSQL + App) in Docker
	@echo "Starting full stack..."
	docker compose --profile full up -d --build
	@echo "Waiting for services to be ready..."
	@sleep 10
	@echo "Services are ready!"

docker-full-down: ## Stop full stack
	@echo "Stopping full stack..."
	docker compose --profile full down

docker-full-logs: ## Show logs for full stack
	docker compose --profile full logs -f

seed: ## Run database seeding script
	@echo "Running database seed..."
	go run scripts/seed.go

seed-clean: docker-clean docker-up seed ## Clean database, restart container, and seed fresh data

build: ## Build the application
	@echo "Building application..."
	go build -o bin/api cmd/main.go

run: ## Run the application
	@echo "Running application..."
	go run cmd/main.go

test: ## Run tests
	@echo "Running tests..."
	go test -v ./...

dev: docker-up seed run ## Setup development environment (docker + seed + run)

logs: ## Show PostgreSQL container logs
	docker compose logs -f postgres

psql: ## Connect to PostgreSQL using psql
	docker compose exec postgres psql -U admin -d api

inspect: ## Show overview of test data in database
	@chmod +x scripts/inspect-data.sh
	@./scripts/inspect-data.sh

# Cloud Run deployment targets

cloud-setup: ## Initialize Artifact Registry repository for Cloud Run
	@echo "Setting up Google Cloud Artifact Registry..."
	@chmod +x scripts/deploy-cloudrun.sh
	@./scripts/deploy-cloudrun.sh setup

cloud-build: ## Build and push Docker image to Artifact Registry
	@echo "Building and pushing Docker image..."
	@chmod +x scripts/deploy-cloudrun.sh
	@./scripts/deploy-cloudrun.sh build

cloud-deploy: ## Deploy to Google Cloud Run
	@echo "Deploying to Cloud Run..."
	@chmod +x scripts/deploy-cloudrun.sh
	@./scripts/deploy-cloudrun.sh deploy

cloud-logs: ## Stream logs from Cloud Run
	@chmod +x scripts/deploy-cloudrun.sh
	@./scripts/deploy-cloudrun.sh logs

cloud-status: ## Show Cloud Run service status
	@chmod +x scripts/deploy-cloudrun.sh
	@./scripts/deploy-cloudrun.sh status

cloud-all: ## Run full deployment (setup + build + deploy)
	@echo "Running full Cloud Run deployment..."
	@chmod +x scripts/deploy-cloudrun.sh
	@./scripts/deploy-cloudrun.sh all

# Database Migration targets

