package main

import (
    "context"
    "fmt"
    "net/http"
    "os"
    "os/signal"
    "sync"
    "time"

    "github.com/joho/godotenv"
    "github.com/labstack/echo/v4"
    echoMiddleware "github.com/labstack/echo/v4/middleware"
    "go.mongodb.org/mongo-driver/bson"
    "go.uber.org/zap"
    "gopkg.in/yaml.v3"

    "github.com/whit3rabbit/beehive/manager/api/admin"
    "github.com/whit3rabbit/beehive/manager/api/handlers"
    "github.com/whit3rabbit/beehive/manager/internal/logger"
    "github.com/whit3rabbit/beehive/manager/internal/mongodb"
    customMiddleware "github.com/whit3rabbit/beehive/manager/middleware"
    "github.com/whit3rabbit/beehive/manager/migrations"
    "github.com/whit3rabbit/beehive/manager/models"
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

// HealthStatus defines the structure for health check response.
type HealthStatus struct {
	Status  string  `json:"status"`
	MongoDB string  `json:"mongodb"`
	Uptime  float64 `json:"uptime"`
	Version string  `json:"version"`
}

var startTime time.Time

var (
    loginMutex sync.RWMutex
    loginAttempts = make(map[string]struct {
        count       int
        lastAttempt time.Time
    })
)

// CleanupLoginAttempts periodically cleans up expired login attempts
func CleanupLoginAttempts() {
    ticker := time.NewTicker(15 * time.Minute)
    for range ticker.C {
        loginMutex.Lock()
        now := time.Now()
        for username, attempt := range loginAttempts {
            if now.Sub(attempt.lastAttempt) > 15*time.Minute {
                delete(loginAttempts, username)
            }
        }
        loginMutex.Unlock()
    }
}

func loadConfig(filename string) (*Config, error) {
	// Load .env first
	godotenv.Load()

	// Read YAML file
	data, err := os.ReadFile(filename)
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

func main() {
	// Record start time
	startTime = time.Now()

	// Load configuration
	config, err := loadConfig("config.yaml")
	if err != nil {
		logger.Fatal("Error loading config", zap.Error(err))
	}

	// Ensure required config values are set
	if config.Auth.JWTSecret == "" {
		logger.Fatal("JWT_SECRET must be set in configuration")
	}
	if config.Admin.DefaultPassword == "" {
		logger.Fatal("ADMIN_DEFAULT_PASSWORD must be set in configuration")
	}

	// Initialize Logger
	if err := logger.Initialize(config.Logging.Level); err != nil {
		logger.Fatal("Failed to initialize logger", zap.Error(err))
	}
	defer logger.Sync()

	// Connect to MongoDB (unchanged)
	if err := mongodb.Connect(config.MongoDB.URI); err != nil {
		logger.Fatal("Error connecting to MongoDB", zap.Error(err))
	}

	// Run migrations
	// Ensure MongoDB client is initialized before running migrations
	if mongodb.Client == nil {
		logger.Fatal("MongoDB client not initialized")
	}

	db := mongodb.Client.Database(config.MongoDB.Database)

	allMigrations := []migrations.Migration{
		migrations.Migration0001,
	}

	if err := migrations.RunMigrations(db, allMigrations); err != nil {
		logger.Fatal("Error running migrations", zap.Error(err))
	}

	// Ensure admin user exists
	adminCollection := mongodb.Client.Database(config.MongoDB.Database).Collection("admins")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var adminUser models.Admin
	if err := adminCollection.FindOne(ctx, bson.M{"username": config.Admin.DefaultUsername}).Decode(&adminUser); err != nil {
		hashedPassword, err := admin.GenerateHashPassword(config.Admin.DefaultPassword)
		if err != nil {
			logger.Fatal("Failed to hash default admin password", zap.Error(err), zap.String("username", config.Admin.DefaultUsername))
		}
		_, err = adminCollection.InsertOne(ctx, models.Admin{
			Username:  config.Admin.DefaultUsername,
			Password:  hashedPassword,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})
		if err != nil {
			logger.Error("Failed to create initial admin user", zap.Error(err), zap.String("username", config.Admin.DefaultUsername))
		}
	}
	defer func() {
		if err := mongodb.Disconnect(); err != nil {
			logger.Error("Error disconnecting MongoDB", zap.Error(err))
		}
	}()

	// Create a new Echo instance
	e := echo.New()

	// Start cleanup routine for login attempts
	go admin.CleanupLoginAttempts()

	// Middleware to set config values in context
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("jwt_secret", config.Auth.JWTSecret)
			c.Set("token_expiration_hours", config.Auth.TokenExpirationHours)
			c.Set("api_key", config.Auth.APIKey)
			c.Set("api_secret", config.Auth.APISecret)
			c.Set("mongodb_database", config.MongoDB.Database)
			return next(c)
		}
	})

	// Global middleware: logging, recovery, and timeout.
	e.Use(echoMiddleware.Logger())
	e.Use(echoMiddleware.Recover())
	e.Use(echoMiddleware.TimeoutWithConfig(echoMiddleware.TimeoutConfig{
		Timeout: 30 * time.Second,
	}))

	// Public routes (no auth required)
	e.POST("/admin/login", admin.LoginHandler)

	// Health check endpoint
	e.GET("/health", func(c echo.Context) error {
		// Check MongoDB connection
		mongoStatus := "healthy"
		if err := mongodb.Client.Ping(context.Background(), nil); err != nil {
			mongoStatus = "unhealthy"
		}

		// Calculate uptime
		uptime := time.Since(startTime).Seconds()

		// Get version information (replace with your actual version retrieval)
		version := "v1.0.0"

		healthStatus := HealthStatus{
			Status:  "healthy",
			MongoDB: mongoStatus,
			Uptime:  uptime,
			Version: version,
		}

		return c.JSON(http.StatusOK, healthStatus)
	})

	// Admin routes (JWT auth)
	adminRoutes := e.Group("/admin")
	adminRoutes.Use(customMiddleware.AdminAuthMiddleware)
	adminRoutes.Use(echoMiddleware.RateLimiter(echoMiddleware.NewRateLimiterMemoryStore(5))) // stricter limit for admin routes

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
	agentRoutes.POST("/task/create", handlers.CreateTask, customMiddleware.RequestValidationMiddleware)
	agentRoutes.GET("/task/status/:task_id", handlers.GetTaskStatus)
	agentRoutes.POST("/task/cancel/:task_id", handlers.CancelTask)

	// Serve static files for React frontend (if available)
	if config.Server.StaticDir != "" {
		e.Static("/", config.Server.StaticDir)
	}

	addr := fmt.Sprintf("%s:%d", config.Server.Host, config.Server.Port)
	if config.Server.TLS.Enabled {
		// Check if TLS cert and key files exist
		if _, err := os.Stat(config.Server.TLS.CertFile); os.IsNotExist(err) {
			logger.Fatal("TLS certificate file not found", zap.String("path", config.Server.TLS.CertFile))
		}
		if _, err := os.Stat(config.Server.TLS.KeyFile); os.IsNotExist(err) {
			logger.Fatal("TLS key file not found", zap.String("path", config.Server.TLS.KeyFile))
		}

		logger.Info("Starting server with TLS", zap.String("address", "https://"+addr))
		if err := e.StartTLS(addr, config.Server.TLS.CertFile, config.Server.TLS.KeyFile); err != nil {
			logger.Fatal("Error starting TLS server", zap.Error(err))
		}
	} else {
		logger.Info("Starting server", zap.String("address", "http://"+addr))
		if err := e.Start(addr); err != nil {
			logger.Fatal("Error starting server", zap.Error(err))
		}
	}

	// Add graceful shutdown handling
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := e.Shutdown(ctx); err != nil {
			logger.Error("Failed to shutdown server gracefully", zap.Error(err))
		}
	}()
}
