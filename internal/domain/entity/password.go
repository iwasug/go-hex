package entity

import (
	"errors"
	"strings"
	"unicode"
)

var (
	// ErrPasswordTooShort is returned when password is too short
	ErrPasswordTooShort = errors.New("password must be at least 8 characters long")
	// ErrPasswordTooWeak is returned when password doesn't meet complexity requirements
	ErrPasswordTooWeak = errors.New("password must contain at least one uppercase letter, one lowercase letter, one number, and one special character")
	// ErrEmptyPassword is returned when password is empty
	ErrEmptyPassword = errors.New("password cannot be empty")
)

// Password represents a password value object
type Password struct {
	value string
}

// NewPassword creates a new Password value object with validation
func NewPassword(password string) (*Password, error) {
	password = strings.TrimSpace(password)
	
	if password == "" {
		return nil, ErrEmptyPassword
	}
	
	if len(password) < 8 {
		return nil, ErrPasswordTooShort
	}
	
	if !isPasswordStrong(password) {
		return nil, ErrPasswordTooWeak
	}
	
	return &Password{value: password}, nil
}

// String returns the password as a string
func (p *Password) String() string {
	return p.value
}

// isPasswordStrong checks if password meets complexity requirements
func isPasswordStrong(password string) bool {
	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)
	
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}
	
	return hasUpper && hasLower && hasNumber && hasSpecial
}
