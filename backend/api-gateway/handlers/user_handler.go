package handlers

import (
    "net/http"
    "time"
    "context"
    "os"
    "strconv"

    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    "google.golang.org/grpc/status"
    "google.golang.org/grpc/codes"
    pb "github.com/louai60/e-commerce_project/backend/user-service/proto"
)

type UserHandler struct {
    client pb.UserServiceClient
    logger *zap.Logger
}

// Request structs
type CreateUserRequest struct {
    Email     string `json:"Email" binding:"required,email"`
    Password  string `json:"Password" binding:"required,min=8"`
    FirstName string `json:"FirstName" binding:"required"`
    LastName  string `json:"LastName" binding:"required"`
}

type UpdateUserRequest struct {
    Email       string `json:"Email"`
    Username    string `json:"Username"`
    FirstName   string `json:"FirstName"`
    LastName    string `json:"LastName"`
    PhoneNumber string `json:"PhoneNumber"`
}

type AddressRequest struct {
    AddressType    string `json:"address_type" binding:"required"`
    StreetAddress1 string `json:"street_address1" binding:"required"`
    StreetAddress2 string `json:"street_address2"`
    City          string `json:"city" binding:"required"`
    State         string `json:"state" binding:"required"`
    PostalCode    string `json:"postal_code" binding:"required"`
    Country       string `json:"country" binding:"required"`
    IsDefault     bool   `json:"is_default"`
}

type PaymentMethodRequest struct {
    PaymentType     string `json:"payment_type" binding:"required"`
    CardLastFour    string `json:"card_last_four"`
    CardBrand       string `json:"card_brand"`
    ExpirationMonth int32  `json:"expiration_month"`
    ExpirationYear  int32  `json:"expiration_year"`
    IsDefault       bool   `json:"is_default"`
    Token          string `json:"token" binding:"required"`
}

func NewUserHandler(userServiceAddr string, logger *zap.Logger) (*UserHandler, error) {
    conn, err := grpc.Dial(userServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        return nil, err
    }

    client := pb.NewUserServiceClient(conn)
    return &UserHandler{
        client: client,
        logger: logger,
    }, nil
}

// Helper function to parse user IDs
func (h *UserHandler) parseUserID(idStr string) (string, error) {
    // First parse as int64 to validate format
    _, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil {
        h.logger.Error("Invalid user ID format", 
            zap.String("user_id", idStr),
            zap.Error(err))
        return "", err
    }
    return idStr, nil
}

func (h *UserHandler) Register(c *gin.Context) {
    var req CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        h.logger.Error("Request validation failed", zap.Error(err))
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
    defer cancel()

    // Create the gRPC request including UserType and Role from the HTTP request
    grpcReq := &pb.CreateUserRequest{
    	Email:     req.Email,
    	Password:  req.Password,
    	FirstName: req.FirstName,
    	LastName:  req.LastName,
    	UserType:  "customer", 
    	Role:      "user",     
    }

    resp, err := h.client.CreateUser(ctx, grpcReq)
    if err != nil {
        st, ok := status.FromError(err)
        if ok {
            statusCode := http.StatusInternalServerError
            switch st.Code() {
            case codes.AlreadyExists:
                statusCode = http.StatusConflict
                c.JSON(statusCode, gin.H{"error": st.Message()})
                return
            case codes.InvalidArgument:
                statusCode = http.StatusBadRequest
                c.JSON(statusCode, gin.H{"error": st.Message()})
                return
            }
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, gin.H{
        "user": resp.User,
        "message": "User registered successfully",
    })
}

func (h *UserHandler) GetProfile(c *gin.Context) {
    userIDStr := c.GetString("user_id")
    userID, err := h.parseUserID(userIDStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
        return
    }

    resp, err := h.client.GetUser(c.Request.Context(), &pb.GetUserRequest{UserId: userID})
    if err != nil {
        h.handleGRPCError(c, err, "Failed to get user profile")
        return
    }

    c.JSON(http.StatusOK, resp)
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
    userIDStr := c.GetString("user_id")
    userID, err := h.parseUserID(userIDStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
        return
    }

    var req UpdateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    grpcReq := &pb.UpdateUserRequest{
        UserId:      userID,
        Username:    req.Username,
        FirstName:   req.FirstName,
        LastName:    req.LastName,
        PhoneNumber: req.PhoneNumber,
    }

    resp, err := h.client.UpdateUser(c.Request.Context(), grpcReq)
    if err != nil {
        h.handleGRPCError(c, err, "Failed to update profile")
        return
    }

    c.JSON(http.StatusOK, resp)
}

func (h *UserHandler) ListUsers(c *gin.Context) {
    page, err := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 32)
    if err != nil || page < 1 {
        page = 1
    }

    limit, err := strconv.ParseInt(c.DefaultQuery("limit", "10"), 10, 32)
    if err != nil || limit < 1 {
        limit = 10
    }

    resp, err := h.client.ListUsers(c.Request.Context(), &pb.ListUsersRequest{
        Page:  int32(page),
        Limit: int32(limit),
    })
    if err != nil {
        h.handleGRPCError(c, err, "Failed to list users")
        return
    }

    c.JSON(http.StatusOK, resp)
}

func (h *UserHandler) GetUser(c *gin.Context) {
    userIDStr := c.Param("id")
    userID, err := h.parseUserID(userIDStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
        return
    }
    
    resp, err := h.client.GetUser(c.Request.Context(), &pb.GetUserRequest{
        UserId: userID,
    })
    if err != nil {
        h.handleGRPCError(c, err, "Failed to get user")
        return
    }

    c.JSON(http.StatusOK, resp)
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
    userIDStr := c.Param("id")
    userID, err := h.parseUserID(userIDStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
        return
    }
    
    _, err = h.client.DeleteUser(c.Request.Context(), &pb.DeleteUserRequest{UserId: userID})
    if err != nil {
        h.handleGRPCError(c, err, "Failed to delete user")
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

func (h *UserHandler) handleGRPCError(c *gin.Context, err error, defaultMsg string) {
    st, ok := status.FromError(err)
    if !ok {
        h.logger.Error(defaultMsg, zap.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{"error": defaultMsg})
        return
    }

    h.logger.Error(st.Message(),
        zap.String("code", st.Code().String()),
        zap.Error(err))

    switch st.Code() {
    case codes.NotFound:
        c.JSON(http.StatusNotFound, gin.H{"error": st.Message()})
    case codes.AlreadyExists:
        c.JSON(http.StatusConflict, gin.H{"error": st.Message()})
    case codes.InvalidArgument:
        c.JSON(http.StatusBadRequest, gin.H{"error": st.Message()})
    case codes.Unauthenticated:
        c.JSON(http.StatusUnauthorized, gin.H{"error": st.Message()})
    case codes.PermissionDenied:
        c.JSON(http.StatusForbidden, gin.H{"error": st.Message()})
    case codes.Unavailable:
        c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Service unavailable"})
    default:
        c.JSON(http.StatusInternalServerError, gin.H{"error": defaultMsg})
    }
}

func (h *UserHandler) Login(c *gin.Context) {
    var req struct {
        Email    string `json:"email" binding:"required,email"`
        Password string `json:"password" binding:"required"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        h.logger.Error("Invalid login payload", zap.Error(err))
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    h.logger.Info("Login attempt", 
        zap.String("email", req.Email))

    ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
    defer cancel()

    resp, err := h.client.Login(ctx, &pb.LoginRequest{
        Email:    req.Email,
        Password: req.Password,
    })

    if err != nil {
        h.logger.Error("Login failed", 
            zap.String("email", req.Email),
            zap.Error(err))
        h.handleGRPCError(c, err, "Login failed")
        return
    }

    // Set the refresh token cookie if provided
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
   
    // Return access token and user details in the response body
    // Refresh token is handled via HttpOnly cookie
    c.JSON(http.StatusOK, gin.H{
    	"access_token": resp.Token, // Keep only one access_token key
        "user":         resp.User,
    })
}

func (h *UserHandler) Logout(c *gin.Context) {
	// Clear the refresh token cookie by setting its MaxAge to -1
	// Use the same attributes (Path, Domain, Secure, HttpOnly) as when setting it
	// Assuming the cookie name is "refresh_token" and path is "/" based on common practice
	// TODO: Confirm cookie name and path from user-service if different
	c.SetCookie(
		"refresh_token", // Cookie name
		"",              // Empty value
		-1,              // MaxAge = -1 deletes the cookie
		"/api/v1/users/refresh", // Path must match the one used when setting the cookie
		"",              // Domain (leave empty for default)
		true,            // Secure flag (should match setting)
		true,            // HttpOnly flag (should match setting)
	)

	h.logger.Info("User logged out, refresh token cookie cleared")
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

func (h *UserHandler) CreateAdmin(c *gin.Context) {
    var req CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    adminKey := c.GetHeader("X-Admin-Key")
    if adminKey != os.Getenv("ADMIN_CREATE_KEY") {
        h.logger.Error("Invalid admin key provided")
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid admin key"})
        return
    }

    ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
    defer cancel()

    // Create the gRPC request including UserType and Role from the HTTP request
    // The user-service will handle setting UserType and Role to 'admin' internally
    grpcReq := &pb.CreateUserRequest{
    	Email:     req.Email,
    	Password:  req.Password,
    	FirstName: req.FirstName,
    	LastName:  req.LastName,
        UserType: "admin",
        Role: "admin",
    }

    resp, err := h.client.CreateUser(ctx, grpcReq)
    if err != nil {
        h.handleGRPCError(c, err, "Failed to create admin user")
        return
    }

    c.JSON(http.StatusCreated, resp)
}

// New methods for address management
func (h *UserHandler) AddAddress(c *gin.Context) {
    userIDStr := c.GetString("user_id")
    userID, err := h.parseUserID(userIDStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
        return
    }

    var req AddressRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    grpcReq := &pb.AddAddressRequest{
        UserId:         userID,
        AddressType:    req.AddressType,
        StreetAddress1: req.StreetAddress1,
        StreetAddress2: req.StreetAddress2,
        City:           req.City,
        State:          req.State,
        PostalCode:     req.PostalCode,
        Country:        req.Country,
        IsDefault:      req.IsDefault,
    }

    resp, err := h.client.AddAddress(c.Request.Context(), grpcReq)
    if err != nil {
        h.handleGRPCError(c, err, "Failed to add address")
        return
    }

    c.JSON(http.StatusCreated, resp)
}

// GetAddresses retrieves all addresses for the current user
func (h *UserHandler) GetAddresses(c *gin.Context) {
    userIDStr := c.GetString("user_id")
    userID, err := h.parseUserID(userIDStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
        return
    }

    resp, err := h.client.GetAddresses(c.Request.Context(), &pb.GetAddressesRequest{
        UserId: userID,
    })
    if err != nil {
        h.handleGRPCError(c, err, "Failed to retrieve addresses")
        return
    }

    c.JSON(http.StatusOK, resp)
}

// UpdateAddress updates an existing address
func (h *UserHandler) UpdateAddress(c *gin.Context) {
    userIDStr := c.GetString("user_id")
    userID, err := h.parseUserID(userIDStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
        return
    }

    addressIDStr := c.Param("addressID")
    addressID, err := h.parseUserID(addressIDStr) // Reusing parseUserID since format is the same
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid address ID format"})
        return
    }

    var req AddressRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    grpcReq := &pb.UpdateAddressRequest{
        UserId:         userID,
        AddressId:      addressID,
        AddressType:    req.AddressType,
        StreetAddress1: req.StreetAddress1,
        StreetAddress2: req.StreetAddress2,
        City:           req.City,
        State:          req.State,
        PostalCode:     req.PostalCode,
        Country:        req.Country,
        IsDefault:      req.IsDefault,
    }

    resp, err := h.client.UpdateAddress(c.Request.Context(), grpcReq)
    if err != nil {
        h.handleGRPCError(c, err, "Failed to update address")
        return
    }

    c.JSON(http.StatusOK, resp)
}

// DeleteAddress removes an existing address
func (h *UserHandler) DeleteAddress(c *gin.Context) {
    userIDStr := c.GetString("user_id")
    userID, err := h.parseUserID(userIDStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
        return
    }

    addressIDStr := c.Param("addressID")
    addressID, err := h.parseUserID(addressIDStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid address ID format"})
        return
    }

    grpcReq := &pb.DeleteAddressRequest{
        UserId:    userID,
        AddressId: addressID,
    }

    _, err = h.client.DeleteAddress(c.Request.Context(), grpcReq)
    if err != nil {
        h.handleGRPCError(c, err, "Failed to delete address")
        return
    }

    c.Status(http.StatusNoContent)
}


// New methods for payment management
func (h *UserHandler) AddPaymentMethod(c *gin.Context) {
    userIDStr := c.GetString("user_id")
    userID, err := strconv.ParseInt(userIDStr, 10, 64)
    if err != nil {
        h.logger.Error("Invalid user ID format", 
            zap.String("user_id", userIDStr),
            zap.Error(err))
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
        return
    }

    var req PaymentMethodRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
    defer cancel()

    grpcReq := &pb.AddPaymentMethodRequest{
        UserId:          strconv.FormatInt(userID, 10),
        PaymentType:     req.PaymentType,
        CardLastFour:    req.CardLastFour,
        CardBrand:       req.CardBrand,
        ExpirationMonth: req.ExpirationMonth,
        ExpirationYear:  req.ExpirationYear,
        IsDefault:       req.IsDefault,
        Token:           req.Token,
    }

    resp, err := h.client.AddPaymentMethod(ctx, grpcReq)
    if err != nil {
        h.handleGRPCError(c, err, "Failed to add payment method")
        return
    }

    c.JSON(http.StatusCreated, resp)
}

func (h *UserHandler) RefreshToken(c *gin.Context) {
    // Read the refresh token from the cookie
    refreshToken, err := c.Cookie("refresh_token")
    if err != nil {
        h.logger.Error("Refresh token cookie not found", zap.Error(err))
        c.JSON(http.StatusBadRequest, gin.H{"error": "Refresh token cookie not found"})
        return
    }
    if refreshToken == "" {
        h.logger.Error("Refresh token cookie is empty")
        c.JSON(http.StatusBadRequest, gin.H{"error": "Refresh token cookie is empty"})
        return
    }

    h.logger.Info("Attempting token refresh via cookie")

    ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
    defer cancel()

    resp, err := h.client.RefreshToken(ctx, &pb.RefreshTokenRequest{
        RefreshToken: refreshToken, // Use token from cookie
    })
    if err != nil {
        h.handleGRPCError(c, err, "Failed to refresh token")
        return
    }

    // Set the refresh token cookie
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

    // Only return the new access token and user details in the body.
    // The new refresh token is handled by the HttpOnly cookie set above.
    c.JSON(http.StatusOK, gin.H{
        "access_token": resp.Token,
        "user":         resp.User, // Include user details if needed by frontend after refresh
    })
}
