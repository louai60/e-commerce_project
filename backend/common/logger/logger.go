package logger

import (
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

var log *zap.Logger

// Initialize sets up the logger
func Initialize(env string) {
    var config zap.Config
    if env == "production" {
        config = zap.NewProductionConfig()
    } else {
        config = zap.NewDevelopmentConfig()
    }
    
    config.EncoderConfig.TimeKey = "timestamp"
    config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
    
    var err error
    log, err = config.Build()
    if err != nil {
        panic("failed to initialize logger: " + err.Error())
    }
}

// GetLogger returns the configured logger instance
func GetLogger() *zap.Logger {
    if log == nil {
        Initialize("development")
    }
    return log
}