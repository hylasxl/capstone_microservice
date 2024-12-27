# Define commands for Docker Compose
up:
	docker-compose up -d

down:
	docker-compose down

build:
	docker-compose build

start:
	docker-compose start

stop:
	docker-compose stop

logs:
	docker-compose logs -f

restart: down up

# Run a service
run:
	docker-compose run $(service) $(cmd)

# Help message
help:
	@echo "Available targets:"
	@echo "  up       - Start services in detached mode"
	@echo "  down     - Stop and remove services"
	@echo "  build    - Build or rebuild services"
	@echo "  start    - Start existing containers"
	@echo "  stop     - Stop running containers"
	@echo "  logs     - Show logs for services"
	@echo "  restart  - Restart all services"
	@echo "  run      - Run a one-off command (use 'make run service=<name> cmd=<command>')"
