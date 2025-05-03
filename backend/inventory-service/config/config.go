package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all configuration for the service
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Logging  LoggingConfig  `mapstructure:"logging"`
}

// ServerConfig holds the configuration for the gRPC server
type ServerConfig struct {
	Port string `mapstructure:"port"`
	Host string `mapstructure:"host"`
}

// DatabaseConfig holds the configuration for the database
type DatabaseConfig struct {
	Host                   string `mapstructure:"host"`
	Port                   string `mapstructure:"port"`
	User                   string `mapstructure:"user"`
	Password               string `mapstructure:"password"`
	Name                   string `mapstructure:"name"`
	MaxOpenConns           int    `mapstructure:"max_open_conns"`
	MaxIdleConns           int    `mapstructure:"max_idle_conns"`
	ConnMaxLifetimeMinutes int    `mapstructure:"conn_max_lifetime_minutes"`
}

// RedisConfig holds the configuration for Redis
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// LoggingConfig holds the configuration for logging
type LoggingConfig struct {
	Level string `mapstructure:"level"`
}

// LoadConfig loads the configuration from config files and environment variables
func LoadConfig() (*Config, error) {
	var config Config

	// Set default configuration file path
	configPath := "config"
	if os.Getenv("CONFIG_PATH") != "" {
		configPath = os.Getenv("CONFIG_PATH")
	}

	// Set environment
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	// Initialize viper
	v := viper.New()
	v.SetConfigName(fmt.Sprintf("config.%s", env))
	v.SetConfigType("yaml")
	v.AddConfigPath(configPath)
	v.AddConfigPath(".")

	// Read the config file
	if err := v.ReadInConfig(); err != nil {
		// It's okay if config file doesn't exist
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Override with environment variables
	v.SetEnvPrefix("INVENTORY")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Set defaults
	setDefaults(v)

	// Unmarshal config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("unable to decode config into struct: %w", err)
	}

	return &config, nil
}

// setDefaults sets default values for configuration
func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("server.port", "50055")
	v.SetDefault("server.host", "0.0.0.0")

	// Database defaults
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", "5432")
	v.SetDefault("database.user", "postgres")
	v.SetDefault("database.password", "postgres")
	v.SetDefault("database.name", "inventory_service")
	v.SetDefault("database.max_open_conns", 25)
	v.SetDefault("database.max_idle_conns", 5)
	v.SetDefault("database.conn_max_lifetime_minutes", 5)

	// Redis defaults
	v.SetDefault("redis.host", "localhost")
	v.SetDefault("redis.port", "6379")
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.db", 0)

	// Logging defaults
	v.SetDefault("logging.level", "info")
}
