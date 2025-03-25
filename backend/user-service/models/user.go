package models

import (
    "time"
)

type User struct {
    ID         string    `json:"id"`
    Email      string    `json:"email"`
    Username   string    `json:"username"`
    Password   string    `json:"password,omitempty"`
    FirstName  string    `json:"first_name"`
    LastName   string    `json:"last_name"`
    Role       string    `json:"role"`
    CreatedAt  time.Time `json:"created_at"`
    UpdatedAt  time.Time `json:"updated_at"`
    LastLogin  *time.Time `json:"last_login,omitempty"`
    IsActive   bool      `json:"is_active"`
    IsVerified bool      `json:"is_verified"`
}

type RegisterRequest struct {
    Email     string `json:"email" validate:"required,email"`
    Username  string `json:"username" validate:"required,min=3,max=50"`
    Password  string `json:"password" validate:"required,min=8"`
    FirstName string `json:"first_name" validate:"required"`
    LastName  string `json:"last_name" validate:"required"`
}

type LoginCredentials struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

type LoginResponse struct {
    Token        string `json:"token"`
    RefreshToken string `json:"refresh_token"`
    User         *User  `json:"user"`
}

// type LoginResponse struct {
//     Token        string `json:"token"`
//     RefreshToken string `json:"refresh_token"`
//     User         *User  `json:"user"`
// }






