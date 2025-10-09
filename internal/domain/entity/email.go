package entity

import (
	"errors"
	"regexp"
	"strings"
)

var (
	// ErrInvalidEmail is returned when email format is invalid
	ErrInvalidEmail = errors.New("invalid email format")
	// ErrEmptyEmail is returned when email is empty
	ErrEmptyEmail = errors.New("email cannot be empty")
)

// Email represents an email value object
type Email struct {
	value string
}

// emailRegex is a simple regex for email validation
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// NewEmail creates a new Email value object with validation
func NewEmail(email string) (*Email, error) {
	email = strings.TrimSpace(email)
	
	if email == "" {
		return nil, ErrEmptyEmail
	}
	
	if !emailRegex.MatchString(email) {
		return nil, ErrInvalidEmail
	}
	
	return &Email{value: strings.ToLower(email)}, nil
}

// String returns the email as a string
func (e *Email) String() string {
	return e.value
}

// Equals checks if two emails are equal
func (e *Email) Equals(other *Email) bool {
	if other == nil {
		return false
	}
	return e.value == other.value
}
