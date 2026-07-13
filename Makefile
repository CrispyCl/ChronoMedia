ifneq ("$(wildcard .env)","")
    include .env
    export
endif

.PHONY: docker-up docker-up-d docker-down

docker-up:
	@echo "Starting containers..."
	@docker compose -f deployments/docker-compose.yaml up

docker-up-d:
	@echo "Starting containers..."
	@docker compose -f deployments/docker-compose.yaml up -d

docker-down:
	@echo "Stopping containers..."
	@docker compose down
