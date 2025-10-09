package handler

import (
	"encoding/json"
	"net/http"

	"golang-chi-hexagonal-base/internal/adapter/http/response"
	"golang-chi-hexagonal-base/internal/pkg/validator"
)

// respondJSON sends a JSON response
func respondJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// respondSuccess sends a success response
func respondSuccess(w http.ResponseWriter, statusCode int, message string, data interface{}) {
	respondJSON(w, statusCode, response.NewSuccessResponse(message, data))
}

// respondError sends an error response
func respondError(w http.ResponseWriter, statusCode int, code, message string, details map[string]interface{}) {
	errorDetail := response.ErrorDetail{
		Code:    code,
		Message: message,
		Details: details,
	}
	respondJSON(w, statusCode, response.NewErrorResponseWrapper(errorDetail))
}

// respondValidationError sends a validation error response
func respondValidationError(w http.ResponseWriter, validationErrors *validator.ValidationErrors) {
	respondError(w, http.StatusBadRequest, response.ErrCodeValidation, "Validation failed", map[string]interface{}{
		"fields": validationErrors.Errors,
	})
}
