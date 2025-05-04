package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Server struct {
		Port        string
		Environment string
		TLS         struct {
			CertPath string
			KeyPath  string
		}
	}
	Database struct {
		Host            string
		Port            string
		User            string
		Password        string
		Name            string
		SSLMode         string
		MaxOpenConns    int
		MaxIdleConns    int
		ConnMaxLifetime time.Duration
		ConnMaxIdleTime time.Duration
		Replicas        []struct {
			Host            string
			Port            string
			User            string
			Password        string
			Name            string
			SSLMode         string
			MaxOpenConns    int
			MaxIdleConns    int
			ConnMaxLifetime time.Duration
			ConnMaxIdleTime time.Duration
		}
		ReplicaSelector string
	}
	Redis struct {
		Host string
		Port string
	}
	RateLimiter struct {
		Attempts int
		Duration time.Duration
	}
	Auth AuthConfig
}

type ServerConfig struct {
	Port                    string        `mapstructure:"port"`
	Environment             string        `mapstructure:"environment"`
	ServiceName             string        `mapstructure:"serviceName"`
	LogLevel                string        `mapstructure:"logLevel"`
	GracefulShutdownTimeout time.Duration `mapstructure:"gracefulShutdownTimeout"`
	TLS                     *TLSConfig    `mapstructure:"tls,omitempty"`
}

type TLSConfig struct {
	CertPath string `mapstructure:"certPath"`
	KeyPath  string `mapstructure:"keyPath"`
}

type DatabaseConfig struct {
	Host     string
	Port     string
	Name     string
	User     string
	Password string
	SSLMode  string
}

type AuthConfig struct {
	SecretKey            string        `mapstructure:"secretKey"`
	AccessTokenDuration  time.Duration `mapstructure:"accessTokenDuration"`
	RefreshTokenDuration time.Duration `mapstructure:"refreshTokenDuration"`
	TokenDuration        time.Duration `mapstructure:"tokenDuration"`
}

type RateLimiter struct {
	Attempts int           `mapstructure:"attempts"`
	Duration time.Duration `mapstructure:"duration"`
}

func (d *DatabaseConfig) DSN() string {
	sslMode := d.SSLMode
	if sslMode == "" {
		sslMode = "disable"
	}

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, sslMode,
	)

	if sslMode == "verify-full" {
		// Add SSL certificate configuration for production
		dsn += " sslcert=/path/to/cert.pem sslkey=/path/to/key.pem sslrootcert=/path/to/ca.pem"
	}

	return dsn
}

func LoadConfig() (*Config, error) {
	// Load .env file
	if err := loadEnv(); err != nil {
		return nil, fmt.Errorf("failed to load env: %w", err)
	}

	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	// Initialize viper
	v := viper.New()
	v.SetConfigName(fmt.Sprintf("config.%s", env))
	v.SetConfigType("yaml")
	v.AddConfigPath("./config")
	v.AddConfigPath("../config")

	// Set default values
	v.SetDefault("auth.accessTokenDuration", "1h")
	v.SetDefault("auth.refreshTokenDuration", "24h")
	v.SetDefault("rateLimiter.attempts", 5)
	v.SetDefault("rateLimiter.duration", "1m")

	// Enable environment variable replacement
	v.AutomaticEnv()
	v.SetEnvPrefix("APP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	// Override with environment variables
	if duration := os.Getenv("JWT_ACCESS_TOKEN_DURATION"); duration != "" {
		v.Set("auth.accessTokenDuration", duration)
	}
	if duration := os.Getenv("JWT_REFRESH_TOKEN_DURATION"); duration != "" {
		v.Set("auth.refreshTokenDuration", duration)
	}
	if attempts := os.Getenv("RATE_LIMIT_ATTEMPTS"); attempts != "" {
		v.Set("rateLimiter.attempts", attempts)
	}
	if duration := os.Getenv("RATE_LIMIT_DURATION"); duration != "" {
		v.Set("rateLimiter.duration", duration)
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Load sensitive values from environment
	if err := loadSensitiveValues(&config); err != nil {
		return nil, fmt.Errorf("failed to load sensitive values: %w", err)
	}

	// Validate config
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &config, nil
}

func loadEnv() error {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	// Load the main .env file
	if err := godotenv.Load(); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	}

	// Load environment-specific .env file
	envFile := fmt.Sprintf(".env.%s", env)
	if err := godotenv.Load(envFile); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	}

	return nil
}

func loadSensitiveValues(config *Config) error {
	// Load database password
	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		return errors.New("DB_PASSWORD environment variable is required")
	}
	config.Database.Password = dbPassword

	// Load JWT secret
	jwtSecret := os.Getenv("JWT_SECRET_KEY")
	if jwtSecret == "" {
		return errors.New("JWT_SECRET_KEY environment variable is required")
	}
	config.Auth.SecretKey = jwtSecret

	return nil
}

func validateConfig(config *Config) error {
	if config.Server.Port == "" {
		return errors.New("server port is required")
	}

	if config.Database.Host == "" || config.Database.Port == "" {
		return errors.New("database host and port are required")
	}

	if config.Server.Environment == "production" {
		if config.Database.SSLMode != "verify-full" {
			return errors.New("production environment requires SSL mode 'verify-full'")
		}
		if config.Server.TLS.CertPath == "" || config.Server.TLS.KeyPath == "" {
			return errors.New("TLS configuration is required in production")
		}
	}

	return nil
}
