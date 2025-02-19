# Stage 1: Build the application
FROM golang:1.24 AS builder

WORKDIR /app

# Copy go.mod and go.sum first for caching dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire project
COPY . .

# Build the application (matching Makefile build flags and paths)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o manager ./cmd/manager
RUN ls -l /app  # Debug step

# Stage 2: Create the final image (smaller and more secure)
FROM alpine:latest

WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/manager .

# Copy the config directory from the local context into the image
COPY config/ config/
COPY .env /app/.env

# Ensure the binary is executable
RUN chmod +x /app/manager && ls -l /app

# Expose the port the app runs on
EXPOSE 8080

# Command to run the executable
CMD ["/app/manager"]
