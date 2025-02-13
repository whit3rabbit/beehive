package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"

	"github.com/whit3rabbit/beehive/manager/api/admin"
	"github.com/whit3rabbit/beehive/manager/api/handlers"
	"github.com/whit3rabbit/beehive/manager/internal/config"
	"github.com/whit3rabbit/beehive/manager/internal/logger"
	"github.com/whit3rabbit/beehive/manager/internal/mongodb"
	"github.com/whit3rabbit/beehive/manager/internal/setup"
	customMiddleware "github.com/whit3rabbit/beehive/manager/middleware"
	"github.com/whit3rabbit/beehive/manager/migrations"
	"github.com/whit3rabbit/beehive/manager/models"
)

func main() {
	// Parse command line flags
	flags := config.ParseFlags()

	// Check if this is a setup command
	if len(os.Args) > 1 && os.Args[1] == "setup" {
		setupCmd := flag.NewFlagSet("setup", flag.ExitOnError)
		skipPrompts := setupCmd.Bool("skip", false, "Skip prompts and generate random values")
		setupCmd.Parse(os.Args[2:])
		
		if err := setup.RunSetup(*skipPrompts); err != nil {
			fmt.Fprintf(os.Stderr, "Setup failed: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Load configuration
	cfg, err := config.LoadConfig(flags.ConfigFile)
	if err != nil {
		logger.Fatal("Error loading config", zap.Error(err))
	}

	// Merge configurations
	config.MergeConfig(cfg, flags)

	// Initialize Logger
	if err := logger.Initialize(cfg.Logging.Level); err != nil {
		logger.Fatal("Failed to initialize logger", zap.Error(err))
	}
	defer logger.Sync()

	// Log startup configuration (excluding sensitive data)
	logger.Info("Starting server with configuration",
		zap.String("host", cfg.Server.Host),
		zap.Int("port", cfg.Server.Port),
		zap.String("mongodb_database", cfg.MongoDB.Database),
		zap.String("log_level", cfg.Logging.Level),
		zap.Bool("tls_enabled", cfg.Server.TLS.Enabled))

	// Connect to MongoDB
	if err := mongodb.Connect(cfg.MongoDB.URI); err != nil {
		logger.Fatal("Error connecting to MongoDB", zap.Error(err))
	}

	// Create root context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start cleanup routine for login attempts
	go admin.CleanupLoginAttempts(ctx)

	// Run migrations
	db := mongodb.Client.Database(cfg.MongoDB.Database)
	allMigrations := []migrations.Migration{
		migrations.Migration0001,
	}

	if err := migrations.RunMigrations(db, allMigrations); err != nil {
		logger.Fatal("Error running migrations", zap.Error(err))
	}

	// Ensure admin user exists
	ensureAdminUser(db, cfg)

	// Create Echo instance and set up middleware
	e := echo.New()

	// Initialize rate limiter
	rateLimiter := customMiddleware.NewRateLimiter(
		cfg.Security.RateLimiting.MaxAttempts,
		time.Duration(cfg.Security.RateLimiting.WindowSeconds)*time.Second,
		time.Duration(cfg.Security.RateLimiting.BlockoutMinutes)*time.Minute,
	)

	setupRoutes(e, rateLimiter)

	// Serve static files for React frontend (if available)
	if cfg.Server.StaticDir != "" {
		e.Static("/", cfg.Server.StaticDir)
	}

	// Start server
	startServer(e, cfg)
}

func ensureAdminUser(db *mongo.Database, cfg *config.Config) {
	adminCollection := db.Collection("admins")
	ctxTimeout, cancelTimeout := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelTimeout()

	passwordPolicy := models.PasswordPolicy{
		MinLength:        cfg.Security.PasswordPolicy.MinLength,
		RequireUppercase: cfg.Security.PasswordPolicy.RequireUppercase,
		RequireLowercase: cfg.Security.PasswordPolicy.RequireLowercase,
		RequireNumbers:   cfg.Security.PasswordPolicy.RequireNumbers,
		RequireSpecial:   cfg.Security.PasswordPolicy.RequireSpecial,
	}

	var adminUser models.Admin
	err := adminCollection.FindOne(ctxTimeout, bson.M{"username": cfg.Admin.DefaultUsername}).Decode(&adminUser)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			hashedPassword, err := admin.GenerateHashPassword(cfg.Admin.DefaultPassword, passwordPolicy)
			if err != nil {
				logger.Fatal("Failed to hash default admin password", zap.Error(err))
			}

			_, err = adminCollection.InsertOne(ctxTimeout, models.Admin{
				Username:  cfg.Admin.DefaultUsername,
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
}

func setupRoutes(e *echo.Echo, rateLimiter customMiddleware.RateLimiter) {
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
}

func startServer(e *echo.Echo, cfg *config.Config) {
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	server := &http.Server{
		Addr:    addr,
		Handler: e,
	}

	if cfg.Server.TLS.Enabled {
		// Check if TLS cert and key files exist
		if _, err := os.Stat(cfg.Server.TLS.CertFile); os.IsNotExist(err) {
			logger.Fatal("TLS certificate file not found", zap.String("path", cfg.Server.TLS.CertFile))
		}
		if _, err := os.Stat(cfg.Server.TLS.KeyFile); os.IsNotExist(err) {
			logger.Fatal("TLS key file not found", zap.String("path", cfg.Server.TLS.KeyFile))
		}

		// Configure TLS settings
		if err := configureTLS(server, cfg); err != nil {
			logger.Fatal("Error configuring TLS", zap.Error(err))
		}

		logger.Info("Starting server with TLS", zap.String("address", "https://"+addr))
		if err := e.StartTLS(addr, cfg.Server.TLS.CertFile, cfg.Server.TLS.KeyFile); err != nil {
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

func configureTLS(server *http.Server, cfg *config.Config) error {
	minVersion, err := parseTLSVersion(cfg.Server.TLS.MinVersion)
	if err != nil {
		return fmt.Errorf("invalid TLS version: %w", err)
	}

	cipherSuites, err := parseCipherSuites(cfg.Server.TLS.CipherSuites)
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
