package request

// CreateSessionRequest represents the request body for creating a WhatsApp session
type CreateSessionRequest struct {
	SessionName   string `json:"session_name"`
	WebhookURL    string `json:"webhook_url,omitempty"`
	WebhookSecret string `json:"webhook_secret,omitempty"`
}

// ListSessionsQueryParams represents query parameters for listing sessions
type ListSessionsQueryParams struct {
	Status   string `json:"status,omitempty"`
	IsActive string `json:"is_active,omitempty"`
	Search   string `json:"search,omitempty"`
	OrderBy  string `json:"order_by,omitempty"`
	Order    string `json:"order,omitempty"`
	Page     string `json:"page,omitempty"`
	Limit    string `json:"limit,omitempty"`
}

// GetQRCodeQueryParams represents query parameters for QR code generation
type GetQRCodeQueryParams struct {
	Timeout string `json:"timeout,omitempty"` // Timeout in seconds
}

// SendTextMessageRequest represents the request body for sending a text message
type SendTextMessageRequest struct {
	To      string `json:"to"`      // Phone number in E.164 format
	Content string `json:"content"` // Message content
}

// SendMediaMessageRequest represents the request body for sending a media message
// This is used with multipart form data
type SendMediaMessageRequest struct {
	To       string `json:"to"`                 // Phone number in E.164 format
	Caption  string `json:"caption,omitempty"`  // Optional caption for media
	FileName string `json:"filename,omitempty"` // Required for documents
}

// GetMessagesQueryParams represents query parameters for retrieving messages
type GetMessagesQueryParams struct {
	Direction string `json:"direction,omitempty"` // inbound or outbound
	StartDate string `json:"start_date,omitempty"` // ISO 8601 format
	EndDate   string `json:"end_date,omitempty"`   // ISO 8601 format
	Page      string `json:"page,omitempty"`       // Page number
	Limit     string `json:"limit,omitempty"`      // Items per page
}
