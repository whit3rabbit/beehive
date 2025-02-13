package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"

	"github.com/whit3rabbit/beehive/manager/api/admin"
	"github.com/whit3rabbit/beehive/manager/api/handlers"
	"github.com/whit3rabbit/beehive/manager/internal/logger"
	"github.com/whit3rabbit/beehive/manager/internal/mongodb"
	"github.com/whit3rabbit/beehive/manager/middleware"
	customMiddleware "github.com/whit3rabbit/beehive/manager/middleware"
	"github.com/whit3rabbit/beehive/manager/migrations"
	"github.com/whit3rabbit/beehive/manager/models"
)

// RateLimiterConfig holds rate limiting configuration
type RateLimiterConfig struct {
	MaxAttempts     int           `yaml:"max_attempts"`
	WindowSeconds   int           `yaml:"window_seconds"`
	BlockoutMinutes int           `yaml:"blockout_minutes"`
}

type Config struct {
	Server struct {
		Host      string `yaml:"host"`
		Port      int    `yaml:"port"`
		StaticDir string `yaml:"static_dir"`
		TLS       struct {
			Enabled      bool     `yaml:"enabled"`
			CertFile     string   `yaml:"cert_file"`
			KeyFile      string   `yaml:"key_file"`
			MinVersion   string   `yaml:"min_version"`
			CipherSuites []string `yaml:"cipher_suites"`
		} `yaml:"tls"`
	} `yaml:"server"`
	Security struct {
		PasswordPolicy struct {
			MinLength        int  `yaml:"min_length"`
			RequireUppercase bool `yaml:"require_uppercase"`
			RequireLowercase bool `yaml:"require_lowercase"`
			RequireNumbers   bool `yaml:"require_numbers"`
			RequireSpecial   bool `yaml:"require_special"`
		} `yaml:"password_policy"`
		RateLimiting RateLimiterConfig `yaml:"rate_limiting"`
	} `yaml:"security"`
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

	// Load configuration
	config, err := loadConfig("config/config.yaml")
	if err != nil {
		logger.Fatal("Error loading config", zap.Error(err))
	}

	// Initialize Logger
	if err := logger.Initialize(config.Logging.Level); err != nil {
		logger.Fatal("Failed to initialize logger", zap.Error(err))
	}
	defer logger.Sync()

	// Connect to MongoDB
	if err := mongodb.Connect(config.MongoDB.URI); err != nil {
		logger.Fatal("Error connecting to MongoDB", zap.Error(err))
	}

	// Create root context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start cleanup routine for login attempts
	go admin.CleanupLoginAttempts(ctx)

	// Run migrations
	db := mongodb.Client.Database(config.MongoDB.Database)
	allMigrations := []migrations.Migration{
		migrations.Migration0001,
	}

	if err := migrations.RunMigrations(db, allMigrations); err != nil {
		logger.Fatal("Error running migrations", zap.Error(err))
	}

	// Ensure admin user exists
	adminCollection := db.Collection("admins")
	ctxTimeout, cancelTimeout := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelTimeout()

	passwordPolicy := models.PasswordPolicy{
		MinLength:        config.Security.PasswordPolicy.MinLength,
		RequireUppercase: config.Security.PasswordPolicy.RequireUppercase,
		RequireLowercase: config.Security.PasswordPolicy.RequireLowercase,
		RequireNumbers:   config.Security.PasswordPolicy.RequireNumbers,
		RequireSpecial:   config.Security.PasswordPolicy.RequireSpecial,
	}

	var adminUser models.Admin
	err = adminCollection.FindOne(ctxTimeout, bson.M{"username": config.Admin.DefaultUsername}).Decode(&adminUser)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			hashedPassword, err := admin.GenerateHashPassword(config.Admin.DefaultPassword, passwordPolicy)
			if err != nil {
				logger.Fatal("Failed to hash default admin password", zap.Error(err))
			}

			_, err = adminCollection.InsertOne(ctxTimeout, models.Admin{
				Username:  config.Admin.DefaultUsername,
				Password:  hashedPassword,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			})
			if err != nil {
				logger.Fatal("Failed to create initial admin user", zap.Error(err))
			}
		} else {
			logger.Error("Error checking for admin user", zap.Error(err))
		}
	}

	// Create Echo instance and set up middleware
	e := echo.New()

	// Initialize rate limiter
	rateLimiter := middleware.NewRateLimiter(
		config.Security.RateLimiting.MaxAttempts,
		time.Duration(config.Security.RateLimiting.WindowSeconds)*time.Second,
		time.Duration(config.Security.RateLimiting.BlockoutMinutes)*time.Minute,
	)

	// Admin routes (JWT auth)
	adminRoutes := e.Group("/admin")
	adminRoutes.Use(customMiddleware.AdminAuthMiddleware(rateLimiter))

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
	agentRoutes.GET("/agent/:uuid/summary", handlers.GetAgentSummary)
	agentRoutes.GET("/agent/:agent_id/tasks", handlers.ListAgentTasks)
	agentRoutes.POST("/task/create", handlers.CreateTask, customMiddleware.RequestValidationMiddleware)
	agentRoutes.GET("/task/status/:task_id", handlers.GetTaskStatus)
	agentRoutes.POST("/task/cancel/:task_id", handlers.CancelTask)

	// Serve static files for React frontend (if available)
	if config.Server.StaticDir != "" {
		e.Static("/", config.Server.StaticDir)
	}

	addr := fmt.Sprintf("%s:%d", config.Server.Host, config.Server.Port)
	server := &http.Server{
		Addr:    addr,
		Handler: e,
	}

	if config.Server.TLS.Enabled {
		// Check if TLS cert and key files exist
		if _, err := os.Stat(config.Server.TLS.CertFile); os.IsNotExist(err) {
			logger.Fatal("TLS certificate file not found", zap.String("path", config.Server.TLS.CertFile))
		}
		if _, err := os.Stat(config.Server.TLS.KeyFile); os.IsNotExist(err) {
			logger.Fatal("TLS key file not found", zap.String("path", config.Server.TLS.KeyFile))
		}

		// Configure TLS settings
		if err := configureTLS(server, config); err != nil {
			logger.Fatal("Error configuring TLS", zap.Error(err))
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

// configureTLS configures the TLS settings for the server.
func configureTLS(server *http.Server, config *Config) error {
	minVersion, err := parseTLSVersion(config.Server.TLS.MinVersion)
	if err != nil {
		return fmt.Errorf("invalid TLS version: %w", err)
	}

	cipherSuites, err := parseCipherSuites(config.Server.TLS.CipherSuites)
	if err != nil {
		return fmt.Errorf("invalid cipher suites: %w", err)
	}

	server.TLSConfig = &tls.Config{
		MinVersion:               minVersion,
		CipherSuites:            cipherSuites,
		PreferServerCipherSuites: true,
		CurvePreferences: []tls.CurveID{
			tls.X25519,
			tls.CurveP256,
		},
	}

	return nil
}

// parseTLSVersion parses the TLS version string and returns the corresponding uint16 value.
func parseTLSVersion(version string) (uint16, error) {
	switch version {
	case "1.0":
		return tls.VersionTLS10, nil
	case "1.1":
		return tls.VersionTLS11, nil
	case "1.2":
		return tls.VersionTLS12, nil
	case "1.3":
		return tls.VersionTLS13, nil
	default:
		return 0, fmt.Errorf("unsupported TLS version: %s", version)
	}
}

// parseCipherSuites parses the cipher suite strings and returns the corresponding uint16 values.
func parseCipherSuites(cipherSuites []string) ([]uint16, error) {
	var suites []uint16
	for _, suite := range cipherSuites {
		switch strings.TrimSpace(suite) {
		case "TLS_RSA_WITH_RC4_128_SHA":
			suites = append(suites, tls.TLS_RSA_WITH_RC4_128_SHA)
		case "TLS_RSA_WITH_3DES_EDE_CBC_SHA":
			suites = append(suites, tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA)
		case "TLS_RSA_WITH_AES_128_CBC_SHA":
			suites = append(suites, tls.TLS_RSA_WITH_AES_128_CBC_SHA)
		case "TLS_RSA_WITH_AES_256_CBC_SHA":
			suites = append(suites, tls.TLS_RSA_WITH_AES_256_CBC_SHA)
		case "TLS_RSA_WITH_AES_128_GCM_SHA256":
			suites = append(suites, tls.TLS_RSA_WITH_AES_128_GCM_SHA256)
		case "TLS_RSA_WITH_AES_256_GCM_SHA384":
			suites = append(suites, tls.TLS_RSA_WITH_AES_256_GCM_SHA384)
		case "TLS_ECDHE_ECDSA_WITH_RC4_128_SHA":
			suites = append(suites, tls.TLS_ECDHE_ECDSA_WITH_RC4_128_SHA)
		case "TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA":
			suites = append(suites, tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA)
		case "TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA":
			suites = append(suites, tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA)
		case "TLS_ECDHE_RSA_WITH_RC4_128_SHA":
			suites = append(suites, tls.TLS_ECDHE_RSA_WITH_RC4_128_SHA)
		case "TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA":
			suites = append(suites, tls.TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA)
		case "TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA":
			suites = append(suites, tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA)
		case "TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA":
			suites = append(suites, tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA)
		case "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256":
			suites = append(suites, tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256)
		case "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256":
			suites = append(suites, tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256)
		case "TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384":
			suites = append(suites, tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384)
		case "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384":
			suites = append(suites, tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384)
		case "TLS_CHACHA20_POLY1305_SHA256":
			suites = append(suites, tls.TLS_CHACHA20_POLY1305_SHA256)
		case "TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256":
			suites = append(suites, tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256)
		case "TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256":
			suites = append(suites, tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256)
		default:
			return nil, fmt.Errorf("unsupported cipher suite: %s", suite)
		}
	}
	return suites, nil
}
