package response

// BaseResponse is the standard wrapper for all API responses
type BaseResponse struct {
	Success bool         `json:"success"`
	Message string       `json:"message,omitempty"`
	Data    interface{}  `json:"data,omitempty"`
	Error   *ErrorDetail `json:"error,omitempty"`
}

// NewSuccessResponse creates a successful response
func NewSuccessResponse(message string, data interface{}) BaseResponse {
	return BaseResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
}

// NewErrorResponseWrapper creates an error response
func NewErrorResponseWrapper(errorDetail ErrorDetail) BaseResponse {
	return BaseResponse{
		Success: false,
		Error:   &errorDetail,
	}
}

// Pagination represents pagination metadata
type Pagination struct {
	Total      int `json:"total"`
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	TotalPages int `json:"total_pages"`
}

// PaginatedResponse represents a paginated response
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Pagination Pagination  `json:"pagination"`
}

// NewSuccessPaginatedResponse creates a successful paginated response
func NewSuccessPaginatedResponse(message string, data interface{}, pagination Pagination) BaseResponse {
	return BaseResponse{
		Success: true,
		Message: message,
		Data: PaginatedResponse{
			Data:       data,
			Pagination: pagination,
		},
	}
}

// MessageResponse represents a simple message response without data
type MessageResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
