ifneq ("$(wildcard .env)","")
    include .env
    export
endif

.PHONY: docker-up docker-up-d docker-down proto-gen

docker-up:
	@echo "Starting containers..."
	@docker compose -f deployments/docker-compose.yaml up

docker-up-d:
	@echo "Starting containers..."
	@docker compose -f deployments/docker-compose.yaml up -d

docker-down:
	@echo "Stopping containers..."
	@docker compose down

proto-gen:
	@echo "Generating gRPC and Protobuf code..."
	@mkdir -p api/pb/go
	protoc --proto_path=api/proto \
	       --go_out=api/pb/go --go_opt=paths=source_relative \
	       --go-grpc_out=api/pb/go --go-grpc_opt=paths=source_relative \
	       api/proto/auth/v1/auth.proto \
	       api/proto/users/v1/users.proto
	@echo "Code generation completed successfully."