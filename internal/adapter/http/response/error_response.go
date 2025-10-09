package response

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains error details
type ErrorDetail struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// Error codes
const (
	ErrCodeValidation   = "VALIDATION_ERROR"
	ErrCodeUnauthorized = "UNAUTHORIZED"
	ErrCodeForbidden    = "FORBIDDEN"
	ErrCodeNotFound     = "NOT_FOUND"
	ErrCodeConflict     = "CONFLICT"
	ErrCodeInternal     = "INTERNAL_ERROR"
)

// NewErrorResponse creates a new error response
func NewErrorResponse(code, message string, details map[string]interface{}) ErrorResponse {
	return ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
}

// NewValidationError creates a validation error response
func NewValidationError(message string, details map[string]interface{}) ErrorResponse {
	return NewErrorResponse(ErrCodeValidation, message, details)
}

// NewUnauthorizedError creates an unauthorized error response
func NewUnauthorizedError(message string) ErrorResponse {
	return NewErrorResponse(ErrCodeUnauthorized, message, nil)
}

// NewForbiddenError creates a forbidden error response
func NewForbiddenError(message string) ErrorResponse {
	return NewErrorResponse(ErrCodeForbidden, message, nil)
}

// NewNotFoundError creates a not found error response
func NewNotFoundError(message string) ErrorResponse {
	return NewErrorResponse(ErrCodeNotFound, message, nil)
}

// NewConflictError creates a conflict error response
func NewConflictError(message string) ErrorResponse {
	return NewErrorResponse(ErrCodeConflict, message, nil)
}

// NewInternalError creates an internal error response
func NewInternalError(message string) ErrorResponse {
	return NewErrorResponse(ErrCodeInternal, message, nil)
}
