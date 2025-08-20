.PHONY: build run docker-run docker-stop swagger deps


# Build the application
build:
	go build -o bin/server ./cmd/server

# Run the application locally
run:
	go run ./cmd/server/main.go


# Run with Docker Compose
docker-run:
	docker-compose up --build


# Stop all services
docker-stop:
	docker-compose down


# Generate Swagger docs
swagger:
	swag init --parseDependency -d internal/api -g ../../cmd/server/main.go


# Install dependencies
deps:
	go mod download
	go mod tidy



