package service

import (
	"context"
	"fmt"
	"time"

	"golang-chi-hexagonal-base/internal/domain/entity"
	"golang-chi-hexagonal-base/internal/domain/repository"
	"golang-chi-hexagonal-base/internal/pkg/errors"
	"golang-chi-hexagonal-base/internal/pkg/logger"
)

// WhatsAppClientManager defines the interface for managing WhatsApp clients
// This is a port to the adapter layer
type WhatsAppClientManager interface {
	CreateClient(ctx context.Context, sessionID, userID string) error
	GetClient(sessionID string) (WhatsAppClient, error)
	DeleteClient(ctx context.Context, sessionID string) error
	RestoreClients(ctx context.Context) error
}

// WhatsAppClient defines the interface for a WhatsApp client
type WhatsAppClient interface {
	Connect() error
	Disconnect() error
	GetQRCode(ctx context.Context, timeout time.Duration) (string, error)
	GetPairingCode(ctx context.Context, phoneNumber string, timeout time.Duration) (string, error)
	SendTextMessage(ctx context.Context, to, content string) (string, error)
	SendMediaMessage(ctx context.Context, to string, media interface{}) (string, error)
	IsConnected() bool
	IsLoggedIn() bool
	GetConnectedPhone() string
	GetSessionID() string
	GetUserID() string
}

// MediaManager defines the interface for managing media files
type MediaManager interface {
	DeleteSessionMedia(sessionID string) error
}

// WebhookClient defines the interface for webhook delivery
type WebhookClient interface {
	SendWebhook(ctx context.Context, req WebhookRequest) *WebhookResponse
}

// WebhookRequest represents a webhook delivery request
type WebhookRequest struct {
	URL     string
	Secret  string
	Payload WebhookPayload
}

// WebhookPayload represents the data sent to webhook endpoints
type WebhookPayload struct {
	Event     string                 `json:"event"`
	SessionID string                 `json:"session_id"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// WebhookResponse represents the result of a webhook delivery attempt
type WebhookResponse struct {
	Success      bool
	StatusCode   int
	ResponseBody string
	Error        error
	Attempts     int
	Duration     time.Duration
}

// WhatsAppService handles WhatsApp session and messaging operations
type WhatsAppService struct {
	sessionRepo   repository.WhatsAppSessionRepository
	messageRepo   repository.WhatsAppMessageRepository
	userRepo      repository.UserRepository
	clientManager WhatsAppClientManager
	webhookClient WebhookClient
	mediaManager  MediaManager
	logger        *logger.WhatsAppLogger
}

// NewWhatsAppService creates a new WhatsAppService instance
func NewWhatsAppService(
	sessionRepo repository.WhatsAppSessionRepository,
	messageRepo repository.WhatsAppMessageRepository,
	userRepo repository.UserRepository,
	clientManager WhatsAppClientManager,
	webhookClient WebhookClient,
	mediaManager MediaManager,
	log logger.LoggerInterface,
) *WhatsAppService {
	return &WhatsAppService{
		sessionRepo:   sessionRepo,
		messageRepo:   messageRepo,
		userRepo:      userRepo,
		clientManager: clientManager,
		webhookClient: webhookClient,
		mediaManager:  mediaManager,
		logger:        logger.NewWhatsAppLogger(log),
	}
}

// CreateSessionInput represents the input for creating a WhatsApp session
type CreateSessionInput struct {
	SessionName   string
	WebhookURL    string
	WebhookSecret string
}

// CreateSession creates a new WhatsApp session for a user
func (s *WhatsAppService) CreateSession(ctx context.Context, userID string, input CreateSessionInput) (*entity.WhatsAppSession, error) {
	// Validate user exists
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		s.logger.LogErrorWithContext("create_session", err, map[string]interface{}{
			"user_id":      userID,
			"session_name": input.SessionName,
		})
		return nil, ErrUserNotFound
	}

	// Check if user is deleted
	if user.IsDeleted {
		s.logger.LogWarningWithContext("create_session", "user is deleted", map[string]interface{}{
			"user_id":      userID,
			"session_name": input.SessionName,
		})
		return nil, ErrUserNotFound
	}

	// Check if session with same name already exists for this user
	existingSession, err := s.sessionRepo.FindByUserIDAndName(ctx, userID, input.SessionName)
	if err == nil && existingSession != nil {
		s.logger.LogWarningWithContext("create_session", "session already exists", map[string]interface{}{
			"user_id":      userID,
			"session_name": input.SessionName,
		})
		return nil, errors.NewSessionAlreadyExistsError(userID, input.SessionName)
	}

	// Create new session entity
	session := entity.NewWhatsAppSession(userID, input.SessionName, &userID)
	
	// Set webhook if provided
	if input.WebhookURL != "" {
		session.SetWebhook(input.WebhookURL, input.WebhookSecret)
	}

	// Save session to database
	err = s.sessionRepo.Create(ctx, session)
	if err != nil {
		s.logger.LogErrorWithContext("create_session", err, map[string]interface{}{
			"user_id":      userID,
			"session_name": input.SessionName,
		})
		return nil, errors.NewSessionCreationFailedError(err)
	}

	// Log session creation
	s.logger.LogSessionCreated(session.ID, userID, input.SessionName)

	// Create WhatsApp client for this session
	err = s.clientManager.CreateClient(ctx, session.ID, userID)
	if err != nil {
		// If client creation fails, we should still return the session
		// The client can be created later when needed
		s.logger.LogWarningWithContext("create_session", "client creation failed", map[string]interface{}{
			"session_id":   session.ID,
			"user_id":      userID,
			"session_name": input.SessionName,
			"error":        err.Error(),
		})
		return session, nil
	}

	return session, nil
}

// GetSession retrieves a WhatsApp session by ID with ownership check
func (s *WhatsAppService) GetSession(ctx context.Context, sessionID, userID string, isAdmin bool) (*entity.WhatsAppSession, error) {
	// Find session by ID
	session, err := s.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		s.logger.LogErrorWithContext("get_session", err, map[string]interface{}{
			"session_id": sessionID,
			"user_id":    userID,
		})
		return nil, errors.NewSessionNotFoundError(sessionID)
	}

	// Check ownership (admins can access any session)
	if !isAdmin && session.UserID != userID {
		s.logger.LogWarningWithContext("get_session", "unauthorized access attempt", map[string]interface{}{
			"session_id":       sessionID,
			"requesting_user":  userID,
			"session_owner":    session.UserID,
		})
		return nil, errors.NewUnauthorizedSessionError(userID, sessionID)
	}

	return session, nil
}

// ListSessionsInput represents the input for listing WhatsApp sessions
type ListSessionsInput struct {
	Status   entity.SessionStatus
	IsActive *bool
	Search   string
	OrderBy  string
	Order    string
	Page     int
	Limit    int
}

// ListSessionsResult represents the result of listing WhatsApp sessions
type ListSessionsResult struct {
	Sessions   []*entity.WhatsAppSession
	Total      int
	Page       int
	Limit      int
	TotalPages int
}

// ListSessions retrieves a list of WhatsApp sessions with filtering
func (s *WhatsAppService) ListSessions(ctx context.Context, userID string, isAdmin bool, input ListSessionsInput) (*ListSessionsResult, error) {
	// Set defaults
	if input.Page < 1 {
		input.Page = 1
	}
	if input.Limit < 1 {
		input.Limit = 10
	}
	if input.Limit > 100 {
		input.Limit = 100
	}
	if input.OrderBy == "" {
		input.OrderBy = "created_at"
	}
	if input.Order == "" {
		input.Order = "desc"
	}

	// Calculate offset
	offset := (input.Page - 1) * input.Limit

	// Build repository params
	params := repository.ListSessionsParams{
		UserID:     userID,
		Status:     input.Status,
		IsActive:   input.IsActive,
		Search:     input.Search,
		OrderBy:    input.OrderBy,
		Order:      input.Order,
		Limit:      input.Limit,
		Offset:     offset,
		IncludeAll: isAdmin, // Admins can see all sessions
	}

	// If admin, allow viewing all sessions
	if isAdmin {
		params.UserID = "" // Empty userID means all users
	}

	// Get sessions from repository
	sessions, total, err := s.sessionRepo.List(ctx, params)
	if err != nil {
		s.logger.LogErrorWithContext("list_sessions", err, map[string]interface{}{
			"user_id":  userID,
			"is_admin": isAdmin,
			"page":     input.Page,
			"limit":    input.Limit,
		})
		return nil, errors.NewWhatsAppError(errors.ErrCodeServiceUnavailable, "Failed to list sessions", 500, err)
	}

	// Calculate total pages
	totalPages := (total + input.Limit - 1) / input.Limit

	return &ListSessionsResult{
		Sessions:   sessions,
		Total:      total,
		Page:       input.Page,
		Limit:      input.Limit,
		TotalPages: totalPages,
	}, nil
}

// DeleteSession deletes a WhatsApp session with cleanup
func (s *WhatsAppService) DeleteSession(ctx context.Context, sessionID, userID string, isAdmin bool) error {
	// Find session by ID
	session, err := s.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		s.logger.LogErrorWithContext("delete_session", err, map[string]interface{}{
			"session_id": sessionID,
			"user_id":    userID,
		})
		return errors.NewSessionNotFoundError(sessionID)
	}

	// Check ownership (admins can delete any session)
	if !isAdmin && session.UserID != userID {
		s.logger.LogWarningWithContext("delete_session", "unauthorized delete attempt", map[string]interface{}{
			"session_id":      sessionID,
			"requesting_user": userID,
			"session_owner":   session.UserID,
		})
		return errors.NewUnauthorizedSessionError(userID, sessionID)
	}

	// Delete WhatsApp client (disconnect and cleanup)
	err = s.clientManager.DeleteClient(ctx, sessionID)
	if err != nil {
		// Log error but continue with session deletion
		// The client might not exist or already be deleted
		s.logger.LogWarningWithContext("delete_session", "client deletion failed", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
	}

	// Delete media files for this session
	if s.mediaManager != nil {
		err = s.mediaManager.DeleteSessionMedia(sessionID)
		if err != nil {
			// Log error but continue with session deletion
			// Media cleanup failure shouldn't prevent session deletion
			s.logger.LogWarningWithContext("delete_session", "media cleanup failed", map[string]interface{}{
				"session_id": sessionID,
				"error":      err.Error(),
			})
		}
	}

	// Delete session from database
	err = s.sessionRepo.Delete(ctx, sessionID)
	if err != nil {
		s.logger.LogErrorWithContext("delete_session", err, map[string]interface{}{
			"session_id": sessionID,
			"user_id":    userID,
		})
		return errors.NewWhatsAppError(errors.ErrCodeSessionDeleteFailed, "Failed to delete session", 500, err)
	}

	// Log session deletion
	s.logger.LogSessionDeleted(sessionID, userID)

	return nil
}

// GenerateQRCodeInput represents the input for generating a QR code
type GenerateQRCodeInput struct {
	Timeout time.Duration // How long to wait for QR code generation
}

// GeneratePairingCodeInput represents the input for generating a pairing code
type GeneratePairingCodeInput struct {
	PhoneNumber string
	Timeout     time.Duration
}

// GeneratePairingCode generates an 8-digit pairing code for WhatsApp authentication via phone number
func (s *WhatsAppService) GeneratePairingCode(ctx context.Context, sessionID, userID string, isAdmin bool, input GeneratePairingCodeInput) (string, error) {
	// Get session from database
	session, err := s.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		s.logger.LogErrorWithContext("generate_pairing_code", err, map[string]interface{}{
			"session_id": sessionID,
		})
		return "", errors.NewSessionNotFoundError(sessionID)
	}

	// Check ownership (admins can generate pairing code for any session)
	if !isAdmin && session.UserID != userID {
		s.logger.LogWarningWithContext("generate_pairing_code", "unauthorized access attempt", map[string]interface{}{
			"session_id":      sessionID,
			"requesting_user": userID,
			"session_owner":   session.UserID,
		})
		return "", errors.NewUnauthorizedSessionError(userID, sessionID)
	}

	// Check if session is already connected
	if session.IsConnected() {
		s.logger.LogWarningWithContext("generate_pairing_code", "session already connected", map[string]interface{}{
			"session_id": sessionID,
		})
		return "", errors.NewWhatsAppError(errors.ErrCodeSessionNotConnected, "Session is already connected", 400, nil)
	}

	// Validate phone number
	if input.PhoneNumber == "" {
		return "", errors.NewInvalidPhoneNumberError(input.PhoneNumber)
	}

	// Set default timeout if not provided
	if input.Timeout == 0 {
		input.Timeout = 60 * time.Second
	}

	// Get WhatsApp client
	client, err := s.clientManager.GetClient(sessionID)
	if err != nil {
		// If client doesn't exist, try to create it
		s.logger.LogWarningWithContext("generate_pairing_code", "client not found, creating new client", map[string]interface{}{
			"session_id": sessionID,
		})
		err = s.clientManager.CreateClient(ctx, sessionID, session.UserID)
		if err != nil {
			s.logger.LogErrorWithContext("generate_pairing_code", err, map[string]interface{}{
				"session_id": sessionID,
				"operation":  "create_client",
			})
			return "", errors.NewClientCreationFailedError(sessionID, err)
		}

		// Get the newly created client
		client, err = s.clientManager.GetClient(sessionID)
		if err != nil {
			s.logger.LogErrorWithContext("generate_pairing_code", err, map[string]interface{}{
				"session_id": sessionID,
				"operation":  "get_client",
			})
			return "", errors.NewClientNotFoundError(sessionID)
		}
	}

	// Generate pairing code
	startTime := time.Now()
	pairingCode, err := client.GetPairingCode(ctx, input.PhoneNumber, input.Timeout)
	if err != nil {
		s.logger.LogErrorWithContext("generate_pairing_code", err, map[string]interface{}{
			"session_id":   sessionID,
			"phone_number": input.PhoneNumber,
			"timeout":      input.Timeout.String(),
		})
		return "", errors.NewQRCodeGenerationFailedError(sessionID, err)
	}

	// Log successful pairing code generation
	s.logger.LogPerformanceMetric("generate_pairing_code", time.Since(startTime), map[string]interface{}{
		"session_id":   sessionID,
		"phone_number": input.PhoneNumber,
		"code_length":  len(pairingCode),
	})

	return pairingCode, nil
}

// GenerateQRCode generates a QR code for WhatsApp authentication
func (s *WhatsAppService) GenerateQRCode(ctx context.Context, sessionID, userID string, isAdmin bool, input GenerateQRCodeInput) (string, error) {
	// Find session by ID
	session, err := s.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		s.logger.LogErrorWithContext("generate_qr_code", err, map[string]interface{}{
			"session_id": sessionID,
			"user_id":    userID,
		})
		return "", errors.NewSessionNotFoundError(sessionID)
	}

	// Check ownership (admins can generate QR for any session)
	if !isAdmin && session.UserID != userID {
		s.logger.LogWarningWithContext("generate_qr_code", "unauthorized access attempt", map[string]interface{}{
			"session_id":      sessionID,
			"requesting_user": userID,
			"session_owner":   session.UserID,
		})
		return "", errors.NewUnauthorizedSessionError(userID, sessionID)
	}

	// Check if session is already connected
	if session.IsConnected() {
		s.logger.LogWarningWithContext("generate_qr_code", "session already connected", map[string]interface{}{
			"session_id": sessionID,
		})
		return "", errors.NewWhatsAppError(errors.ErrCodeSessionNotConnected, "Session is already connected", 400, nil)
	}

	// Set default timeout if not provided
	if input.Timeout == 0 {
		input.Timeout = 60 * time.Second
	}

	// Get WhatsApp client
	client, err := s.clientManager.GetClient(sessionID)
	if err != nil {
		// If client doesn't exist, try to create it
		s.logger.LogWarningWithContext("generate_qr_code", "client not found, creating new client", map[string]interface{}{
			"session_id": sessionID,
		})
		err = s.clientManager.CreateClient(ctx, sessionID, session.UserID)
		if err != nil {
			s.logger.LogErrorWithContext("generate_qr_code", err, map[string]interface{}{
				"session_id": sessionID,
				"operation":  "create_client",
			})
			return "", errors.NewClientCreationFailedError(sessionID, err)
		}
		
		// Get the newly created client
		client, err = s.clientManager.GetClient(sessionID)
		if err != nil {
			s.logger.LogErrorWithContext("generate_qr_code", err, map[string]interface{}{
				"session_id": sessionID,
				"operation":  "get_client",
			})
			return "", errors.NewClientNotFoundError(sessionID)
		}
	}

	// Generate QR code
	startTime := time.Now()
	qrCode, err := client.GetQRCode(ctx, input.Timeout)
	if err != nil {
		s.logger.LogErrorWithContext("generate_qr_code", err, map[string]interface{}{
			"session_id": sessionID,
			"timeout":    input.Timeout.String(),
		})
		return "", errors.NewQRCodeGenerationFailedError(sessionID, err)
	}

	// Update session with QR code and expiry
	expiry := time.Now().Add(input.Timeout)
	session.SetQRCode(qrCode, expiry)
	
	err = s.sessionRepo.Update(ctx, session)
	if err != nil {
		// Log error but still return the QR code
		// The QR code is still valid even if we couldn't save it
		s.logger.LogWarningWithContext("generate_qr_code", "failed to save QR code to database", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
	}

	// Log QR code generation
	s.logger.LogQRCodeGenerated(sessionID, expiry)
	s.logger.LogPerformanceMetric("generate_qr_code", time.Since(startTime), map[string]interface{}{
		"session_id": sessionID,
	})

	return qrCode, nil
}

// ValidateSessionOwnership validates that a user owns a session
func (s *WhatsAppService) ValidateSessionOwnership(ctx context.Context, sessionID, userID string, isAdmin bool) error {
	// Find session by ID
	session, err := s.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		s.logger.LogErrorWithContext("validate_session_ownership", err, map[string]interface{}{
			"session_id": sessionID,
			"user_id":    userID,
		})
		return errors.NewSessionNotFoundError(sessionID)
	}

	// Check ownership (admins can access any session)
	if !isAdmin && session.UserID != userID {
		s.logger.LogWarningWithContext("validate_session_ownership", "unauthorized access attempt", map[string]interface{}{
			"session_id":      sessionID,
			"requesting_user": userID,
			"session_owner":   session.UserID,
		})
		return errors.NewUnauthorizedSessionError(userID, sessionID)
	}

	return nil
}

// GetSessionStatus retrieves the current status of a session
func (s *WhatsAppService) GetSessionStatus(ctx context.Context, sessionID, userID string, isAdmin bool) (*SessionStatusInfo, error) {
	// Find session by ID
	session, err := s.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		s.logger.LogErrorWithContext("get_session_status", err, map[string]interface{}{
			"session_id": sessionID,
			"user_id":    userID,
		})
		return nil, errors.NewSessionNotFoundError(sessionID)
	}

	// Check ownership (admins can access any session)
	if !isAdmin && session.UserID != userID {
		s.logger.LogWarningWithContext("get_session_status", "unauthorized access attempt", map[string]interface{}{
			"session_id":      sessionID,
			"requesting_user": userID,
			"session_owner":   session.UserID,
		})
		return nil, errors.NewUnauthorizedSessionError(userID, sessionID)
	}

	// Try to get client to check real-time connection status
	client, err := s.clientManager.GetClient(sessionID)
	
	statusInfo := &SessionStatusInfo{
		SessionID:    session.ID,
		Status:       session.Status,
		PhoneNumber:  session.PhoneNumber,
		LastActivity: session.LastActivity,
		IsActive:     session.IsActive,
	}

	// If client exists, get real-time status
	if err == nil && client != nil {
		statusInfo.IsConnected = client.IsConnected()
		statusInfo.IsLoggedIn = client.IsLoggedIn()
		if client.IsConnected() {
			statusInfo.ConnectedPhone = client.GetConnectedPhone()
		}
	}

	return statusInfo, nil
}

// SessionStatusInfo represents the status information of a session
type SessionStatusInfo struct {
	SessionID      string
	Status         entity.SessionStatus
	PhoneNumber    string
	ConnectedPhone string
	LastActivity   time.Time
	IsActive       bool
	IsConnected    bool
	IsLoggedIn     bool
}

// MediaType represents the type of media being sent
type MediaType string

const (
	MediaTypeImage    MediaType = "image"
	MediaTypeVideo    MediaType = "video"
	MediaTypeAudio    MediaType = "audio"
	MediaTypeDocument MediaType = "document"
)

// MediaMessage represents a media message to be sent
type MediaMessage struct {
	Type     MediaType
	Data     []byte
	MimeType string
	Caption  string
	FileName string
}

// SendTextMessageInput represents the input for sending a text message
type SendTextMessageInput struct {
	To      string // Phone number in E.164 format
	Content string
}

// SendMediaMessageInput represents the input for sending a media message
type SendMediaMessageInput struct {
	To       string // Phone number in E.164 format
	Media    MediaMessage
	MaxSizeMB int // Maximum file size in MB (default: 16MB)
}

// SendTextMessage sends a text message through WhatsApp
func (s *WhatsAppService) SendTextMessage(ctx context.Context, sessionID, userID string, isAdmin bool, input SendTextMessageInput) (*entity.WhatsAppMessage, error) {
	// Validate session ownership
	session, err := s.GetSession(ctx, sessionID, userID, isAdmin)
	if err != nil {
		return nil, err
	}

	// Validate session is connected
	if !session.IsConnected() {
		s.logger.LogWarningWithContext("send_text_message", "session not connected", map[string]interface{}{
			"session_id": sessionID,
			"recipient":  input.To,
		})
		return nil, errors.NewSessionNotConnectedError(sessionID)
	}

	// Validate phone number (E.164 format)
	if err := validatePhoneNumber(input.To); err != nil {
		s.logger.LogWarningWithContext("send_text_message", "invalid phone number", map[string]interface{}{
			"session_id":   sessionID,
			"phone_number": input.To,
			"error":        err.Error(),
		})
		return nil, errors.NewInvalidPhoneNumberError(input.To)
	}

	// Validate content is not empty
	if input.Content == "" {
		return nil, errors.NewWhatsAppError(errors.ErrCodeInvalidMessageType, "Message content cannot be empty", 400, nil)
	}

	// Validate content length (WhatsApp limit is 65536 characters)
	if len(input.Content) > 65536 {
		return nil, errors.NewWhatsAppError(errors.ErrCodeMessageTooLarge, "Message content exceeds maximum length of 65536 characters", 400, nil)
	}

	// Get WhatsApp client
	client, err := s.clientManager.GetClient(sessionID)
	if err != nil {
		s.logger.LogErrorWithContext("send_text_message", err, map[string]interface{}{
			"session_id": sessionID,
			"recipient":  input.To,
		})
		return nil, errors.NewClientNotFoundError(sessionID)
	}

	// Check if client is connected and logged in
	if !client.IsConnected() || !client.IsLoggedIn() {
		// Update session status
		session.UpdateStatus(entity.SessionStatusDisconnected)
		_ = s.sessionRepo.Update(ctx, session)
		s.logger.LogSessionDisconnected(sessionID, "client not connected during send")
		return nil, errors.NewSessionNotConnectedError(sessionID)
	}

	// Log message sending
	tempMessageID := fmt.Sprintf("temp_%d", time.Now().UnixNano())
	s.logger.LogMessageSending(sessionID, tempMessageID, input.To, "text")

	// Send message through WhatsApp client
	startTime := time.Now()
	messageID, err := client.SendTextMessage(ctx, input.To, input.Content)
	duration := time.Since(startTime)
	
	if err != nil {
		s.logger.LogMessageSendFailed(sessionID, tempMessageID, input.To, err)
		return nil, errors.NewMessageSendFailedError(sessionID, input.To, err)
	}

	// Log successful send
	s.logger.LogMessageSent(sessionID, messageID, input.To, duration)

	// Create message entity
	message := entity.NewOutboundMessage(
		sessionID,
		input.To,
		input.Content,
		"text",
	)

	// Set the message ID and mark as sent
	message.MarkAsSent(messageID)
	
	// Set the from phone number
	message.SetFrom(session.PhoneNumber)

	// Persist message to database
	err = s.messageRepo.Create(ctx, message)
	if err != nil {
		// Log error but don't fail the operation
		// The message was sent successfully, we just couldn't persist it
		s.logger.LogMessageStorageFailed(sessionID, messageID, err)
		return message, nil
	}

	s.logger.LogMessageStored(sessionID, messageID, "outbound")

	// Update session last activity
	session.UpdateLastActivity()
	_ = s.sessionRepo.Update(ctx, session)

	return message, nil
}

// SendMediaMessage sends a media message through WhatsApp
func (s *WhatsAppService) SendMediaMessage(ctx context.Context, sessionID, userID string, isAdmin bool, input SendMediaMessageInput) (*entity.WhatsAppMessage, error) {
	// Validate session ownership
	session, err := s.GetSession(ctx, sessionID, userID, isAdmin)
	if err != nil {
		return nil, err
	}

	// Validate session is connected
	if !session.IsConnected() {
		s.logger.LogWarningWithContext("send_media_message", "session not connected", map[string]interface{}{
			"session_id": sessionID,
			"recipient":  input.To,
			"media_type": input.Media.Type,
		})
		return nil, errors.NewSessionNotConnectedError(sessionID)
	}

	// Validate phone number (E.164 format)
	if err := validatePhoneNumber(input.To); err != nil {
		s.logger.LogWarningWithContext("send_media_message", "invalid phone number", map[string]interface{}{
			"session_id":   sessionID,
			"phone_number": input.To,
			"error":        err.Error(),
		})
		return nil, errors.NewInvalidPhoneNumberError(input.To)
	}

	// Validate media data is not empty
	if len(input.Media.Data) == 0 {
		return nil, errors.NewWhatsAppError(errors.ErrCodeInvalidMediaType, "Media data cannot be empty", 400, nil)
	}

	// Set default max size if not provided
	if input.MaxSizeMB == 0 {
		input.MaxSizeMB = 16 // Default 16MB
	}

	// Validate file size
	maxSizeBytes := int64(input.MaxSizeMB * 1024 * 1024)
	if int64(len(input.Media.Data)) > maxSizeBytes {
		s.logger.LogWarningWithContext("send_media_message", "media file too large", map[string]interface{}{
			"session_id": sessionID,
			"size":       len(input.Media.Data),
			"max_size":   maxSizeBytes,
		})
		return nil, errors.NewMediaTooLargeError(int64(len(input.Media.Data)), maxSizeBytes)
	}

	// Validate MIME type
	if input.Media.MimeType == "" {
		return nil, errors.NewWhatsAppError(errors.ErrCodeInvalidMediaType, "Media MIME type is required", 400, nil)
	}

	// Validate media type
	if err := validateMediaType(input.Media.Type, input.Media.MimeType); err != nil {
		s.logger.LogWarningWithContext("send_media_message", "invalid media type", map[string]interface{}{
			"session_id": sessionID,
			"media_type": input.Media.Type,
			"mime_type":  input.Media.MimeType,
			"error":      err.Error(),
		})
		return nil, errors.NewInvalidMediaTypeError(string(input.Media.Type))
	}

	// Validate document requires filename
	if input.Media.Type == MediaTypeDocument && input.Media.FileName == "" {
		return nil, errors.NewWhatsAppError(errors.ErrCodeInvalidMediaType, "Filename is required for document messages", 400, nil)
	}

	// Get WhatsApp client
	client, err := s.clientManager.GetClient(sessionID)
	if err != nil {
		s.logger.LogErrorWithContext("send_media_message", err, map[string]interface{}{
			"session_id": sessionID,
			"recipient":  input.To,
		})
		return nil, errors.NewClientNotFoundError(sessionID)
	}

	// Check if client is connected and logged in
	if !client.IsConnected() || !client.IsLoggedIn() {
		// Update session status
		session.UpdateStatus(entity.SessionStatusDisconnected)
		_ = s.sessionRepo.Update(ctx, session)
		s.logger.LogSessionDisconnected(sessionID, "client not connected during send")
		return nil, errors.NewSessionNotConnectedError(sessionID)
	}

	// Convert service MediaMessage to adapter MediaMessage
	adapterMedia := convertToAdapterMediaMessage(input.Media)

	// Log media upload started
	tempMessageID := fmt.Sprintf("temp_%d", time.Now().UnixNano())
	s.logger.LogMediaUploadStarted(sessionID, tempMessageID, string(input.Media.Type), int64(len(input.Media.Data)))

	// Send message through WhatsApp client
	startTime := time.Now()
	messageID, err := client.SendMediaMessage(ctx, input.To, adapterMedia)
	duration := time.Since(startTime)
	
	if err != nil {
		s.logger.LogMediaUploadFailed(sessionID, tempMessageID, string(input.Media.Type), err)
		return nil, errors.NewMessageSendFailedError(sessionID, input.To, err)
	}

	// Log successful upload
	s.logger.LogMediaUploadCompleted(sessionID, messageID, string(input.Media.Type), duration)

	// Create message entity
	content := input.Media.Caption
	if content == "" {
		content = fmt.Sprintf("[%s]", input.Media.Type)
	}

	message := entity.NewOutboundMessage(
		sessionID,
		input.To,
		content,
		string(input.Media.Type),
	)

	// Set the message ID and mark as sent
	message.MarkAsSent(messageID)
	
	// Set the from phone number
	message.SetFrom(session.PhoneNumber)

	// Persist message to database
	err = s.messageRepo.Create(ctx, message)
	if err != nil {
		// Log error but don't fail the operation
		// The message was sent successfully, we just couldn't persist it
		s.logger.LogMessageStorageFailed(sessionID, messageID, err)
		return message, nil
	}

	s.logger.LogMessageStored(sessionID, messageID, "outbound")

	// Update session last activity
	session.UpdateLastActivity()
	_ = s.sessionRepo.Update(ctx, session)

	return message, nil
}

// validatePhoneNumber validates a phone number in E.164 format
func validatePhoneNumber(phone string) error {
	if phone == "" {
		return fmt.Errorf("phone number cannot be empty")
	}

	// E.164 format: +[country code][number]
	// Must start with + and contain only digits after that
	// Length should be between 8 and 15 characters (including +)
	if phone[0] != '+' {
		return fmt.Errorf("phone number must start with + (E.164 format)")
	}

	// Remove the + and check the rest
	digits := phone[1:]
	if len(digits) < 7 || len(digits) > 15 {
		return fmt.Errorf("phone number must be between 8 and 16 characters (including +)")
	}

	// Check all characters are digits
	for _, c := range digits {
		if c < '0' || c > '9' {
			return fmt.Errorf("phone number must contain only digits after +")
		}
	}

	return nil
}

// validateMediaType validates the media type and MIME type combination
func validateMediaType(mediaType MediaType, mimeType string) error {
	validMimeTypes := map[MediaType][]string{
		MediaTypeImage: {
			"image/jpeg", "image/jpg", "image/png", "image/gif", "image/webp",
		},
		MediaTypeVideo: {
			"video/mp4", "video/3gpp", "video/quicktime", "video/x-matroska",
		},
		MediaTypeAudio: {
			"audio/aac", "audio/mp4", "audio/mpeg", "audio/amr", "audio/ogg",
		},
		MediaTypeDocument: {
			"application/pdf", "application/msword", "application/vnd.ms-excel",
			"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
			"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
			"application/vnd.openxmlformats-officedocument.presentationml.presentation",
			"text/plain", "application/zip", "application/x-zip-compressed",
		},
	}

	validTypes, ok := validMimeTypes[mediaType]
	if !ok {
		return fmt.Errorf("unsupported media type: %s", mediaType)
	}

	// Check if MIME type is valid for the media type
	for _, validType := range validTypes {
		if mimeType == validType {
			return nil
		}
	}

	return fmt.Errorf("MIME type %s is not valid for media type %s", mimeType, mediaType)
}

// AdapterMediaMessage represents the media message structure expected by the adapter layer
// This mirrors the structure in internal/adapter/whatsapp/client.go
type AdapterMediaMessage struct {
	Type     string
	Data     []byte
	MimeType string
	Caption  string
	FileName string
}

// convertToAdapterMediaMessage converts service MediaMessage to adapter MediaMessage
func convertToAdapterMediaMessage(media MediaMessage) AdapterMediaMessage {
	return AdapterMediaMessage{
		Type:     string(media.Type),
		Data:     media.Data,
		MimeType: media.MimeType,
		Caption:  media.Caption,
		FileName: media.FileName,
	}
}

// IncomingMessageData represents data for an incoming message
type IncomingMessageData struct {
	MessageID   string
	From        string
	To          string
	Content     string
	MessageType string
	MediaURL    string
	Caption     string
	Timestamp   time.Time
}

// HandleIncomingMessage processes an incoming WhatsApp message
// This method handles message deduplication, storage, and extraction of message details
func (s *WhatsAppService) HandleIncomingMessage(ctx context.Context, sessionID string, message IncomingMessageData) error {
	// Validate session exists
	session, err := s.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		s.logger.LogErrorWithContext("handle_incoming_message", err, map[string]interface{}{
			"session_id": sessionID,
			"message_id": message.MessageID,
		})
		return errors.NewSessionNotFoundError(sessionID)
	}

	// Log message received
	s.logger.LogMessageReceived(sessionID, message.MessageID, message.From, message.MessageType)

	// Check for message deduplication - see if message already exists
	existingMessage, err := s.messageRepo.FindByMessageID(ctx, message.MessageID)
	if err == nil && existingMessage != nil {
		// Message already exists, skip processing
		s.logger.LogWarningWithContext("handle_incoming_message", "duplicate message detected", map[string]interface{}{
			"session_id": sessionID,
			"message_id": message.MessageID,
		})
		return nil
	}

	// Extract message content based on type
	content := message.Content
	if content == "" {
		// For media messages without content, use caption or type indicator
		if message.Caption != "" {
			content = message.Caption
		} else if message.MessageType != "text" {
			content = fmt.Sprintf("[%s]", message.MessageType)
		}
	}

	// Handle media messages - store media URL if available
	if message.MediaURL != "" && message.MessageType != "text" {
		// For media messages, append media URL to content for reference
		if content != "" {
			content = fmt.Sprintf("%s\n[Media: %s]", content, message.MediaURL)
		} else {
			content = fmt.Sprintf("[Media: %s]", message.MediaURL)
		}
	}

	// Create inbound message entity
	inboundMessage := entity.NewInboundMessage(
		sessionID,
		message.MessageID,
		message.From,
		content,
		message.MessageType,
		message.Timestamp,
	)

	// Set the 'To' field (recipient - our WhatsApp number)
	inboundMessage.To = message.To

	// Store message in database
	err = s.messageRepo.Create(ctx, inboundMessage)
	if err != nil {
		// Log error but don't fail - requirement 8.5
		// The message was received, we just couldn't persist it
		s.logger.LogMessageStorageFailed(sessionID, message.MessageID, err)
		return errors.NewMessageStorageFailedError(message.MessageID, err)
	}

	s.logger.LogMessageStored(sessionID, message.MessageID, "inbound")

	// Update session last activity
	session.UpdateLastActivity()
	err = s.sessionRepo.Update(ctx, session)
	if err != nil {
		// Log error but don't fail - this is not critical
		// The message was stored successfully
		s.logger.LogWarningWithContext("handle_incoming_message", "failed to update session activity", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
	}

	// Send webhook if configured (requirement 4.3, 4.4, 4.5, 4.6)
	if session.WebhookURL != "" && s.webhookClient != nil {
		go s.deliverWebhook(context.Background(), session, inboundMessage)
	}

	return nil
}

// deliverWebhook delivers a webhook for an incoming message with retry logic
func (s *WhatsAppService) deliverWebhook(ctx context.Context, session *entity.WhatsAppSession, message *entity.WhatsAppMessage) {
	// Prepare webhook payload
	payload := WebhookPayload{
		Event:     "message.received",
		SessionID: session.ID,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"message_id":   message.MessageID,
			"from":         message.From,
			"to":           message.To,
			"content":      message.Content,
			"message_type": message.MessageType,
			"timestamp":    message.Timestamp,
			"direction":    message.Direction,
		},
	}

	// Create webhook request
	webhookReq := WebhookRequest{
		URL:     session.WebhookURL,
		Secret:  session.WebhookSecret,
		Payload: payload,
	}

	// Log webhook delivery started
	s.logger.LogWebhookDeliveryStarted(session.ID, message.MessageID, session.WebhookURL, 1)

	// Send webhook with retry logic
	response := s.webhookClient.SendWebhook(ctx, webhookReq)

	// Update message with webhook delivery status
	if response.Success {
		message.MarkWebhookDelivered()
		s.logger.LogWebhookDeliverySuccess(session.ID, message.MessageID, session.WebhookURL, response.StatusCode, response.Duration)
	} else {
		// Track attempts even on failure
		for i := 0; i < response.Attempts; i++ {
			message.IncrementWebhookAttempts()
		}
		
		if response.Attempts >= 3 {
			s.logger.LogWebhookMaxRetriesReached(session.ID, message.MessageID, session.WebhookURL, response.Attempts)
		} else {
			s.logger.LogWebhookDeliveryFailed(session.ID, message.MessageID, session.WebhookURL, response.Attempts, response.StatusCode, response.Error)
		}
	}

	// Update message in database
	err := s.messageRepo.Update(ctx, message)
	if err != nil {
		// Log error but don't fail - webhook was attempted
		// This is a background operation
		s.logger.LogWarningWithContext("deliver_webhook", "failed to update message webhook status", map[string]interface{}{
			"session_id": session.ID,
			"message_id": message.MessageID,
			"error":      err.Error(),
		})
	}
}

// GetMessagesInput represents the input for retrieving messages
type GetMessagesInput struct {
	Direction  entity.MessageDirection
	StartDate  *time.Time
	EndDate    *time.Time
	Page       int
	Limit      int
	OrderBy    string
	Order      string
}

// GetMessagesResult represents the result of retrieving messages
type GetMessagesResult struct {
	Messages   []*entity.WhatsAppMessage
	Total      int
	Page       int
	Limit      int
	TotalPages int
}

// GetMessages retrieves message history for a session
func (s *WhatsAppService) GetMessages(ctx context.Context, sessionID, userID string, isAdmin bool, input GetMessagesInput) (*GetMessagesResult, error) {
	// Validate session ownership
	_, err := s.GetSession(ctx, sessionID, userID, isAdmin)
	if err != nil {
		return nil, err
	}

	// Set defaults
	if input.Page < 1 {
		input.Page = 1
	}
	if input.Limit < 1 {
		input.Limit = 50
	}
	if input.Limit > 100 {
		input.Limit = 100
	}

	// Calculate offset
	offset := (input.Page - 1) * input.Limit

	// Build repository params
	params := repository.ListMessagesParams{
		SessionID: sessionID,
		Direction: input.Direction,
		StartDate: input.StartDate,
		EndDate:   input.EndDate,
		Limit:     input.Limit,
		Offset:    offset,
	}

	// Get messages from repository
	messages, total, err := s.messageRepo.List(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve messages: %w", err)
	}

	// Calculate total pages
	totalPages := (total + input.Limit - 1) / input.Limit

	return &GetMessagesResult{
		Messages:   messages,
		Total:      total,
		Page:       input.Page,
		Limit:      input.Limit,
		TotalPages: totalPages,
	}, nil
}

// ReconnectSession attempts to reconnect a disconnected WhatsApp session
func (s *WhatsAppService) ReconnectSession(ctx context.Context, sessionID, userID string, isAdmin bool) error {
	// Validate session ownership
	session, err := s.GetSession(ctx, sessionID, userID, isAdmin)
	if err != nil {
		return err
	}

	// Check if session is already connected
	client, err := s.clientManager.GetClient(sessionID)
	if err == nil && client != nil && client.IsConnected() {
		s.logger.LogWarningWithContext("reconnect_session", "session already connected", map[string]interface{}{
			"session_id": sessionID,
		})
		return errors.NewWhatsAppError(errors.ErrCodeSessionNotConnected, "Session is already connected", 400, nil)
	}

	// Log reconnection started
	s.logger.LogReconnectionStarted(sessionID, 1, 1)

	// Update session status to connecting
	oldStatus := session.Status
	session.Status = entity.SessionStatusConnecting
	session.LastActivity = time.Now()
	if err := s.sessionRepo.Update(ctx, session); err != nil {
		s.logger.LogErrorWithContext("reconnect_session", err, map[string]interface{}{
			"session_id": sessionID,
			"operation":  "update_status",
		})
		return errors.NewWhatsAppError(errors.ErrCodeSessionUpdateFailed, "Failed to update session status", 500, err)
	}

	s.logger.LogSessionStatusUpdate(sessionID, string(oldStatus), string(entity.SessionStatusConnecting))

	// If client doesn't exist, create it
	if client == nil {
		s.logger.LogWarningWithContext("reconnect_session", "client not found, creating new client", map[string]interface{}{
			"session_id": sessionID,
		})
		
		if err := s.clientManager.CreateClient(ctx, sessionID, userID); err != nil {
			// Update status to failed
			session.Status = entity.SessionStatusFailed
			_ = s.sessionRepo.Update(ctx, session)
			s.logger.LogReconnectionFailed(sessionID, 1, 1, err)
			return errors.NewClientCreationFailedError(sessionID, err)
		}
		
		// Get the newly created client
		client, err = s.clientManager.GetClient(sessionID)
		if err != nil {
			session.Status = entity.SessionStatusFailed
			_ = s.sessionRepo.Update(ctx, session)
			s.logger.LogReconnectionFailed(sessionID, 1, 1, err)
			return errors.NewClientNotFoundError(sessionID)
		}
	}

	// Attempt to connect
	startTime := time.Now()
	if err := client.Connect(); err != nil {
		// Update status to failed
		session.Status = entity.SessionStatusFailed
		_ = s.sessionRepo.Update(ctx, session)
		s.logger.LogReconnectionFailed(sessionID, 1, 1, err)
		return errors.NewClientConnectionFailedError(sessionID, err)
	}

	// Log successful reconnection
	s.logger.LogReconnectionSuccess(sessionID, 1, time.Since(startTime))

	return nil
}
