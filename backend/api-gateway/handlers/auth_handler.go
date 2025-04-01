package handlers

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	pb "github.com/louai60/e-commerce_project/backend/user-service/proto"
)

type AuthHandler struct {
	userClient pb.UserServiceClient
}

func NewAuthHandler(userClient pb.UserServiceClient) *AuthHandler {
	return &AuthHandler{
		userClient: userClient,
	}
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := context.Background()
	resp, err := h.userClient.Login(ctx, &pb.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Set the cookie from the gRPC response
	if resp.Cookie != nil {
		c.SetCookie(
			resp.Cookie.Name,
			resp.Cookie.Value,
			int(resp.Cookie.MaxAge),
			resp.Cookie.Path,
			resp.Cookie.Domain,
			resp.Cookie.Secure,
			resp.Cookie.HttpOnly,
		)
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": resp.Token,
		"user":        resp.User,
	})
}

func (h *AuthHandler) handleError(c *gin.Context, err error) {
	log.Printf("Error: %v", err)
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}