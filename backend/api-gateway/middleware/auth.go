package middleware

import (
	"crypto/rsa"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

// Global variable to hold the parsed public key
var jwtPublicKey *rsa.PublicKey

// LoadPublicKey loads the JWT public key from the specified path.
// It should be called once during application startup.
func LoadPublicKey() error {
	publicKeyPath := os.Getenv("JWT_PUBLIC_KEY_PATH")
	if publicKeyPath == "" {
		publicKeyPath = "certificates/public_key.pem" // Default path
	}

	keyBytes, err := ioutil.ReadFile(publicKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read public key file '%s': %w", publicKeyPath, err)
	}

	jwtPublicKey, err = jwt.ParseRSAPublicKeyFromPEM(keyBytes)
	if err != nil {
		return fmt.Errorf("failed to parse public key: %w", err)
	}
	log.Printf("Successfully loaded JWT public key from %s", publicKeyPath)
	return nil
}


func AuthRequired() gin.HandlerFunc {
	// Ensure the public key is loaded before returning the handler
	if jwtPublicKey == nil {
		// This should ideally not happen if LoadPublicKey is called at startup
		log.Fatal("JWT Public Key not loaded. Call LoadPublicKey() during initialization.")
		// Or return a handler that always returns an error
		// return func(c *gin.Context) {
		// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "JWT public key not configured"})
		// 	c.Abort()
		// }
	}

	return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header is required"})
            c.Abort()
            return
        }

        bearerToken := strings.Split(authHeader, " ")
        if len(bearerToken) != 2 || bearerToken[0] != "Bearer" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
            c.Abort()
            return
        }

        token := bearerToken[1]
  claims, err := validateToken(token, jwtPublicKey) // Pass the loaded key
  if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": fmt.Sprintf("invalid token: %v", err)})
            c.Abort()
            return
        }

        // Set user information in context
        c.Set("user_id", claims["user_id"])
        c.Set("user_role", claims["role"])
        c.Set("user_email", claims["email"])

        c.Next()
    }
}

func validateToken(tokenString string, publicKey *rsa.PublicKey) (jwt.MapClaims, error) {
	if publicKey == nil {
		return nil, fmt.Errorf("public key is nil, cannot validate token")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the alg is RS256
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})

	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorExpired != 0 {
				return nil, fmt.Errorf("token has expired")
			} else if ve.Errors&jwt.ValidationErrorNotValidYet != 0 {
				return nil, fmt.Errorf("token not active yet")
			}
		}
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Double-check expiration explicitly
	if exp, ok := claims["exp"].(float64); !ok || time.Now().Unix() > int64(exp) {
		return nil, fmt.Errorf("token has expired")
	}

	// Ensure it's an access token
	tokenType, ok := claims["type"].(string)
	if !ok || tokenType != "access" {
		return nil, fmt.Errorf("invalid token type: expected 'access'")
	}

	// Validate required claims for an access token
	requiredClaims := []string{"user_id", "email", "username", "user_type", "role"}
	for _, claim := range requiredClaims {
		if claims[claim] == nil {
			return nil, fmt.Errorf("missing required claim: %s", claim)
		}
	}

	// Modified user_id validation to handle string UUID
	if userID, ok := claims["user_id"].(string); !ok {
		return nil, fmt.Errorf("invalid type for user_id claim: expected string UUID")
	} else {
		// Validate that it's a valid UUID
		if _, err := uuid.Parse(userID); err != nil {
			return nil, fmt.Errorf("invalid user_id format: not a valid UUID")
		}
	}

	return claims, nil
}
