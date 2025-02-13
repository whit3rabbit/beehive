package config

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

// RateLimiterConfig holds rate limiting configuration
type RateLimiterConfig struct {
	MaxAttempts     int `yaml:"max_attempts"`
	WindowSeconds   int `yaml:"window_seconds"`
	BlockoutMinutes int `yaml:"blockout_minutes"`
}

type ServerConfig struct {
    Host      string `yaml:"host"`
    Port      int    `yaml:"port"`
    StaticDir string `yaml:"static_dir"`
    TLS       TLSConfig `yaml:"tls"`
}

type TLSConfig struct {
    Enabled      bool   `yaml:"enabled"`
    CertFile     string `yaml:"cert_file"`
    KeyFile      string `yaml:"key_file"`
    MinVersion   string `yaml:"min_version"`
    CipherSuites []string `yaml:"cipher_suites"`
}

type MongoDBConfig struct {
    URI      string `yaml:"uri"`
    Database string `yaml:"database"`
    Host     string `yaml:"host"`
    Port     int    `yaml:"port"`
    User     string `yaml:"user"`
    Pass     string `yaml:"pass"`
}

type AuthConfig struct {
    JWTSecret            string `yaml:"jwt_secret"`
    TokenExpirationHours int    `yaml:"token_expiration_hours"`
    APIKey              string `yaml:"api_key"`
    APISecret           string `yaml:"api_secret"`
}

type AdminConfig struct {
    DefaultUsername string `yaml:"default_username"`
    DefaultPassword string `yaml:"default_password"`
}

type LoggingConfig struct {
    Level string `yaml:"level"`
}

// Config holds all configuration settings
type Config struct {
    Server   ServerConfig   `yaml:"server"`
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
    MongoDB  MongoDBConfig  `yaml:"mongodb"`
    Auth     AuthConfig     `yaml:"auth"`
    Admin    AdminConfig    `yaml:"admin"`
    Logging  LoggingConfig  `yaml:"logging"`
}

// CLIFlags holds all command line arguments
type CLIFlags struct {
	ConfigFile     string
	ServerHost     string
	ServerPort     int
	MongoURI       string
	MongoDatabase  string
	LogLevel       string
	TLSEnabled     bool
	TLSCertFile    string
	TLSKeyFile     string
	JWTSecret      string
	AdminUsername  string
	AdminPassword  string
}

// LoadConfig loads configuration from files and environment variables
func LoadConfig(filename string) (*Config, error) {
	// Load .env file if it exists
	godotenv.Load()

	// Read YAML file
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	// Expand environment variables in YAML
	expandedData := []byte(os.ExpandEnv(string(data)))

	config := &Config{}
	if err := yaml.Unmarshal(expandedData, config); err != nil {
		return nil, fmt.Errorf("error parsing YAML: %w", err)
	}

	// Set defaults for any missing values
	setConfigDefaults(config)

	// Validate configuration
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// ParseFlags parses command line arguments
func ParseFlags() *CLIFlags {
	flags := &CLIFlags{}

	// Configuration file
	flag.StringVar(&flags.ConfigFile, "config", "config/config.yaml", "Path to configuration file")

	// Server settings
	flag.StringVar(&flags.ServerHost, "host", "", "Server host address")
	flag.IntVar(&flags.ServerPort, "port", 0, "Server port")

	// MongoDB settings
	flag.StringVar(&flags.MongoURI, "mongo-uri", "", "MongoDB URI")
	flag.StringVar(&flags.MongoDatabase, "mongo-db", "", "MongoDB database name")

	// Logging
	flag.StringVar(&flags.LogLevel, "log-level", "", "Log level (debug, info, warn, error)")

	// TLS settings
	flag.BoolVar(&flags.TLSEnabled, "tls", false, "Enable TLS")
	flag.StringVar(&flags.TLSCertFile, "tls-cert", "", "TLS certificate file path")
	flag.StringVar(&flags.TLSKeyFile, "tls-key", "", "TLS key file path")

	// Security settings
	flag.StringVar(&flags.JWTSecret, "jwt-secret", "", "JWT secret key")
	flag.StringVar(&flags.AdminUsername, "admin-user", "", "Default admin username")
	flag.StringVar(&flags.AdminPassword, "admin-pass", "", "Default admin password")

	flag.Parse()

	return flags
}

// MergeConfig merges configuration from different sources
func MergeConfig(config *Config, flags *CLIFlags) {
	if flags.ServerHost != "" {
		config.Server.Host = flags.ServerHost
	}
	if flags.ServerPort != 0 {
		config.Server.Port = flags.ServerPort
	}
	if flags.MongoURI != "" {
		config.MongoDB.URI = flags.MongoURI
	}
	if flags.MongoDatabase != "" {
		config.MongoDB.Database = flags.MongoDatabase
	}
	if flags.LogLevel != "" {
		config.Logging.Level = flags.LogLevel
	}
	if flags.TLSEnabled {
		config.Server.TLS.Enabled = true
		if flags.TLSCertFile != "" {
			config.Server.TLS.CertFile = flags.TLSCertFile
		}
		if flags.TLSKeyFile != "" {
			config.Server.TLS.KeyFile = flags.TLSKeyFile
		}
	}
	if flags.JWTSecret != "" {
		config.Auth.JWTSecret = flags.JWTSecret
	}
	if flags.AdminUsername != "" {
		config.Admin.DefaultUsername = flags.AdminUsername
	}
	if flags.AdminPassword != "" {
		config.Admin.DefaultPassword = flags.AdminPassword
	}
}

// setConfigDefaults sets default values for configuration
func setConfigDefaults(config *Config) {
	if config.Server.Host == "" {
		config.Server.Host = "0.0.0.0"
	}
	if config.Server.Port == 0 {
		config.Server.Port = 8080
	}
	if config.Server.StaticDir == "" {
		config.Server.StaticDir = "./frontend/build"
	}
	if config.Auth.TokenExpirationHours == 0 {
		config.Auth.TokenExpirationHours = 24
	}
	if config.Logging.Level == "" {
		config.Logging.Level = "info"
	}
	if config.Security.PasswordPolicy.MinLength == 0 {
		config.Security.PasswordPolicy.MinLength = 8
	}
	if config.Security.RateLimiting.MaxAttempts == 0 {
		config.Security.RateLimiting.MaxAttempts = 5
	}
	if config.Security.RateLimiting.WindowSeconds == 0 {
		config.Security.RateLimiting.WindowSeconds = 300 // 5 minutes
	}
	if config.Security.RateLimiting.BlockoutMinutes == 0 {
		config.Security.RateLimiting.BlockoutMinutes = 15
	}
}

// validateConfig checks if the configuration is valid
func validateConfig(config *Config) error {
	var errors []string

	// Validate MongoDB configuration
	if config.MongoDB.URI == "" {
		errors = append(errors, "MongoDB URI is required")
	}
	if config.MongoDB.Database == "" {
		errors = append(errors, "MongoDB database name is required")
	}

	// Validate TLS configuration if enabled
	if config.Server.TLS.Enabled {
		if config.Server.TLS.CertFile == "" {
			errors = append(errors, "TLS certificate file is required when TLS is enabled")
		}
		if config.Server.TLS.KeyFile == "" {
			errors = append(errors, "TLS key file is required when TLS is enabled")
		}
	}

	// Validate authentication configuration
	if config.Auth.JWTSecret == "" {
		errors = append(errors, "JWT secret is required")
	}
	if config.Admin.DefaultUsername == "" {
		errors = append(errors, "Default admin username is required")
	}
	if config.Admin.DefaultPassword == "" {
		errors = append(errors, "Default admin password is required")
	}

	// Validate password policy
	if config.Security.PasswordPolicy.MinLength < 8 {
		errors = append(errors, "Password minimum length must be at least 8 characters")
	}

	// Validate rate limiting
	if config.Security.RateLimiting.MaxAttempts < 1 {
		errors = append(errors, "Rate limiting max attempts must be at least 1")
	}
	if config.Security.RateLimiting.WindowSeconds < 1 {
		errors = append(errors, "Rate limiting window must be at least 1 second")
	}
	if config.Security.RateLimiting.BlockoutMinutes < 1 {
		errors = append(errors, "Rate limiting blockout period must be at least 1 minute")
	}

	if len(errors) > 0 {
		return fmt.Errorf("configuration validation errors:\n- %s", strings.Join(errors, "\n- "))
	}

	return nil
}

// WriteConfig writes the configuration to a YAML file
func WriteConfig(config *Config, filename string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("error marshaling config: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}

	return nil
}
