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

    "github.com/whit3rabbit/beehive/manager/internal/config"
)

// RunSetup performs the complete setup process
func RunSetup(skipPrompts bool) error {
    cfg, err := getConfig(skipPrompts)
    if err != nil {
        return fmt.Errorf("failed to get configuration: %w", err)
    }

    // Create necessary directories
    if err := createDirectories(cfg); err != nil {
        return fmt.Errorf("failed to create directories: %w", err)
    }

    // Generate SSL certificates
    if err := generateSSLCert(cfg); err != nil {
        return fmt.Errorf("failed to generate SSL certificates: %w", err)
    }

    // Setup MongoDB
    if err := setupMongoDB(cfg); err != nil {
        return fmt.Errorf("failed to setup MongoDB: %w", err)
    }

    // Write configuration to .env file
    if err := writeEnvFile(cfg); err != nil {
        return fmt.Errorf("failed to write .env file: %w", err)
    }

    printSummary(cfg)
    return nil
}

func getConfig(skipPrompts bool) (*config.Config, error) {
    if skipPrompts {
        return getDefaultConfig(), nil
    }
    return promptForConfig()
}

func getDefaultConfig() *config.Config {
    return &config.Config{
        Server: config.ServerConfig{
            Host: "0.0.0.0",
            Port: 8080,
            StaticDir: "./frontend/build",
            TLS: config.TLSConfig{
                Enabled:  true,
                CertFile: "certs/server.crt",
                KeyFile:  "certs/server.key",
            },
        },
        MongoDB: config.MongoDBConfig{
            Host:     "localhost",
            Port:     27017,
            User:     "manager",
            Pass:     generateSecureString(24),
            Database: "manager_db",
        },
        Auth: config.AuthConfig{
            JWTSecret:            generateSecureString(32),
            TokenExpirationHours: 24,
            APIKey:               generateSecureString(32),
            APISecret:            generateSecureString(32),
        },
        Admin: config.AdminConfig{
            DefaultUsername: "admin",
            DefaultPassword: generateSecureString(24),
        },
        Logging: config.LoggingConfig{
            Level: "info",
        },
        Security: struct {
            PasswordPolicy struct {
                MinLength        int  `yaml:"min_length"`
                RequireUppercase bool `yaml:"require_uppercase"`
                RequireLowercase bool `yaml:"require_lowercase"`
                RequireNumbers   bool `yaml:"require_numbers"`
                RequireSpecial   bool `yaml:"require_special"`
            } `yaml:"password_policy"`
            RateLimiting config.RateLimiterConfig `yaml:"rate_limiting"`
        }{
            PasswordPolicy: struct {
                MinLength        int  `yaml:"min_length"`
                RequireUppercase bool `yaml:"require_uppercase"`
                RequireLowercase bool `yaml:"require_lowercase"`
                RequireNumbers   bool `yaml:"require_numbers"`
                RequireSpecial   bool `yaml:"require_special"`
            }{
                MinLength:        8,
                RequireUppercase: true,
                RequireLowercase: true,
                RequireNumbers:   true,
                RequireSpecial:   true,
            },
            RateLimiting: config.RateLimiterConfig{
                MaxAttempts:     5,
                WindowSeconds:   300,
                BlockoutMinutes: 15,
            },
        },
    }
}

func promptForConfig() (*config.Config, error) {
    cfg := getDefaultConfig()
    questions := []*survey.Question{
        {
            Name: "serverHost",
            Prompt: &survey.Input{
                Message: "Server Host:",
                Default: cfg.Server.Host,
            },
        },
        {
            Name: "serverPort",
            Prompt: &survey.Input{
                Message: "Server Port:",
                Default: fmt.Sprintf("%d", cfg.Server.Port),
            },
        },
        {
            Name: "mongoHost",
            Prompt: &survey.Input{
                Message: "MongoDB Host:",
                Default: cfg.MongoDB.Host,
            },
        },
        {
            Name: "mongoPort",
            Prompt: &survey.Input{
                Message: "MongoDB Port:",
                Default: fmt.Sprintf("%d", cfg.MongoDB.Port),
            },
        },
        {
            Name: "mongoUser",
            Prompt: &survey.Input{
                Message: "MongoDB Username:",
                Default: cfg.MongoDB.User,
            },
        },
        {
            Name: "adminUser",
            Prompt: &survey.Input{
                Message: "Admin Username:",
                Default: cfg.Admin.DefaultUsername,
            },
        },
        {
            Name: "logLevel",
            Prompt: &survey.Select{
                Message: "Log Level:",
                Options: []string{"debug", "info", "warn", "error"},
                Default: cfg.Logging.Level,
            },
        },
    }

    answers := make(map[string]interface{})
    if err := survey.Ask(questions, &answers); err != nil {
        return nil, err
    }

    // Update config with answers
    cfg.Server.Host = answers["serverHost"].(string)
    if port, ok := answers["serverPort"].(string); ok {
        fmt.Sscanf(port, "%d", &cfg.Server.Port)
    }
    cfg.MongoDB.Host = answers["mongoHost"].(string)
    if port, ok := answers["mongoPort"].(string); ok {
        fmt.Sscanf(port, "%d", &cfg.MongoDB.Port)
    }
    cfg.MongoDB.User = answers["mongoUser"].(string)
    cfg.Admin.DefaultUsername = answers["adminUser"].(string)
    cfg.Logging.Level = answers["logLevel"].(string)

    return cfg, nil
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

func createDirectories(cfg *config.Config) error {
    dirs := []string{
        filepath.Dir(cfg.Server.TLS.CertFile),
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

func generateSSLCert(cfg *config.Config) error {
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
        ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
        BasicConstraintsValid: true,
        IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
        DNSNames:              []string{"localhost"},
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

    certOut, err := os.Create(cfg.Server.TLS.CertFile)
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

    keyOut, err := os.OpenFile(cfg.Server.TLS.KeyFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
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

func setupMongoDB(cfg *config.Config) error {
    ctx := context.Background()

    // Connect to MongoDB without auth first
    mongoURI := fmt.Sprintf("mongodb://%s:%d", cfg.MongoDB.Host, cfg.MongoDB.Port)
    client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
    if err != nil {
        return fmt.Errorf("failed to connect to MongoDB: %w", err)
    }
    defer client.Disconnect(ctx)

    // Create database and user
    if err := createMongoDBUser(ctx, client, cfg); err != nil {
        return fmt.Errorf("failed to create MongoDB user: %w", err)
    }

    // Reconnect with new user credentials
    mongoURIWithAuth := fmt.Sprintf("mongodb://%s:%s@%s:%d/%s",
        cfg.MongoDB.User,
        cfg.MongoDB.Pass,
        cfg.MongoDB.Host,
        cfg.MongoDB.Port,
        cfg.MongoDB.Database,
    )
    clientWithAuth, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURIWithAuth))
    if err != nil {
        return fmt.Errorf("failed to connect with auth: %w", err)
    }
    defer clientWithAuth.Disconnect(ctx)

    // Create admin user
    if err := createAdminUser(ctx, clientWithAuth, cfg); err != nil {
        return fmt.Errorf("failed to create admin user: %w", err)
    }

    // Create default worker role
    if err := createDefaultRole(ctx, clientWithAuth, cfg); err != nil {
        return fmt.Errorf("failed to create default role: %w", err)
    }

    return nil
}

func createMongoDBUser(ctx context.Context, client *mongo.Client, cfg *config.Config) error {
    cmd := bson.D{
        {Key: "createUser", Value: cfg.MongoDB.User},
        {Key: "pwd", Value: cfg.MongoDB.Pass},
        {Key: "roles", Value: bson.A{
            bson.D{
                {Key: "role", Value: "readWrite"},
                {Key: "db", Value: cfg.MongoDB.Database},
            },
        }},
    }

    err := client.Database(cfg.MongoDB.Database).RunCommand(ctx, cmd).Err()
    if err != nil {
        return fmt.Errorf("failed to create MongoDB user: %w", err)
    }

    return nil
}

func createAdminUser(ctx context.Context, client *mongo.Client, cfg *config.Config) error {
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(cfg.Admin.DefaultPassword), bcrypt.DefaultCost)
    if err != nil {
        return fmt.Errorf("failed to hash password: %w", err)
    }

    collection := client.Database(cfg.MongoDB.Database).Collection("admins")
    _, err = collection.InsertOne(ctx, bson.M{
        "username":   cfg.Admin.DefaultUsername,
        "password":   string(hashedPassword),
        "created_at": time.Now(),
        "updated_at": time.Now(),
    })
    if err != nil {
        return fmt.Errorf("failed to create admin user: %w", err)
    }

    return nil
}

func createDefaultRole(ctx context.Context, client *mongo.Client, cfg *config.Config) error {
    collection := client.Database(cfg.MongoDB.Database).Collection("roles")
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

func writeEnvFile(cfg *config.Config) error {
    envContent := fmt.Sprintf(`# Server Configuration
SERVER_HOST=%s
SERVER_PORT=%d
STATIC_DIR=%s

# TLS Configuration
TLS_ENABLED=%t
TLS_CERT_FILE=%s
TLS_KEY_FILE=%s

# MongoDB Configuration
MONGODB_URI=mongodb://%s:%s@%s:%d/%s
MONGODB_DATABASE=%s

# Authentication
JWT_SECRET=%s
TOKEN_EXPIRATION_HOURS=%d
API_KEY=%s
API_SECRET=%s

# Admin Configuration
ADMIN_DEFAULT_USERNAME=%s
ADMIN_DEFAULT_PASSWORD=%s

# Logging
LOG_LEVEL=%s`,
        cfg.Server.Host,
        cfg.Server.Port,
        cfg.Server.StaticDir,
        cfg.Server.TLS.Enabled,
        cfg.Server.TLS.CertFile,
        cfg.Server.TLS.KeyFile,
        cfg.MongoDB.User,
        cfg.MongoDB.Pass,
        cfg.MongoDB.Host,
        cfg.MongoDB.Port,
        cfg.MongoDB.Database,
        cfg.MongoDB.Database,
        cfg.Auth.JWTSecret,
        cfg.Auth.TokenExpirationHours,
        cfg.Auth.APIKey,
        cfg.Auth.APISecret,
        cfg.Admin.DefaultUsername,
        cfg.Admin.DefaultPassword,
        cfg.Logging.Level)

    return os.WriteFile(".env", []byte(envContent), 0600)
}

func printSummary(cfg *config.Config) {
    fmt.Println("\nSetup completed successfully!")
    fmt.Printf("Admin Credentials:\n")
    fmt.Printf("  Username: %s\n", cfg.Admin.DefaultUsername)
    fmt.Printf("  Password: %s\n", cfg.Admin.DefaultPassword)
    fmt.Printf("\nMongoDB Credentials:\n")
    fmt.Printf("  Username: %s\n", cfg.MongoDB.User)
    fmt.Printf("  Password: %s\n", cfg.MongoDB.Pass)
    fmt.Printf("\nAPI Credentials:\n")
    fmt.Printf("  API Key: %s\n", cfg.Auth.APIKey)
    fmt.Printf("  API Secret: %s\n", cfg.Auth.APISecret)
    fmt.Printf("\nConfiguration has been saved to .env file\n")
    fmt.Printf("SSL certificates have been generated in %s\n", filepath.Dir(cfg.Server.TLS.CertFile))
}