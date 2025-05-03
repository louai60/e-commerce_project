package service

import (
	"errors"
	"fmt"
	"github.com/nbutton23/zxcvbn-go"
)

type PasswordValidator struct {
	minLength int
}

func NewPasswordValidator() *PasswordValidator {
	return &PasswordValidator{
		minLength: 12,
	}
}

func (v *PasswordValidator) Validate(password string) error {
	if len(password) < v.minLength {
		return fmt.Errorf("password must be at least %d characters", v.minLength)
	}

	strength := zxcvbn.PasswordStrength(password, nil)
	if strength.Score < 3 {
		return errors.New("password is too weak. Include a mix of uppercase, lowercase, numbers, and symbols")
	}

	return nil
}
