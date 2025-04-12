package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	adminpb "github.com/louai60/e-commerce_project/backend/admin-service/proto"
)

// AdminHandler handles requests related to the admin dashboard.
type AdminHandler struct {
	client adminpb.AdminServiceClient
	logger *zap.Logger
}

// NewAdminHandler creates a new AdminHandler.
func NewAdminHandler(client adminpb.AdminServiceClient, logger *zap.Logger) *AdminHandler {
	return &AdminHandler{
		client: client,
		logger: logger,
	}
}

// GetDashboardStats retrieves dashboard statistics.
func (h *AdminHandler) GetDashboardStats(c *gin.Context) {
	h.logger.Info("API Gateway: GetDashboardStats called")

	// Prepare the gRPC request
	req := &adminpb.GetDashboardStatsRequest{} // Empty for now, add filters if needed

	// Call the admin gRPC service
	res, err := h.client.GetDashboardStats(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to call GetDashboardStats on admin service", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve dashboard stats"})
		return
	}

	// Return the response from the admin service
	c.JSON(http.StatusOK, res)
}
