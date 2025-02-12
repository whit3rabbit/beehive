package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"manager/internal/mongodb"
	"manager/api/handlers"
	"manager/internal/admin"
	"manager/middleware"
)

type Config struct {
	Server struct {
		Host      string `yaml:"host"`
		Port      int    `yaml:"port"`
		StaticDir string `yaml:"static_dir"`
		TLS       struct {
			Enabled  bool   `yaml:"enabled"`
			CertFile string `yaml:"cert_file"`
			KeyFile  string `yaml:"key_file"`
		} `yaml:"tls"`
	} `yaml:"server"`
	MongoDB struct {
		URI      string `yaml:"uri"`
		Database string `yaml:"database"`
	} `yaml:"mongodb"`
	Auth struct {
		JWTSecret            string `yaml:"jwt_secret"`
		TokenExpirationHours int    `yaml:"token_expiration_hours"`
		APIKey               string `yaml:"api_key"`
		APISecret            string `yaml:"api_secret"`
	} `yaml:"auth"`
	Admin struct {
		DefaultUsername string `yaml:"default_username"`
		DefaultPassword string `yaml:"default_password"`
	} `yaml:"admin"`
	Logging struct {
		Level string `yaml:"level"`
	} `yaml:"logging"`
}

func loadConfig(filename string) (*Config, error) {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found or error loading it: %v", err)
	}

	// Create config with values from environment variables
	config := &Config{}
	
	// Server config
	config.Server.Host = getEnv("SERVER_HOST", "0.0.0.0")
	config.Server.Port = getEnvAsInt("SERVER_PORT", 8080)
	config.Server.StaticDir = getEnv("STATIC_DIR", "./frontend/build")
	
	// TLS config
	config.Server.TLS.Enabled = getEnvAsBool("TLS_ENABLED", true)
	config.Server.TLS.CertFile = getEnv("TLS_CERT_FILE", "certs/server.crt")
	config.Server.TLS.KeyFile = getEnv("TLS_KEY_FILE", "certs/server.key")
	
	// MongoDB config
	config.MongoDB.URI = getEnv("MONGODB_URI", "mongodb://localhost:27017")
	config.MongoDB.Database = getEnv("MONGODB_DATABASE", "manager_db")
	
	// Auth config
	config.Auth.JWTSecret = getEnv("JWT_SECRET", "your_super_secret_jwt_key")
	config.Auth.TokenExpirationHours = getEnvAsInt("TOKEN_EXPIRATION_HOURS", 24)
	config.Auth.APIKey = getEnv("API_KEY", "expected_api_key")
	config.Auth.APISecret = getEnv("API_SECRET", "expected_api_secret")
	
	// Admin config
	config.Admin.DefaultUsername = getEnv("ADMIN_DEFAULT_USERNAME", "admin")
	config.Admin.DefaultPassword = getEnv("ADMIN_DEFAULT_PASSWORD", "changeme")
	
	// Logging config
	config.Logging.Level = getEnv("LOG_LEVEL", "info")

	return config, nil
}

// Helper functions to get environment variables
func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

func getEnvAsInt(key string, defaultVal int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultVal
}

func getEnvAsBool(key string, defaultVal bool) bool {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseBool(valueStr); err == nil {
		return value
	}
	return defaultVal
}

func main() {
	// Load configuration
	config, err := loadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Connect to MongoDB
	if err := mongodb.Connect(config.MongoDB.URI); err != nil {
		log.Fatalf("Error connecting to MongoDB: %v", err)
	}
	defer func() {
		if err := mongodb.Disconnect(); err != nil {
			log.Printf("Error disconnecting MongoDB: %v", err)
		}
	}()

	// Create a new Echo instance
	e := echo.New()

	// Global middleware: logging and recovery
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// (Optional) Use our custom middleware for API auth on protected routes.
	// e.Use(customMiddleware.APIAuthMiddleware)
	// e.Use(customMiddleware.AdminAuthMiddleware)

	// Admin login route
	e.POST("/admin/login", admin.LoginHandler)

	// Agent endpoints
	e.POST("/agent/register", handlers.RegisterAgent)
	e.POST("/agent/heartbeat", handlers.AgentHeartbeat)
	e.GET("/agent/:agent_id/tasks", handlers.ListAgentTasks)

	// Task endpoints
	e.POST("/task/create", handlers.CreateTask)
	e.GET("/task/status/:task_id", handlers.GetTaskStatus)
	e.POST("/task/cancel/:task_id", handlers.CancelTask)

	// Role endpoints
	e.GET("/roles", handlers.ListRoles)
	e.POST("/roles", handlers.CreateRole)
	e.GET("/roles/:role_id", handlers.GetRole)

	// Serve static files for React frontend (if available)
	if config.Server.StaticDir != "" {
		e.Static("/", config.Server.StaticDir)
	}

	addr := fmt.Sprintf("%s:%d", config.Server.Host, config.Server.Port)
	if config.Server.TLS.Enabled {
		log.Printf("Starting server with TLS on https://%s", addr)
		if err := e.StartTLS(addr, config.Server.TLS.CertFile, config.Server.TLS.KeyFile); err != nil {
			log.Fatalf("Error starting TLS server: %v", err)
		}
	} else {
		log.Printf("Starting server on http://%s", addr)
		if err := e.Start(addr); err != nil {
			log.Fatalf("Error starting server: %v", err)
		}
	}
}
