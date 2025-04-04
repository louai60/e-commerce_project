package handlers

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	pb "github.com/louai60/e-commerce_project/backend/user-service/proto"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/codes"
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

type OAuthLoginRequest struct {
	Email             string `json:"email"`
	FirstName         string `json:"first_name"`
	LastName          string `json:"last_name"`
	Provider          string `json:"provider"`
	ProviderAccountId string `json:"provider_account_id"`
}

func (h *AuthHandler) OAuthLogin(c *gin.Context) {
	var req OAuthLoginRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := context.Background()
	resp, err := h.userClient.OAuthLogin(ctx, &pb.OAuthLoginRequest{
		Email:             req.Email,
		FirstName:         req.FirstName, 
		LastName:          req.LastName,
		Provider:          req.Provider,
		ProviderAccountId: req.ProviderAccountId,
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
		"access_token":  resp.Token,        // Changed from AccessToken to Token
		"refresh_token": resp.RefreshToken,
		"user":         resp.User,
	})
}

func (h *AuthHandler) handleError(c *gin.Context, err error) {
	log.Printf("Handling gRPC error: %v", err)

	// Default error response
	httpStatus := http.StatusInternalServerError
	errorMessage := "An internal server error occurred"

	// Check if it's a gRPC status error
	if st, ok := status.FromError(err); ok {
		errorMessage = st.Message() // Use the message from the gRPC error

		// Map gRPC codes to HTTP status codes
		switch st.Code() {
		case codes.Unauthenticated:
			httpStatus = http.StatusUnauthorized
		case codes.NotFound:
			httpStatus = http.StatusNotFound
		case codes.InvalidArgument:
			httpStatus = http.StatusBadRequest
		case codes.AlreadyExists:
			httpStatus = http.StatusConflict
		// Add other mappings as needed
		default:
			httpStatus = http.StatusInternalServerError
		}
	} else {
		// Handle non-gRPC errors if necessary, or keep the default
		log.Printf("Non-gRPC error encountered: %v", err)
		// Optionally, you could use err.Error() here, but it might expose internal details
	}

	c.JSON(httpStatus, gin.H{"error": errorMessage})
}
