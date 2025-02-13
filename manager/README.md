# Beehive Manager

A secure and scalable manager service for handling agent tasks and communication.

## Prerequisites

- Go 1.22 or higher
- MongoDB 4.4 or higher
- make (optional, for using Makefile commands)

## Initial Setup

1. Clone the repository:

```bash
git clone https://github.com/whit3rabbit/beehive/manager
cd manager
```

2. Build the application:

```bash
make build
```

3. Run the setup command. You have two options:

   a. Interactive setup (recommended for first-time setup):

   ```bash
   ./bin/manager setup
   ```

   b. Automatic setup with random secure values:

   ```bash
   ./bin/manager setup --skip
   ```

   The setup command will:
   - Generate SSL certificates
   - Create MongoDB user and database
   - Set up the admin user with secure password
   - Create default worker role
   - Generate secure random values for JWT secret and API keys
   - Create the .env file with all configurations

## Configuration

Configuration can be managed through:

- YAML configuration file (config/config.yaml)
- Command line flags
- Environment variables

### Command Line Flags

```bash
Usage of ./bin/manager:
  -admin-pass string
        Default admin password
  -admin-user string
        Default admin username
  -config string
        Path to configuration file (default "config/config.yaml")
  -host string
        Server host address
  -jwt-secret string
        JWT secret key
  -log-level string
        Log level (debug, info, warn, error)
  -mongo-db string
        MongoDB database name
  -mongo-uri string
        MongoDB URI
  -port int
        Server port
  -tls
        Enable TLS
  -tls-cert string
        TLS certificate file path
  -tls-key string
        TLS key file path
```

### Running the Server

After setup, start the server:

```bash
./bin/manager
```

Or with custom configuration:

```bash
./bin/manager -config /path/to/config.yaml
```

## Development

### Directory Structure

```
.
├── api/            # API handlers and routes
├── cmd/            # Application entrypoints
├── config/         # Configuration files
├── docs/           # Documentation
├── internal/       # Internal packages
├── middleware/     # HTTP middleware
├── migrations/     # Database migrations
├── models/         # Data models
└── README.md
```

### Build Commands

```bash
# Build the application
make build

# Run tests
make test

# Clean build artifacts
make clean

# Format code
make fmt

# Run linter
make lint

# Generate test coverage
make coverage

# Build for multiple platforms
make build-all
```

## API Documentation

Detailed API documentation is available in the `docs/api` directory:

- OpenAPI specification: `docs/api/openapi.yaml`
- API Guide: `docs/api/README.md`

## Security

- All passwords are hashed using bcrypt
- TLS 1.2+ with secure cipher suites
- Rate limiting on authentication endpoints
- API key and signature-based authentication for agents
- JWT-based authentication for admin routes
