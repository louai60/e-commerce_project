package models

import (
	"errors"
)

// Common errors
var (
	ErrNotFound                = errors.New("resource not found")
	ErrAlreadyExists           = errors.New("resource already exists")
	ErrInvalidInput            = errors.New("invalid input")
	ErrInsufficientInventory   = errors.New("insufficient inventory")
	ErrReservationExpired      = errors.New("reservation expired")
	ErrReservationNotFound     = errors.New("reservation not found")
	ErrReservationInvalidState = errors.New("reservation in invalid state")
	ErrWarehouseNotFound       = errors.New("warehouse not found")
	ErrWarehouseInactive       = errors.New("warehouse is inactive")
	ErrInternalError           = errors.New("internal server error")
	ErrInvalidQuantity         = errors.New("invalid quantity")
	ErrDatabaseError           = errors.New("database error")
)
