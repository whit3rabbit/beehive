package main

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gopkg.in/yaml.v2"

	"manager/internal/mongodb"
	"manager/internal/api/handlers"
	"manager/internal/admin"
	customMiddleware "manager/middleware" // our custom middleware package
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
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var config Config
	if err := yaml.Unmarshal(bytes, &config); err != nil {
		return nil, err
	}
	return &config, nil
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
