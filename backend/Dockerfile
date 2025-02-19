# Stage 1: Build the application
FROM golang:1.24 AS builder

WORKDIR /app

# Copy go.mod and go.sum first for caching dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire project (relative to ./backend)
COPY . .

# Build the application (matching Makefile build flags and paths)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o manager ./cmd/manager

# Stage 2: Create the final image (smaller and more secure)
FROM alpine:latest

WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/manager .

# Copy the config directory from the local context into the image
COPY --from=builder /app/config/ config/  # Correct: Relative to the BUILDER stage
COPY --from=builder /app/.env /app/.env
# Copy certs (if needed inside the container, often not if using Nginx for TLS)
# COPY --from=builder /app/certs /certs  #Probably dont need this if using nginx

# Ensure the binary is executable
RUN chmod +x /app/manager

# Expose the port the app runs on
EXPOSE 8080

# Command to run the executable
CMD ["/app/manager"]