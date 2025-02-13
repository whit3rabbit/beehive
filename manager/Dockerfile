# Build stage
FROM golang:1.22-alpine AS builder

# Install required packages for building
RUN apk add --no-cache gcc musl-dev make

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application
RUN make build

# Final stage
FROM alpine:latest

# Install required runtime packages
RUN apk add --no-cache ca-certificates tzdata jq mongodb-tools

# Create non-root user
RUN adduser -D -g '' appuser

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/bin/manager .

# Copy required files
COPY --from=builder /app/config/config.yaml ./config/
COPY --from=builder /app/docs ./docs

# Create necessary directories with proper permissions
RUN mkdir -p /app/certs /app/logs /app/data && \
    chown -R appuser:appuser /app

# Create startup script
RUN echo '#!/bin/sh\n\
\n\
generate_password() {\n\
    tr -dc "A-Za-z0-9!#$%&*+-=?@^_" < /dev/urandom | head -c 20\n\
}\n\
\n\
# Generate MongoDB passwords if not provided\n\
if [ -z "$MONGODB_PASSWORD" ]; then\n\
    export MONGODB_PASSWORD=$(generate_password)\n\
fi\n\
\n\
if [ -z "$MONGO_ROOT_PASSWORD" ]; then\n\
    export MONGO_ROOT_PASSWORD=$(generate_password)\n\
fi\n\
\n\
echo "Running setup..."\n\
./manager setup --skip > setup_output.txt 2>&1\n\
\n\
if [ $? -eq 0 ]; then\n\
    echo "=== IMPORTANT: SAVE THESE CREDENTIALS ==="\n\
    echo "Admin Username: $(grep "Username:" setup_output.txt | cut -d" " -f3)"\n\
    echo "Admin Password: $(grep "Password:" setup_output.txt | cut -d" " -f3)"\n\
    echo "API Key: $(grep "API Key:" setup_output.txt | cut -d" " -f3)"\n\
    echo "API Secret: $(grep "API Secret:" setup_output.txt | cut -d" " -f3)"\n\
    echo ""\n\
    echo "MongoDB Credentials:"\n\
    echo "Root Username: ${MONGO_ROOT_USERNAME:-admin}"\n\
    echo "Root Password: $MONGO_ROOT_PASSWORD"\n\
    echo "App Username: ${MONGODB_USER:-manager}"\n\
    echo "App Password: $MONGODB_PASSWORD"\n\
    echo "======================================"\n\
    \n\
    # Write credentials to .env file\n\
    cat > .env << EOL\n\
MONGO_ROOT_USERNAME=${MONGO_ROOT_USERNAME:-admin}\n\
MONGO_ROOT_PASSWORD=$MONGO_ROOT_PASSWORD\n\
MONGODB_USER=${MONGODB_USER:-manager}\n\
MONGODB_PASSWORD=$MONGODB_PASSWORD\n\
MONGODB_URI=mongodb://${MONGODB_USER:-manager}:${MONGODB_PASSWORD}@${MONGODB_HOST:-mongo}:${MONGODB_PORT:-27017}/${MONGODB_DATABASE:-manager_db}\n\
JWT_SECRET=$(grep "JWT_SECRET=" setup_output.txt | cut -d"=" -f2)\n\
API_KEY=$(grep "API Key:" setup_output.txt | cut -d" " -f3)\n\
API_SECRET=$(grep "API Secret:" setup_output.txt | cut -d" " -f3)\n\
ADMIN_DEFAULT_PASSWORD=$(grep "Password:" setup_output.txt | cut -d" " -f3)\n\
EOL\n\
\n\
    exec ./manager\n\
else\n\
    echo "Setup failed. Check setup_output.txt for details"\n\
    exit 1\n\
fi' > /app/start.sh && \
    chmod +x /app/start.sh

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1" --spider" "http://localhost:8080/health || exit 1

# Command to run
ENTRYPOINT ["/app/start.sh"]