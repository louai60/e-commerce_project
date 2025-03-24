package config

import (
    "github.com/spf13/viper"
)

type Config struct {
    Server   ServerConfig   `mapstructure:"server"`
    Services ServicesConfig `mapstructure:"services"`
}

type ServerConfig struct {
    Port        string `mapstructure:"port"`
    Environment string `mapstructure:"environment"`
    JWTSecret   string `mapstructure:"jwtSecret"`
}

type ServicesConfig struct {
    Product       ServiceConfig `mapstructure:"product"`
    User          ServiceConfig `mapstructure:"user"`
    Order         ServiceConfig `mapstructure:"order"`
    Payment       ServiceConfig `mapstructure:"payment"`
    Inventory     ServiceConfig `mapstructure:"inventory"`
    Cart          ServiceConfig `mapstructure:"cart"`
    Search        ServiceConfig `mapstructure:"search"`
    Review        ServiceConfig `mapstructure:"review"`
    Notification  ServiceConfig `mapstructure:"notification"`
    Shipping      ServiceConfig `mapstructure:"shipping"`
    Promotion     ServiceConfig `mapstructure:"promotion"`
    Recommendation ServiceConfig `mapstructure:"recommendation"`
}

type ServiceConfig struct {
    Host string `mapstructure:"host"`
    Port string `mapstructure:"port"`
}

func LoadConfig() (*Config, error) {
    viper.SetConfigName("config")
    viper.SetConfigType("yaml")
    viper.AddConfigPath("./config")
    viper.AddConfigPath("../config")

    viper.AutomaticEnv()

    if err := viper.ReadInConfig(); err != nil {
        return nil, err
    }

    config := &Config{}
    if err := viper.Unmarshal(config); err != nil {
        return nil, err
    }

    return config, nil
}
