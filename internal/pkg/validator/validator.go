package validator

import (
	"fmt"
	"regexp"
	"strings"
)

// ValidationError represents a single validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrors represents a collection of validation errors
type ValidationErrors struct {
	Errors []ValidationError `json:"errors"`
}

// Error implements the error interface
func (v *ValidationErrors) Error() string {
	if len(v.Errors) == 0 {
		return "validation failed"
	}

	var messages []string
	for _, err := range v.Errors {
		messages = append(messages, fmt.Sprintf("%s: %s", err.Field, err.Message))
	}
	return strings.Join(messages, "; ")
}

// HasErrors returns true if there are validation errors
func (v *ValidationErrors) HasErrors() bool {
	return len(v.Errors) > 0
}

// Add adds a validation error
func (v *ValidationErrors) Add(field, message string) {
	v.Errors = append(v.Errors, ValidationError{
		Field:   field,
		Message: message,
	})
}

// Validator provides validation methods
type Validator struct {
	errors ValidationErrors
}

// New creates a new Validator instance
func New() *Validator {
	return &Validator{
		errors: ValidationErrors{
			Errors: make([]ValidationError, 0),
		},
	}
}

// Email validation regex pattern
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// ValidateEmail validates an email address
func (v *Validator) ValidateEmail(field, email string) {
	if email == "" {
		v.errors.Add(field, "email is required")
		return
	}

	if len(email) > 255 {
		v.errors.Add(field, "email must not exceed 255 characters")
		return
	}

	if !emailRegex.MatchString(email) {
		v.errors.Add(field, "email format is invalid")
	}
}

// ValidatePassword validates password strength
func (v *Validator) ValidatePassword(field, password string) {
	if password == "" {
		v.errors.Add(field, "password is required")
		return
	}

	if len(password) < 8 {
		v.errors.Add(field, "password must be at least 8 characters long")
		return
	}

	if len(password) > 72 {
		v.errors.Add(field, "password must not exceed 72 characters")
		return
	}

	// Check for at least one uppercase letter
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	if !hasUpper {
		v.errors.Add(field, "password must contain at least one uppercase letter")
	}

	// Check for at least one lowercase letter
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	if !hasLower {
		v.errors.Add(field, "password must contain at least one lowercase letter")
	}

	// Check for at least one digit
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(password)
	if !hasDigit {
		v.errors.Add(field, "password must contain at least one digit")
	}
}

// ValidateRequired validates that a field is not empty
func (v *Validator) ValidateRequired(field, value string) {
	if strings.TrimSpace(value) == "" {
		v.errors.Add(field, fmt.Sprintf("%s is required", field))
	}
}

// ValidateEnum validates that a value is in the allowed list
func (v *Validator) ValidateEnum(field, value string, allowedValues []string) {
	if value == "" {
		return // Empty values are handled by ValidateRequired if needed
	}

	for _, allowed := range allowedValues {
		if value == allowed {
			return
		}
	}

	v.errors.Add(field, fmt.Sprintf("%s must be one of: %s", field, strings.Join(allowedValues, ", ")))
}

// ValidateOrderBy validates order_by field against whitelist
func (v *Validator) ValidateOrderBy(field, value string, allowedFields []string) {
	if value == "" {
		return // Empty values will use default
	}

	v.ValidateEnum(field, value, allowedFields)
}

// ValidateOrderDirection validates order direction (asc or desc)
func (v *Validator) ValidateOrderDirection(field, value string) {
	if value == "" {
		return // Empty values will use default
	}

	v.ValidateEnum(field, value, []string{"asc", "desc"})
}

// ValidateMinLength validates minimum string length
func (v *Validator) ValidateMinLength(field, value string, minLength int) {
	if len(value) < minLength {
		v.errors.Add(field, fmt.Sprintf("%s must be at least %d characters long", field, minLength))
	}
}

// ValidateMaxLength validates maximum string length
func (v *Validator) ValidateMaxLength(field, value string, maxLength int) {
	if len(value) > maxLength {
		v.errors.Add(field, fmt.Sprintf("%s must not exceed %d characters", field, maxLength))
	}
}

// ValidateRange validates that an integer is within a range
func (v *Validator) ValidateRange(field string, value, min, max int) {
	if value < min || value > max {
		v.errors.Add(field, fmt.Sprintf("%s must be between %d and %d", field, min, max))
	}
}

// ValidatePositive validates that an integer is positive
func (v *Validator) ValidatePositive(field string, value int) {
	if value <= 0 {
		v.errors.Add(field, fmt.Sprintf("%s must be a positive number", field))
	}
}

// Errors returns the validation errors
func (v *Validator) Errors() *ValidationErrors {
	if len(v.errors.Errors) == 0 {
		return nil
	}
	return &v.errors
}

// HasErrors returns true if there are validation errors
func (v *Validator) HasErrors() bool {
	return len(v.errors.Errors) > 0
}

// IsValidEmail checks if an email is valid (standalone function)
func IsValidEmail(email string) bool {
	if email == "" || len(email) > 255 {
		return false
	}
	return emailRegex.MatchString(email)
}

// IsValidPassword checks if a password meets strength requirements (standalone function)
func IsValidPassword(password string) bool {
	if len(password) < 8 || len(password) > 72 {
		return false
	}

	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(password)

	return hasUpper && hasLower && hasDigit
}
