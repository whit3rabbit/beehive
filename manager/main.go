package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"gopkg.in/yaml.v3"
	"io/ioutil"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware" // alias for Echo's middleware
	
	customMiddleware "github.com/whit3rabbit/beehive/manager/middleware" // alias for your custom middleware
	"github.com/whit3rabbit/beehive/manager/internal/mongodb"
	"github.com/whit3rabbit/beehive/manager/api/handlers"
	"github.com/whit3rabbit/beehive/manager/api/admin"
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
	// Load .env first
	godotenv.Load()

	// Read YAML file
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %v", err)
	}

	// Expand environment variables in YAML
	expandedData := []byte(os.ExpandEnv(string(data)))

	config := &Config{}
	if err := yaml.Unmarshal(expandedData, config); err != nil {
		return nil, fmt.Errorf("error parsing YAML: %v", err)
	}

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

	// Ensure required environment variables are set
	requiredEnvs := []string{"JWT_SECRET", "API_KEY", "API_SECRET", "ADMIN_DEFAULT_PASSWORD"}
	for _, envVar := range requiredEnvs {
		if os.Getenv(envVar) == "" {
			log.Fatalf("Environment variable %s must be set", envVar)
		}
	}

	// Connect to MongoDB (unchanged)
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

	// Global middleware: logging, recovery, and rate limiting.
	e.Use(echoMiddleware.Logger())
	e.Use(echoMiddleware.Recover())
	// Example: limit to 20 requests per minute per client
	e.Use(echoMiddleware.RateLimiter(echoMiddleware.NewRateLimiterMemoryStore(20)))

	// Public routes (no auth required)
	e.POST("/admin/login", admin.LoginHandler)

	// Admin routes (JWT auth)
	adminRoutes := e.Group("/admin")
	adminRoutes.Use(customMiddleware.AdminAuthMiddleware)
	
	// Admin protected routes
	adminRoutes.GET("/roles", handlers.ListRoles)
	adminRoutes.POST("/roles", handlers.CreateRole)
	adminRoutes.GET("/roles/:role_id", handlers.GetRole)

	// Agent routes (API key auth)
	agentRoutes := e.Group("/api")
	agentRoutes.Use(customMiddleware.APIAuthMiddleware)

	// Agent endpoints
	agentRoutes.POST("/agent/register", handlers.RegisterAgent)
	agentRoutes.POST("/agent/heartbeat", handlers.AgentHeartbeat)
	agentRoutes.GET("/agent/:agent_id/tasks", handlers.ListAgentTasks)
	agentRoutes.POST("/tasks", handlers.CreateTask)
	agentRoutes.GET("/tasks/:task_id", handlers.GetTaskStatus)
	agentRoutes.POST("/tasks/:task_id/cancel", handlers.CancelTask)

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
