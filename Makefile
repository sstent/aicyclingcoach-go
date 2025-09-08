.PHONY: up down build start stop restart logs backend-logs frontend-logs db-logs

# Start all services in detached mode
up:
	docker-compose up -d

# Stop and remove all containers
down:
	docker-compose down

# Rebuild all Docker images
build:
	docker-compose build --no-cache

# Start services if not running, otherwise restart
start:
	docker-compose start || docker-compose up -d

# Stop running services
stop:
	docker-compose stop

# Restart all services
restart:
	docker-compose restart

# Show logs for all services
logs:
	docker-compose logs -f

# Show backend logs
backend-logs:
	docker-compose logs -f backend

# Show frontend logs
frontend-logs:
	docker-compose logs -f frontend

# Show database logs
db-logs:
	docker-compose logs -f db

# Initialize database and run migrations
init-db:
	docker-compose run --rm backend alembic upgrade head

# Create new database migration
migration:
	docker-compose run --rm backend alembic revision --autogenerate -m "$(m)"

# Run tests
test:
	docker-compose run --rm backend pytest

# Open database shell
db-shell:
	docker-compose exec db psql -U appuser -d cyclingdb