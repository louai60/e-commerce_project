package config

import (
    "github.com/spf13/viper"
)

type Config struct {
    Server struct {
        Port        string `mapstructure:"port"`
        Environment string `mapstructure:"environment"`
    }
    Services struct {
        Product struct {
            Host string `mapstructure:"host"`
            Port string `mapstructure:"port"`
        }
        User struct {
            Host string `mapstructure:"host"`
            Port string `mapstructure:"port"`
        }
    }
    Auth struct {
        AdminSecretKey string `mapstructure:"adminSecretKey"`
        TokenDuration  string `mapstructure:"tokenDuration"`
    }
}

func LoadConfig() (*Config, error) {
    viper.SetConfigName("config")
    viper.SetConfigType("yaml")
    viper.AddConfigPath("./config")

    var config Config
    if err := viper.ReadInConfig(); err != nil {
        return nil, err
    }

    if err := viper.Unmarshal(&config); err != nil {
        return nil, err
    }

    return &config, nil
}