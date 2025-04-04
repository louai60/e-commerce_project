package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/louai60/e-commerce_project/backend/user-service/models"
	pb "github.com/louai60/e-commerce_project/backend/user-service/proto"
	"github.com/louai60/e-commerce_project/backend/user-service/service"
	"go.uber.org/zap"
)

type OAuthHandler struct {
	userService  *service.UserService
	tokenManager service.TokenManager
	logger       *zap.Logger
}

func NewOAuthHandler(userService *service.UserService, tokenManager service.TokenManager, logger *zap.Logger) *OAuthHandler {
	return &OAuthHandler{
		userService:  userService,
		tokenManager: tokenManager,
		logger:       logger,
	}
}

type OAuthLoginRequest struct {
	Email            string `json:"email" validate:"required,email"`
	FirstName        string `json:"first_name"`
	LastName         string `json:"last_name"`
	Provider         string `json:"provider" validate:"required"`
	ProviderAccountId string `json:"provider_account_id" validate:"required"`
}

func (h *OAuthHandler) HandleOAuthLogin(ctx context.Context, req *OAuthLoginRequest) (*pb.LoginResponse, error) {
	h.logger.Info("Processing OAuth login request",
		zap.String("email", req.Email),
		zap.String("provider", req.Provider))

	// Create RegisterRequest for new user
	registerReq := &models.RegisterRequest{
		Email:      req.Email,
		Username:   generateUsername(req.FirstName, req.LastName),
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		UserType:   "customer",
		Role:       "user",
		Password:   "", // OAuth users don't have passwords
	}

	existingUser, err := h.userService.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			// Create new user
			user, err := h.userService.CreateUser(ctx, registerReq)
			if err != nil {
				h.logger.Error("Failed to create OAuth user", zap.Error(err))
				return nil, fmt.Errorf("failed to create user: %w", err)
			}
			existingUser = user
		} else {
			h.logger.Error("Failed to check existing user", zap.Error(err))
			return nil, fmt.Errorf("failed to check existing user: %w", err)
		}
	} else {
		// Update existing user's OAuth information if needed
		if existingUser.Provider == "" {
			existingUser.Provider = req.Provider
			existingUser.ProviderAccountId = req.ProviderAccountId
			existingUser.EmailVerified = true
			existingUser.UpdatedAt = time.Now()

			_, err = h.userService.UpdateUser(ctx, existingUser)
			if err != nil {
				h.logger.Error("Failed to update existing user with OAuth info", zap.Error(err))
				return nil, fmt.Errorf("failed to update user: %w", err)
			}
		}
	}

	// Generate tokens
	accessToken, refreshToken, refreshTokenCookie, err := h.tokenManager.GenerateTokenPair(existingUser)
	if err != nil {
		h.logger.Error("Failed to generate tokens", zap.Error(err))
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	return &pb.LoginResponse{
		Token:        accessToken,
		RefreshToken: refreshToken,
		User: &pb.User{
			UserId:      existingUser.UserID,
			Email:       existingUser.Email,
			Username:    existingUser.Username,
			FirstName:   existingUser.FirstName,
			LastName:    existingUser.LastName,
			UserType:    existingUser.UserType,
			Role:        existingUser.Role,
		},
		Cookie: &pb.CookieInfo{
			Name:     refreshTokenCookie.Name,
			Value:    refreshTokenCookie.Value,
			MaxAge:   int32(refreshTokenCookie.MaxAge),
			Path:     refreshTokenCookie.Path,
			Domain:   refreshTokenCookie.Domain,
			Secure:   refreshTokenCookie.Secure,
			HttpOnly: refreshTokenCookie.HttpOnly,
		},
	}, nil
}

func generateUsername(firstName, lastName string) string {
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	return fmt.Sprintf("%s%s%d", 
		sanitizeUsername(firstName), 
		sanitizeUsername(lastName), 
		timestamp%1000)
}

func sanitizeUsername(s string) string {
	if len(s) > 0 {
		return strings.ToLower(strings.Map(func(r rune) rune {
			if unicode.IsLetter(r) || unicode.IsNumber(r) {
				return r
			}
			return -1
		}, s))
	}
	return ""
}

