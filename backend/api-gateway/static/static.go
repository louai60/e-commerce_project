package static

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SetupStaticFileServer configures routes for serving static files
func SetupStaticFileServer(router *gin.Engine, logger *zap.Logger) {
	// Get upload directory from environment or use default
	uploadDir := os.Getenv("LOCAL_STORAGE_PATH")
	if uploadDir == "" {
		uploadDir = "./uploads"
	}

	// Ensure the directory exists
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		logger.Error("Failed to create upload directory", zap.Error(err))
		return
	}

	// Create an absolute path
	absUploadDir, err := filepath.Abs(uploadDir)
	if err != nil {
		logger.Error("Failed to get absolute path for upload directory", zap.Error(err))
		return
	}

	logger.Info("Setting up static file server", zap.String("path", absUploadDir))

	// Serve static files
	router.StaticFS("/uploads", http.Dir(absUploadDir))
}
