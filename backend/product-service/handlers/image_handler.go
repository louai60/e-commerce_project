package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/louai60/e-commerce_project/backend/product-service/service"
)

type ImageHandler struct {
	imageService *service.ImageService
}

func NewImageHandler(imageService *service.ImageService) *ImageHandler {
	return &ImageHandler{
		imageService: imageService,
	}
}

func (h *ImageHandler) UploadImage(c *gin.Context) {
	// Get the file from the request
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	// Get additional fields
	altText := c.PostForm("alt_text")
	positionStr := c.PostForm("position")
	folder := c.DefaultQuery("folder", "products")

	// Convert position to integer
	position := 1 // Default position
	if positionStr != "" {
		pos, err := strconv.Atoi(positionStr)
		if err == nil {
			position = pos
		}
	}

	// Upload the image
	result, err := h.imageService.UploadImage(c.Request.Context(), file, altText, position, folder)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"url":      result.URL,
		"alt_text": result.AltText,
		"position": result.Position,
	})
}

func (h *ImageHandler) DeleteImage(c *gin.Context) {
	publicID := c.Param("public_id")
	if publicID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Public ID is required"})
		return
	}

	err := h.imageService.DeleteImage(c.Request.Context(), publicID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Image deleted successfully",
	})
}
