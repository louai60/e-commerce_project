package service

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/louai60/e-commerce_project/backend/product-service/config"
	"go.uber.org/zap"
)

type ImageUploadResult struct {
	URL      string
	AltText  string
	Position int
}

type ImageService struct {
	cld    *cloudinary.Cloudinary
	logger *zap.Logger
}

func NewImageService(cfg *config.Config, logger *zap.Logger) (*ImageService, error) {
	cld, err := cloudinary.NewFromParams(cfg.Cloudinary.CloudName, cfg.Cloudinary.APIKey, cfg.Cloudinary.APISecret)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Cloudinary: %v", err)
	}

	return &ImageService{
		cld:    cld,
		logger: logger,
	}, nil
}

func (s *ImageService) UploadImage(ctx context.Context, file *multipart.FileHeader, altText string, position int, folder string) (*ImageUploadResult, error) {
	// Open the file
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer src.Close()

	// Get file extension
	ext := filepath.Ext(file.Filename)

	// Upload the file to Cloudinary
	uploadResult, err := s.cld.Upload.Upload(ctx, src, uploader.UploadParams{
		Folder:   folder,
		PublicID: file.Filename[:len(file.Filename)-len(ext)],
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %v", err)
	}

	s.logger.Info("Image uploaded successfully",
		zap.String("public_id", uploadResult.PublicID),
		zap.String("url", uploadResult.SecureURL),
		zap.String("alt_text", altText),
		zap.Int("position", position),
	)

	return &ImageUploadResult{
		URL:      uploadResult.SecureURL,
		AltText:  altText,
		Position: position,
	}, nil
}

func (s *ImageService) DeleteImage(ctx context.Context, publicID string) error {
	_, err := s.cld.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID: publicID,
	})
	if err != nil {
		return fmt.Errorf("failed to delete image: %v", err)
	}

	s.logger.Info("Image deleted successfully",
		zap.String("public_id", publicID),
	)

	return nil
}
