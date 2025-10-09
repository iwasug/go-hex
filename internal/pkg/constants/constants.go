package constants

import "time"

const (
	DefaultPageSize   = 10
	MaxPageSize       = 100
	MinPasswordLength = 8
	MaxPasswordLength = 128
	MaxNameLength     = 255
	MaxEmailLength    = 255
	MaxDescLength     = 500

	DefaultRequestTimeout  = 30 * time.Second
	DefaultDBTimeout       = 10 * time.Second
	DefaultShutdownTimeout = 30 * time.Second

	HeaderRequestID     = "X-Request-ID"
	HeaderContentType   = "Content-Type"
	HeaderAuthorization = "Authorization"

	ContentTypeJSON = "application/json"

	ErrorCodeValidation   = "VALIDATION_ERROR"
	ErrorCodeUnauthorized = "UNAUTHORIZED"
	ErrorCodeForbidden    = "FORBIDDEN"
	ErrorCodeNotFound     = "NOT_FOUND"
	ErrorCodeConflict     = "CONFLICT"
	ErrorCodeInternal     = "INTERNAL_ERROR"
	ErrorCodeRateLimit    = "RATE_LIMIT_EXCEEDED"
	ErrorCodeBadRequest   = "BAD_REQUEST"

	DefaultRoleUser  = "user"
	DefaultRoleAdmin = "admin"

	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"

	ContextKeyUserID    = "user_id"
	ContextKeyUserEmail = "user_email"
	ContextKeyUserRoles = "user_roles"
	ContextKeyRequestID = "request_id"
)
