package entity

import "time"

// SessionStatus represents the connection status of a WhatsApp session
type SessionStatus string

const (
	SessionStatusDisconnected   SessionStatus = "disconnected"
	SessionStatusConnecting     SessionStatus = "connecting"
	SessionStatusConnected      SessionStatus = "connected"
	SessionStatusAuthenticating SessionStatus = "authenticating"
	SessionStatusFailed         SessionStatus = "failed"
)

// WhatsAppSession represents a WhatsApp session entity in the domain
type WhatsAppSession struct {
	ID                   string
	UserID               string
	SessionName          string
	PhoneNumber          string
	Status               SessionStatus
	QRCode               string
	QRCodeExpiry         *time.Time
	DeviceInfo           string
	LastActivity         time.Time
	WebhookURL           string
	WebhookSecret        string
	IsActive             bool
	ReconnectAttempts    int
	LastReconnectAttempt *time.Time
	CreatedBy            *string
	UpdatedBy            *string
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// NewWhatsAppSession creates a new WhatsAppSession instance
func NewWhatsAppSession(userID, sessionName string, createdBy *string) *WhatsAppSession {
	now := time.Now()
	return &WhatsAppSession{
		UserID:       userID,
		SessionName:  sessionName,
		Status:       SessionStatusDisconnected,
		LastActivity: now,
		IsActive:     true,
		CreatedBy:    createdBy,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// Connect updates the session status to connected and sets the phone number
func (s *WhatsAppSession) Connect(phoneNumber string) {
	s.Status = SessionStatusConnected
	s.PhoneNumber = phoneNumber
	s.LastActivity = time.Now()
	s.UpdatedAt = time.Now()
	s.QRCode = ""
	s.QRCodeExpiry = nil
}

// Disconnect updates the session status to disconnected
func (s *WhatsAppSession) Disconnect() {
	s.Status = SessionStatusDisconnected
	s.LastActivity = time.Now()
	s.UpdatedAt = time.Now()
}

// UpdateStatus updates the session status
func (s *WhatsAppSession) UpdateStatus(status SessionStatus) {
	s.Status = status
	s.LastActivity = time.Now()
	s.UpdatedAt = time.Now()
}

// SetQRCode sets the QR code and its expiry time
func (s *WhatsAppSession) SetQRCode(qrCode string, expiry time.Time) {
	s.QRCode = qrCode
	s.QRCodeExpiry = &expiry
	s.Status = SessionStatusAuthenticating
	s.UpdatedAt = time.Now()
}

// ClearQRCode clears the QR code data
func (s *WhatsAppSession) ClearQRCode() {
	s.QRCode = ""
	s.QRCodeExpiry = nil
	s.UpdatedAt = time.Now()
}

// IsQRCodeExpired checks if the QR code has expired
func (s *WhatsAppSession) IsQRCodeExpired() bool {
	if s.QRCodeExpiry == nil {
		return true
	}
	return time.Now().After(*s.QRCodeExpiry)
}

// UpdateDeviceInfo updates the device information
func (s *WhatsAppSession) UpdateDeviceInfo(deviceInfo string) {
	s.DeviceInfo = deviceInfo
	s.UpdatedAt = time.Now()
}

// SetWebhook sets the webhook URL and secret
func (s *WhatsAppSession) SetWebhook(webhookURL, webhookSecret string) {
	s.WebhookURL = webhookURL
	s.WebhookSecret = webhookSecret
	s.UpdatedAt = time.Now()
}

// UpdateLastActivity updates the last activity timestamp
func (s *WhatsAppSession) UpdateLastActivity() {
	s.LastActivity = time.Now()
	s.UpdatedAt = time.Now()
}

// Deactivate marks the session as inactive
func (s *WhatsAppSession) Deactivate(updatedBy *string) {
	s.IsActive = false
	s.Status = SessionStatusDisconnected
	s.UpdatedBy = updatedBy
	s.UpdatedAt = time.Now()
}

// Activate marks the session as active
func (s *WhatsAppSession) Activate(updatedBy *string) {
	s.IsActive = true
	s.UpdatedBy = updatedBy
	s.UpdatedAt = time.Now()
}

// IsConnected checks if the session is currently connected
func (s *WhatsAppSession) IsConnected() bool {
	return s.Status == SessionStatusConnected
}

// Update updates the session's basic information
func (s *WhatsAppSession) Update(sessionName string, updatedBy *string) {
	s.SessionName = sessionName
	s.UpdatedBy = updatedBy
	s.UpdatedAt = time.Now()
}

// IncrementReconnectAttempts increments the reconnection attempt counter
func (s *WhatsAppSession) IncrementReconnectAttempts() {
	s.ReconnectAttempts++
	now := time.Now()
	s.LastReconnectAttempt = &now
	s.UpdatedAt = now
}

// ResetReconnectAttempts resets the reconnection attempt counter
func (s *WhatsAppSession) ResetReconnectAttempts() {
	s.ReconnectAttempts = 0
	s.LastReconnectAttempt = nil
	s.UpdatedAt = time.Now()
}

// MarkAsFailed marks the session as failed after max reconnection attempts
func (s *WhatsAppSession) MarkAsFailed() {
	s.Status = SessionStatusFailed
	s.UpdatedAt = time.Now()
}

// ShouldAttemptReconnect checks if reconnection should be attempted based on max attempts
func (s *WhatsAppSession) ShouldAttemptReconnect(maxAttempts int) bool {
	return s.ReconnectAttempts < maxAttempts
}
