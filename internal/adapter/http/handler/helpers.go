package handler

import (
	"encoding/json"
	"net/http"

	"golang-chi-hexagonal-base/internal/pkg/validator"
)

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail represents error details
type ErrorDetail struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// respondJSON sends a JSON response
func respondJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// respondError sends an error response
func respondError(w http.ResponseWriter, statusCode int, code, message string, details interface{}) {
	respondJSON(w, statusCode, ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}

// respondValidationError sends a validation error response
func respondValidationError(w http.ResponseWriter, validationErrors *validator.ValidationErrors) {
	respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Validation failed", validationErrors.Errors)
}
