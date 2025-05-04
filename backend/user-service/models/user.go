package models

import (
	"database/sql"
	"github.com/google/uuid"
	"time"
)

const (
	// User Roles
	RoleUser  = "user"
	RoleAdmin = "admin"
	// RoleSuperAdmin is now defined in roles.go
)

type User struct {
	UserID         uuid.UUID    `json:"user_id" db:"user_id"`
	Email          string       `json:"email" db:"email"`
	Username       string       `json:"username" db:"username"`
	HashedPassword string       `json:"-" db:"hashed_password"`
	FirstName      string       `json:"first_name" db:"first_name"`
	LastName       string       `json:"last_name" db:"last_name"`
	PhoneNumber    string       `json:"phone_number" db:"phone_number"`
	UserType       string       `json:"user_type" db:"user_type"`
	Role           string       `json:"role" db:"role"`
	AccountStatus  string       `json:"account_status" db:"account_status"`
	EmailVerified  bool         `json:"email_verified" db:"email_verified"`
	PhoneVerified  bool         `json:"phone_verified" db:"phone_verified"`
	CreatedAt      time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at" db:"updated_at"`
	LastLogin      sql.NullTime `json:"last_login" db:"last_login"` // Changed to sql.NullTime
	RefreshTokenID string       `json:"-" db:"refresh_token_id"`    // JTI of the current valid refresh token
}

type UserAddress struct {
	AddressID      uuid.UUID `json:"address_id" db:"address_id"`
	UserID         uuid.UUID `json:"user_id" db:"user_id"`
	AddressType    string    `json:"address_type" db:"address_type"`
	StreetAddress1 string    `json:"street_address1" db:"street_address1"`
	StreetAddress2 string    `json:"street_address2" db:"street_address2"`
	City           string    `json:"city" db:"city"`
	State          string    `json:"state" db:"state"`
	PostalCode     string    `json:"postal_code" db:"postal_code"`
	Country        string    `json:"country" db:"country"`
	IsDefault      bool      `json:"is_default" db:"is_default"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

type PaymentMethod struct {
	PaymentMethodID  uuid.UUID  `json:"payment_method_id" db:"payment_method_id"`
	UserID           uuid.UUID  `json:"user_id" db:"user_id"`
	PaymentType      string     `json:"payment_type" db:"payment_type"`
	CardLastFour     string     `json:"card_last_four,omitempty" db:"card_last_four"`
	CardBrand        string     `json:"card_brand,omitempty" db:"card_brand"`
	ExpirationMonth  int16      `json:"expiration_month,omitempty" db:"expiration_month"`
	ExpirationYear   int16      `json:"expiration_year,omitempty" db:"expiration_year"`
	IsDefault        bool       `json:"is_default" db:"is_default"`
	BillingAddressID *uuid.UUID `json:"billing_address_id,omitempty" db:"billing_address_id"`
	Token            string     `json:"token" db:"token"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
}

type UserPreferences struct {
	PreferenceID      uuid.UUID `json:"preference_id" db:"preference_id"`
	UserID            uuid.UUID `json:"user_id" db:"user_id"`
	Language          string    `json:"language" db:"language"`
	Currency          string    `json:"currency" db:"currency"`
	NotificationEmail bool      `json:"notification_email" db:"notification_email"`
	NotificationSMS   bool      `json:"notification_sms" db:"notification_sms"`
	Theme             string    `json:"theme" db:"theme"`
	Timezone          string    `json:"timezone" db:"timezone"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

type RegisterRequest struct {
	Email       string `json:"email" validate:"required,email"`
	Username    string `json:"username" validate:"required,min=3,max=50"`
	Password    string `json:"password" validate:"required,min=8"`
	FirstName   string `json:"first_name" validate:"required"`
	LastName    string `json:"last_name" validate:"required"`
	UserType    string `json:"user_type" validate:"required,oneof=customer seller admin"`
	Role        string `json:"role"`
	PhoneNumber string `json:"phone_number"`
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

// HasPermission checks if the user has a specific permission
func (u *User) HasPermission(permission Permission) bool {
	permissions, exists := RolePermissions[u.Role]
	if !exists {
		return false
	}

	// Super admin has all permissions
	if u.Role == RoleSuperAdmin {
		return true
	}

	for _, p := range permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// IsValidUserType validates the user type
func IsValidUserType(userType string) bool {
	validTypes := []string{UserTypeCustomer, UserTypeSeller, UserTypeAdmin}
	for _, t := range validTypes {
		if t == userType {
			return true
		}
	}
	return false
}

// IsValidRole validates the role
func IsValidRole(userType, role string) bool {
	if userType == UserTypeAdmin && (role == RoleAdmin || role == RoleSuperAdmin) {
		return true
	}
	if userType == UserTypeCustomer && role == RoleUser {
		return true
	}
	if userType == UserTypeSeller && (role == "basic_seller" || role == "verified_seller") {
		return true
	}
	return false
}
