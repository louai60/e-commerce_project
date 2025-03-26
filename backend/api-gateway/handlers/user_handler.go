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

func (h *UserHandler) Register(c *gin.Context) {
    var req struct {
        Email     string `json:"email" binding:"required,email"`
        Username  string `json:"username" binding:"required,min=3,max=50"`
        Password  string `json:"password" binding:"required,min=8"`
        FirstName string `json:"first_name" binding:"required"`
        LastName  string `json:"last_name" binding:"required"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        h.logger.Error("Invalid request payload", zap.Error(err))
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    h.logger.Info("Attempting to register user",
        zap.String("email", req.Email),
        zap.String("username", req.Username),
    )

    ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
    defer cancel()

    grpcReq := &pb.CreateUserRequest{
        Email:     req.Email,
        Username:  req.Username,
        Password:  req.Password,
        FirstName: req.FirstName,
        LastName:  req.LastName,
    }

    resp, err := h.client.CreateUser(ctx, grpcReq)
    if err != nil {
        st := status.Convert(err)
        h.logger.Error("Failed to create user",
            zap.Error(err),
            zap.String("code", st.Code().String()),
            zap.String("message", st.Message()),
        )

        switch st.Code() {
        case codes.AlreadyExists:
            c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
        case codes.InvalidArgument:
            c.JSON(http.StatusBadRequest, gin.H{"error": st.Message()})
        case codes.Unavailable:
            c.JSON(http.StatusServiceUnavailable, gin.H{"error": "User service is unavailable"})
        default:
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user: " + st.Message()})
        }
        return
    }

    c.JSON(http.StatusCreated, resp)
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

    h.logger.Info("Processing login request",
        zap.String("email", req.Email))

    ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
    defer cancel()

    loginReq := &pb.LoginRequest{
        Email:    req.Email,
        Password: req.Password,
    }

    resp, err := h.client.Login(ctx, loginReq)
    if err != nil {
        st := status.Convert(err)
        h.logger.Error("Login failed",
            zap.Error(err),
            zap.String("code", st.Code().String()),
            zap.String("message", st.Message()),
            zap.String("email", req.Email),
        )

        switch st.Code() {
        case codes.Unauthenticated:
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
        case codes.ResourceExhausted:
            c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many login attempts"})
        default:
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process login"})
        }
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "token": resp.Token,
        "user": resp.User,
    })
}

func (h *UserHandler) GetProfile(c *gin.Context) {
    userID := c.GetString("user_id")
    resp, err := h.client.GetUser(c.Request.Context(), &pb.GetUserRequest{Id: userID})
    if err != nil {
        h.logger.Error("Failed to get user profile", zap.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get profile"})
        return
    }

    c.JSON(http.StatusOK, resp)
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
    var req pb.UpdateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    req.Id = c.GetString("user_id")
    resp, err := h.client.UpdateUser(c.Request.Context(), &req)
    if err != nil {
        h.logger.Error("Failed to update user profile", zap.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
        return
    }

    c.JSON(http.StatusOK, resp)
}

func (h *UserHandler) ListUsers(c *gin.Context) {
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

    resp, err := h.client.ListUsers(c.Request.Context(), &pb.ListUsersRequest{
        Page:  int32(page),
        Limit: int32(limit),
    })
    if err != nil {
        h.logger.Error("Failed to list users", zap.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list users"})
        return
    }

    c.JSON(http.StatusOK, resp)
}

func (h *UserHandler) GetUser(c *gin.Context) {
    userID := c.Param("id")
    resp, err := h.client.GetUser(c.Request.Context(), &pb.GetUserRequest{Id: userID})
    if err != nil {
        h.logger.Error("Failed to get user", zap.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
        return
    }

    c.JSON(http.StatusOK, resp)
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
    userID := c.Param("id")
    _, err := h.client.DeleteUser(c.Request.Context(), &pb.DeleteUserRequest{Id: userID})
    if err != nil {
        h.logger.Error("Failed to delete user", zap.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

func (h *UserHandler) CreateAdmin(c *gin.Context) {
    var req struct {
        Email     string `json:"email" binding:"required,email"`
        Username  string `json:"username" binding:"required,min=3,max=50"`
        Password  string `json:"password" binding:"required,min=8"`
        FirstName string `json:"first_name" binding:"required"`
        LastName  string `json:"last_name" binding:"required"`
        AdminKey  string `json:"admin_key" binding:"required"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        h.logger.Error("Invalid request payload", zap.Error(err))
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Debug logging
    expectedAdminKey := os.Getenv("ADMIN_CREATE_KEY")
    h.logger.Debug("Admin creation attempt",
        zap.String("provided_key", req.AdminKey),
        zap.String("expected_key", expectedAdminKey))

    if req.AdminKey != expectedAdminKey {
        h.logger.Error("Invalid admin key provided",
            zap.String("provided_key", req.AdminKey),
            zap.String("expected_key", expectedAdminKey))
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid admin key"})
        return
    }

    ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
    defer cancel()

    grpcReq := &pb.CreateUserRequest{
        Email:     req.Email,
        Username:  req.Username,
        Password:  req.Password,
        FirstName: req.FirstName,
        LastName:  req.LastName,
        Role:      "admin",  // Explicitly set the role to "admin"
    }

    resp, err := h.client.CreateUser(ctx, grpcReq)
    if err != nil {
        st := status.Convert(err)
        h.logger.Error("Failed to create admin user",
            zap.Error(err),
            zap.String("code", st.Code().String()),
        )
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create admin user"})
        return
    }

    c.JSON(http.StatusCreated, resp)
}




