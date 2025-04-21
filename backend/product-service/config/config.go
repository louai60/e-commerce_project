package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// Config holds all configuration for our program
type Config struct {
	Server     ServerConfig   `yaml:"server"`
	Database   DatabaseConfig `yaml:"database"`
	Redis      RedisConfig    `yaml:"redis"`
	Secrets    SecretsConfig  `yaml:"secrets"`
	Cloudinary struct {
		CloudName string
		APIKey    string
		APISecret string
	}
}

// ServerConfig holds all server-related configuration
type ServerConfig struct {
	Port         string `mapstructure:"port"`
	Environment  string `mapstructure:"environment"`
	ServiceName  string `mapstructure:"serviceName"`
	LogLevel     string `mapstructure:"logLevel"`
	AllowOrigins string `mapstructure:"allowOrigins"`
}

// DatabaseConfig holds all database-related configuration
type DatabaseConfig struct {
	Host string `mapstructure:"host"`
	Port string `mapstructure:"port"`
	Name string `mapstructure:"name"`
	User string `mapstructure:"user"`
	// Password string `mapstructure:"-"`
	SSLMode string `mapstructure:"sslMode"`
}

// SecretsConfig holds all sensitive configuration that comes from env vars
type SecretsConfig struct {
	DatabasePassword string
	JWTSecret        string
	APIKeys          map[string]string
}

type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

// LoadConfig reads configuration from files and environment variables
func LoadConfig(logger *zap.Logger) (*Config, error) {
	config := &Config{}

	// Set default environment if not set
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	// Initialize viper for config file
	v := viper.New()
	v.SetConfigName(fmt.Sprintf("config.%s", env))
	v.SetConfigType("yaml")
	v.AddConfigPath("./config")
	v.AddConfigPath("../config")

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Unmarshal config file
	if err := v.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Load secrets from environment variables
	if err := loadSecrets(config); err != nil {
		return nil, fmt.Errorf("failed to load secrets: %w", err)
	}

	// Validate config
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	// Log non-sensitive configuration
	logger.Info("Configuration loaded successfully",
		zap.String("environment", env),
		zap.String("service", config.Server.ServiceName),
		zap.String("port", config.Server.Port),
	)

	return config, nil
}

// loadSecrets loads sensitive configuration from environment variables
func loadSecrets(config *Config) error {
	// Initialize secrets map
	config.Secrets = SecretsConfig{
		APIKeys: make(map[string]string),
	}

	// Load database password
	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		return fmt.Errorf("DB_PASSWORD environment variable is required")
	}
	config.Secrets.DatabasePassword = dbPassword

	// Load JWT secret
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return fmt.Errorf("JWT_SECRET environment variable is required")
	}
	config.Secrets.JWTSecret = jwtSecret

	// Load API keys
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, "API_KEY_") {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) == 2 {
				keyName := strings.TrimPrefix(parts[0], "API_KEY_")
				config.Secrets.APIKeys[keyName] = parts[1]
			}
		}
	}

	return nil
}

// validateConfig ensures all required configuration is present
func validateConfig(config *Config) error {
	if config.Server.Port == "" {
		return fmt.Errorf("server port is required")
	}
	if config.Server.ServiceName == "" {
		return fmt.Errorf("service name is required")
	}
	if config.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	// Add more validation as needed
	return nil
}

// GetDSN returns the database connection string
func (c *Config) GetDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Secrets.DatabasePassword,
		c.Database.Name,
		c.Database.SSLMode,
	)
}
