.PHONY: build run test clean

# Build the application
build:
	go build -o bin/manager ./cmd/manager

# Run the application
run: build
	./bin/manager

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Download dependencies
deps:
	go mod download

# Tidy up dependencies
tidy:
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Run linter
lint:
	go vet ./...

# Generate test coverage
coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# Build for multiple platforms
build-all:
	GOOS=linux GOARCH=amd64 go build -o bin/manager-linux-amd64 ./cmd/manager
	GOOS=windows GOARCH=amd64 go build -o bin/manager-windows-amd64.exe ./cmd/manager
	GOOS=darwin GOARCH=amd64 go build -o bin/manager-darwin-amd64 ./cmd/manager

# Default target
all: clean build test
