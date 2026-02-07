.PHONY: help build up down restart logs clean dev test update status setup
export DOTENV := .env

# Default target
help:
	@echo "What Did I Miss - Build Commands"
	@echo "================================"
	@echo ""
	@echo "Development:"
	@echo "  make dev        - Run locally without Docker"
	@echo "  make test       - Run tests"
	@echo ""
	@echo "Production:"
	@echo "  make build      - Build Docker images"
	@echo "  make up         - Start all services"
	@echo "  make down       - Stop all services"
	@echo "  make restart    - Rebuild and restart services"
	@echo "  make logs       - View logs (follow mode)"
	@echo ""
	@echo "Maintenance:"
	@echo "  make clean      - Remove containers and volumes"
	@echo "  make update     - Pull latest code and restart"

# Build Docker images
build:
	docker compose build

# Start all services
up:
	docker compose up -d

# Stop all services
down:
	docker compose down

# Full restart with rebuild
restart:
	docker compose down
	docker compose build
	docker compose up -d

# View logs
logs:
	docker compose logs -f

# Clean everything including volumes
clean:
	docker compose down -v
	docker system prune -f

# Local development (requires Go)
dev:
	go get whatdidimiss
	@set -o allexport; [ -f $(DOTENV) ] && source $(DOTENV); set +o allexport; \
	go run .

# Run tests
test:
	go test -v ./...

# Pull latest and restart (for deployment)
update:
	git pull
	docker compose build
	docker compose up -d

# Check service status
status:
	docker compose ps
	@echo ""
	@echo "Health check:"
	@curl -s http://localhost/api/health || echo "API not responding"

# Initial server setup (run once on new server)
setup:
	@echo "Setting up server..."
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo "Created .env from .env.example"; \
		echo "Please edit .env and add your GEMINI_API_KEY"; \
	fi
	@mkdir -p ssl
	@echo "Setup complete. Run 'make up' to start services."
