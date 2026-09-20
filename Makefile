.PHONY: dev run build test generate db-push db-migrate tidy clean

# Run the server (dev alias)
dev:
	go run ./cmd/server

# Run the server
run:
	go run ./cmd/server

# Build the executable
build:
	go build -o bin/server ./cmd/server

# Run unit tests
test:
	go test -v ./...

# Generate Prisma Client Go
generate:
	go run github.com/steebchen/prisma-client-go generate

# Push Prisma schema to PostgreSQL database (prototyping/sync)
db-push:
	go run github.com/steebchen/prisma-client-go db push

# Create and apply migrations
db-migrate:
	go run github.com/steebchen/prisma-client-go migrate dev

# Tidy Go dependencies
tidy:
	go mod tidy

# Clean build artifacts
clean:
	rm -rf bin/ server.exe
