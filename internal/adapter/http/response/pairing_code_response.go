package response

import "time"

// PairingCodeResponse represents the response for pairing code generation
type PairingCodeResponse struct {
	SessionID   string    `json:"session_id"`
	PairingCode string    `json:"pairing_code"`
	PhoneNumber string    `json:"phone_number"`
	ExpiresAt   time.Time `json:"expires_at"`
	Status      string    `json:"status"`
}
