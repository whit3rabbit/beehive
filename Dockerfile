# Use an official Go runtime as a parent image
FROM golang:1.22-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the application
RUN go build -o bin/manager main.go

# Use a minimal alpine image for the final stage
FROM alpine:latest

# Set the working directory inside the container
WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/bin/manager ./

# Expose the port the app listens on
EXPOSE 8080

# Command to run the executable
CMD ["./manager"]
