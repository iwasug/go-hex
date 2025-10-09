package validator

import (
	"testing"
)

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name      string
		email     string
		wantError bool
		errorMsg  string
	}{
		{
			name:      "valid email",
			email:     "test@example.com",
			wantError: false,
		},
		{
			name:      "valid email with subdomain",
			email:     "user@mail.example.com",
			wantError: false,
		},
		{
			name:      "valid email with plus",
			email:     "user+tag@example.com",
			wantError: false,
		},
		{
			name:      "empty email",
			email:     "",
			wantError: true,
			errorMsg:  "email is required",
		},
		{
			name:      "invalid email format - no @",
			email:     "invalidemail.com",
			wantError: true,
			errorMsg:  "email format is invalid",
		},
		{
			name:      "invalid email format - no domain",
			email:     "user@",
			wantError: true,
			errorMsg:  "email format is invalid",
		},
		{
			name:      "invalid email format - no TLD",
			email:     "user@domain",
			wantError: true,
			errorMsg:  "email format is invalid",
		},
		{
			name:      "email too long",
			email:     "verylongemailaddressthatexceedsthemaximumlengthof255characters@verylongdomainname.verylongdomainname.verylongdomainname.verylongdomainname.verylongdomainname.verylongdomainname.verylongdomainname.verylongdomainname.verylongdomainname.verylongdomainname.com",
			wantError: true,
			errorMsg:  "email must not exceed 255 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := New()
			v.ValidateEmail("email", tt.email)

			if tt.wantError && !v.HasErrors() {
				t.Errorf("expected error but got none")
			}

			if !tt.wantError && v.HasErrors() {
				t.Errorf("expected no error but got: %v", v.Errors())
			}

			if tt.wantError && v.HasErrors() {
				if v.errors.Errors[0].Message != tt.errorMsg {
					t.Errorf("expected error message '%s' but got '%s'", tt.errorMsg, v.errors.Errors[0].Message)
				}
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name      string
		password  string
		wantError bool
		errorMsg  string
	}{
		{
			name:      "valid password",
			password:  "Password123",
			wantError: false,
		},
		{
			name:      "valid password with special chars",
			password:  "P@ssw0rd!",
			wantError: false,
		},
		{
			name:      "empty password",
			password:  "",
			wantError: true,
			errorMsg:  "password is required",
		},
		{
			name:      "password too short",
			password:  "Pass1",
			wantError: true,
			errorMsg:  "password must be at least 8 characters long",
		},
		{
			name:      "password too long",
			password:  "Password123" + string(make([]byte, 70)),
			wantError: true,
			errorMsg:  "password must not exceed 72 characters",
		},
		{
			name:      "password without uppercase",
			password:  "password123",
			wantError: true,
			errorMsg:  "password must contain at least one uppercase letter",
		},
		{
			name:      "password without lowercase",
			password:  "PASSWORD123",
			wantError: true,
			errorMsg:  "password must contain at least one lowercase letter",
		},
		{
			name:      "password without digit",
			password:  "PasswordABC",
			wantError: true,
			errorMsg:  "password must contain at least one digit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := New()
			v.ValidatePassword("password", tt.password)

			if tt.wantError && !v.HasErrors() {
				t.Errorf("expected error but got none")
			}

			if !tt.wantError && v.HasErrors() {
				t.Errorf("expected no error but got: %v", v.Errors())
			}

			if tt.wantError && v.HasErrors() {
				found := false
				for _, err := range v.errors.Errors {
					if err.Message == tt.errorMsg {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected error message '%s' but got: %v", tt.errorMsg, v.errors.Errors)
				}
			}
		})
	}
}

func TestValidateRequired(t *testing.T) {
	tests := []struct {
		name      string
		field     string
		value     string
		wantError bool
	}{
		{
			name:      "valid non-empty value",
			field:     "name",
			value:     "John Doe",
			wantError: false,
		},
		{
			name:      "empty value",
			field:     "name",
			value:     "",
			wantError: true,
		},
		{
			name:      "whitespace only",
			field:     "name",
			value:     "   ",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := New()
			v.ValidateRequired(tt.field, tt.value)

			if tt.wantError && !v.HasErrors() {
				t.Errorf("expected error but got none")
			}

			if !tt.wantError && v.HasErrors() {
				t.Errorf("expected no error but got: %v", v.Errors())
			}
		})
	}
}

func TestValidateEnum(t *testing.T) {
	tests := []struct {
		name          string
		field         string
		value         string
		allowedValues []string
		wantError     bool
	}{
		{
			name:          "valid enum value",
			field:         "status",
			value:         "active",
			allowedValues: []string{"active", "inactive", "pending"},
			wantError:     false,
		},
		{
			name:          "invalid enum value",
			field:         "status",
			value:         "deleted",
			allowedValues: []string{"active", "inactive", "pending"},
			wantError:     true,
		},
		{
			name:          "empty value",
			field:         "status",
			value:         "",
			allowedValues: []string{"active", "inactive", "pending"},
			wantError:     false, // Empty values are allowed, handled by ValidateRequired
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := New()
			v.ValidateEnum(tt.field, tt.value, tt.allowedValues)

			if tt.wantError && !v.HasErrors() {
				t.Errorf("expected error but got none")
			}

			if !tt.wantError && v.HasErrors() {
				t.Errorf("expected no error but got: %v", v.Errors())
			}
		})
	}
}

func TestValidateOrderBy(t *testing.T) {
	tests := []struct {
		name          string
		value         string
		allowedFields []string
		wantError     bool
	}{
		{
			name:          "valid order_by field",
			value:         "name",
			allowedFields: []string{"name", "email", "created_at"},
			wantError:     false,
		},
		{
			name:          "invalid order_by field",
			value:         "password",
			allowedFields: []string{"name", "email", "created_at"},
			wantError:     true,
		},
		{
			name:          "empty order_by",
			value:         "",
			allowedFields: []string{"name", "email", "created_at"},
			wantError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := New()
			v.ValidateOrderBy("order_by", tt.value, tt.allowedFields)

			if tt.wantError && !v.HasErrors() {
				t.Errorf("expected error but got none")
			}

			if !tt.wantError && v.HasErrors() {
				t.Errorf("expected no error but got: %v", v.Errors())
			}
		})
	}
}

func TestValidateOrderDirection(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		wantError bool
	}{
		{
			name:      "valid asc",
			value:     "asc",
			wantError: false,
		},
		{
			name:      "valid desc",
			value:     "desc",
			wantError: false,
		},
		{
			name:      "invalid direction",
			value:     "ascending",
			wantError: true,
		},
		{
			name:      "empty direction",
			value:     "",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := New()
			v.ValidateOrderDirection("order", tt.value)

			if tt.wantError && !v.HasErrors() {
				t.Errorf("expected error but got none")
			}

			if !tt.wantError && v.HasErrors() {
				t.Errorf("expected no error but got: %v", v.Errors())
			}
		})
	}
}

func TestValidateMinLength(t *testing.T) {
	v := New()
	v.ValidateMinLength("name", "ab", 3)

	if !v.HasErrors() {
		t.Errorf("expected error for string shorter than min length")
	}

	v2 := New()
	v2.ValidateMinLength("name", "abc", 3)

	if v2.HasErrors() {
		t.Errorf("expected no error for string equal to min length")
	}
}

func TestValidateMaxLength(t *testing.T) {
	v := New()
	v.ValidateMaxLength("name", "abcde", 3)

	if !v.HasErrors() {
		t.Errorf("expected error for string longer than max length")
	}

	v2 := New()
	v2.ValidateMaxLength("name", "abc", 3)

	if v2.HasErrors() {
		t.Errorf("expected no error for string equal to max length")
	}
}

func TestValidateRange(t *testing.T) {
	v := New()
	v.ValidateRange("age", 5, 10, 20)

	if !v.HasErrors() {
		t.Errorf("expected error for value below range")
	}

	v2 := New()
	v2.ValidateRange("age", 25, 10, 20)

	if !v2.HasErrors() {
		t.Errorf("expected error for value above range")
	}

	v3 := New()
	v3.ValidateRange("age", 15, 10, 20)

	if v3.HasErrors() {
		t.Errorf("expected no error for value within range")
	}
}

func TestValidatePositive(t *testing.T) {
	v := New()
	v.ValidatePositive("count", 0)

	if !v.HasErrors() {
		t.Errorf("expected error for zero value")
	}

	v2 := New()
	v2.ValidatePositive("count", -5)

	if !v2.HasErrors() {
		t.Errorf("expected error for negative value")
	}

	v3 := New()
	v3.ValidatePositive("count", 5)

	if v3.HasErrors() {
		t.Errorf("expected no error for positive value")
	}
}

func TestMultipleValidations(t *testing.T) {
	v := New()
	v.ValidateEmail("email", "invalid")
	v.ValidatePassword("password", "weak")
	v.ValidateRequired("name", "")

	if !v.HasErrors() {
		t.Errorf("expected multiple errors")
	}

	if len(v.errors.Errors) != 3 {
		t.Errorf("expected 3 errors but got %d", len(v.errors.Errors))
	}
}

func TestValidationErrorsError(t *testing.T) {
	v := New()
	v.ValidateRequired("name", "")
	v.ValidateEmail("email", "invalid")

	errMsg := v.Errors().Error()
	if errMsg == "" {
		t.Errorf("expected non-empty error message")
	}

	if errMsg != "name: name is required; email: email format is invalid" {
		t.Errorf("unexpected error message format: %s", errMsg)
	}
}

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		email string
		valid bool
	}{
		{"test@example.com", true},
		{"user+tag@example.com", true},
		{"invalid", false},
		{"", false},
		{"@example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			result := IsValidEmail(tt.email)
			if result != tt.valid {
				t.Errorf("IsValidEmail(%s) = %v, want %v", tt.email, result, tt.valid)
			}
		})
	}
}

func TestIsValidPassword(t *testing.T) {
	tests := []struct {
		password string
		valid    bool
	}{
		{"Password123", true},
		{"P@ssw0rd!", true},
		{"weak", false},
		{"password123", false},
		{"PASSWORD123", false},
		{"PasswordABC", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.password, func(t *testing.T) {
			result := IsValidPassword(tt.password)
			if result != tt.valid {
				t.Errorf("IsValidPassword(%s) = %v, want %v", tt.password, result, tt.valid)
			}
		})
	}
}
