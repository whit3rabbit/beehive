package setup

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

type Config struct {
	ServerHost          string
	ServerPort          int
	MongoHost           string
	MongoPort          int
	MongoUser          string
	MongoPass          string
	MongoDatabase      string
	AdminUser          string
	AdminPass          string
	JWTSecret          string
	APIKey             string
	APISecret          string
	TLSEnabled         bool
	TLSCertFile        string
	TLSKeyFile         string
	LogLevel           string
}

// RunSetup performs the complete setup process
func RunSetup(skipPrompts bool) error {
	config, err := getConfig(skipPrompts)
	if err != nil {
		return fmt.Errorf("failed to get configuration: %w", err)
	}

	// Create necessary directories
	if err := createDirectories(config); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Generate SSL certificates
	if err := generateSSLCert(config); err != nil {
		return fmt.Errorf("failed to generate SSL certificates: %w", err)
	}

	// Setup MongoDB
	if err := setupMongoDB(config); err != nil {
		return fmt.Errorf("failed to setup MongoDB: %w", err)
	}

	// Write configuration to .env file
	if err := writeEnvFile(config); err != nil {
		return fmt.Errorf("failed to write .env file: %w", err)
	}

	printSummary(config)
	return nil
}

func getConfig(skipPrompts bool) (*Config, error) {
	if skipPrompts {
		return getDefaultConfig(), nil
	}
	return promptForConfig()
}

func getDefaultConfig() *Config {
	return &Config{
		ServerHost:     "0.0.0.0",
		ServerPort:     8080,
		MongoHost:      "localhost",
		MongoPort:      27017,
		MongoUser:      "manager",
		MongoPass:      generateSecureString(24),
		MongoDatabase:  "manager_db",
		AdminUser:      "admin",
		AdminPass:      generateSecureString(24),
		JWTSecret:      generateSecureString(32),
		APIKey:         generateSecureString(32),
		APISecret:      generateSecureString(32),
		TLSEnabled:     true,
		TLSCertFile:    "certs/server.crt",
		TLSKeyFile:     "certs/server.key",
		LogLevel:       "info",
	}
}

func promptForConfig() (*Config, error) {
	config := &Config{}
	questions := []*survey.Question{
		{
			Name: "serverHost",
			Prompt: &survey.Input{
				Message: "Server Host:",
				Default: "0.0.0.0",
			},
		},
		{
			Name: "serverPort",
			Prompt: &survey.Input{
				Message: "Server Port:",
				Default: "8080",
			},
		},
		{
			Name: "mongoHost",
			Prompt: &survey.Input{
				Message: "MongoDB Host:",
				Default: "localhost",
			},
		},
		{
			Name: "mongoPort",
			Prompt: &survey.Input{
				Message: "MongoDB Port:",
				Default: "27017",
			},
		},
		{
			Name: "mongoUser",
			Prompt: &survey.Input{
				Message: "MongoDB Username:",
				Default: "manager",
			},
		},
		{
			Name: "adminUser",
			Prompt: &survey.Input{
				Message: "Admin Username:",
				Default: "admin",
			},
		},
		{
			Name: "logLevel",
			Prompt: &survey.Select{
				Message: "Log Level:",
				Options: []string{"debug", "info", "warn", "error"},
				Default: "info",
			},
		},
	}

	answers := make(map[string]interface{})
	if err := survey.Ask(questions, &answers); err != nil {
		return nil, err
	}

	// Map answers to config
	config.ServerHost = answers["serverHost"].(string)
	config.ServerPort = 8080 // Default port
	config.MongoHost = answers["mongoHost"].(string)
	config.MongoPort = 27017 // Default MongoDB port
	config.MongoUser = answers["mongoUser"].(string)
	config.MongoPass = generateSecureString(24)
	config.MongoDatabase = "manager_db"
	config.AdminUser = answers["adminUser"].(string)
	config.AdminPass = generateSecureString(24)
	config.JWTSecret = generateSecureString(32)
	config.APIKey = generateSecureString(32)
	config.APISecret = generateSecureString(32)
	config.TLSEnabled = true
	config.TLSCertFile = "certs/server.crt"
	config.TLSKeyFile = "certs/server.key"
	config.LogLevel = answers["logLevel"].(string)

	return config, nil
}

func generateSecureString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	b := make([]byte, length)
	rand.Read(b)
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	return string(b)
}

func createDirectories(config *Config) error {
	dirs := []string{
		filepath.Dir(config.TLSCertFile),
		"logs",
		"data",
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	return nil
}

func generateSSLCert(config *Config) error {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("failed to generate private key: %w", err)
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Manager"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:          []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IPAddresses:          []net.IP{net.ParseIP("127.0.0.1")},
		DNSNames:             []string{"localhost"},
	}

	certBytes, err := x509.CreateCertificate(
		rand.Reader,
		&template,
		&template,
		&privateKey.PublicKey,
		privateKey,
	)
	if err != nil {
		return fmt.Errorf("failed to create certificate: %w", err)
	}

	certOut, err := os.Create(config.TLSCertFile)
	if err != nil {
		return fmt.Errorf("failed to create certificate file: %w", err)
	}
	defer certOut.Close()

	if err := pem.Encode(certOut, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	}); err != nil {
		return fmt.Errorf("failed to write certificate: %w", err)
	}

	keyOut, err := os.OpenFile(config.TLSKeyFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to create private key file: %w", err)
	}
	defer keyOut.Close()

	if err := pem.Encode(keyOut, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}); err != nil {
		return fmt.Errorf("failed to write private key: %w", err)
	}

	return nil
}

func setupMongoDB(config *Config) error {
	ctx := context.Background()

	// Connect to MongoDB without auth first
	mongoURI := fmt.Sprintf("mongodb://%s:%d", config.MongoHost, config.MongoPort)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}
	defer client.Disconnect(ctx)

	// Create database and user
	if err := createMongoDBUser(ctx, client, config); err != nil {
		return fmt.Errorf("failed to create MongoDB user: %w", err)
	}

	// Reconnect with new user credentials
	mongoURIWithAuth := fmt.Sprintf("mongodb://%s:%s@%s:%d/%s",
		config.MongoUser,
		config.MongoPass,
		config.MongoHost,
		config.MongoPort,
		config.MongoDatabase,
	)
	clientWithAuth, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURIWithAuth))
	if err != nil {
		return fmt.Errorf("failed to connect with auth: %w", err)
	}
	defer clientWithAuth.Disconnect(ctx)

	// Create admin user
	if err := createAdminUser(ctx, clientWithAuth, config); err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	// Create default worker role
	if err := createDefaultRole(ctx, clientWithAuth, config); err != nil {
		return fmt.Errorf("failed to create default role: %w", err)
	}

	return nil
}

func createMongoDBUser(ctx context.Context, client *mongo.Client, config *Config) error {
	cmd := bson.D{
		{Key: "createUser", Value: config.MongoUser},
		{Key: "pwd", Value: config.MongoPass},
		{Key: "roles", Value: bson.A{
			bson.D{
				{Key: "role", Value: "readWrite"},
				{Key: "db", Value: config.MongoDatabase},
			},
		}},
	}

	err := client.Database(config.MongoDatabase).RunCommand(ctx, cmd).Err()
	if err != nil {
		return fmt.Errorf("failed to create MongoDB user: %w", err)
	}

	return nil
}

func createAdminUser(ctx context.Context, client *mongo.Client, config *Config) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(config.AdminPass), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	collection := client.Database(config.MongoDatabase).Collection("admins")
	_, err = collection.InsertOne(ctx, bson.M{
		"username":   config.AdminUser,
		"password":   string(hashedPassword),
		"created_at": time.Now(),
		"updated_at": time.Now(),
	})
	if err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	return nil
}

func createDefaultRole(ctx context.Context, client *mongo.Client, config *Config) error {
	collection := client.Database(config.MongoDatabase).Collection("roles")
	_, err := collection.InsertOne(ctx, bson.M{
		"name":        "worker",
		"description": "Default worker role",
		"created_at":  time.Now(),
	})
	if err != nil {
		return fmt.Errorf("failed to create default role: %w", err)
	}

	return nil
}

func writeEnvFile(config *Config) error {
	envContent := fmt.Sprintf(`# Server Configuration
SERVER_HOST=%s
SERVER_PORT=%d
STATIC_DIR=./frontend/build

# TLS Configuration
TLS_ENABLED=%t
TLS_CERT_FILE=%s
TLS_KEY_FILE=%s

# MongoDB Configuration
MONGODB_URI=mongodb://%s:%s@%s:%d/%s
MONGODB_DATABASE=%s

# Authentication
JWT_SECRET=%s
TOKEN_EXPIRATION_HOURS=24
API_KEY=%s
API_SECRET=%s

# Admin Configuration
ADMIN_DEFAULT_USERNAME=%s
ADMIN_DEFAULT_PASSWORD=%s

# Logging
LOG_LEVEL=%s
`,
		config.ServerHost,
		config.ServerPort,
		config.TLSEnabled,
		config.TLSCertFile,
		config.TLSKeyFile,
		config.MongoUser,
		config.MongoPass,
		config.MongoHost,
		config.MongoPort,
		config.MongoDatabase,
		config.MongoDatabase,
		config.JWTSecret,
		config.APIKey,
		config.APISecret,
		config.AdminUser,
		config.AdminPass,
		config.LogLevel,
	)

	return os.WriteFile(".env", []byte(envContent), 0600)
}

func printSummary(config *Config) {
	fmt.Println("\nSetup completed successfully!")
	fmt.Printf("Admin Credentials:\n")
	fmt.Printf("  Username: %s\n", config.AdminUser)
	fmt.Printf("  Password: %s\n", config.AdminPass)
	fmt.Printf("\nMongoDB Credentials:\n")
	fmt.Printf("  Username: %s\n", config.MongoUser)
	fmt.Printf("  Password: %s\n", config.MongoPass)
	fmt.Printf("\nAPI Credentials:\n")
	fmt.Printf("  API Key: %s\n", config.APIKey)
	fmt.Printf("  API Secret: %s\n", config.APISecret)
	fmt.Printf("\nConfiguration has been saved to .env file\n")
	fmt.Printf("SSL certificates have been generated in %s\n", filepath.Dir(config.TLSCertFile))
}