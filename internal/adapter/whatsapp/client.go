package whatsapp

import (
	"context"
	"fmt"
	"sync"
	"time"

	"golang-chi-hexagonal-base/internal/pkg/logger"

	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
)

// Client wraps a whatsmeow.Client and provides additional functionality
type Client struct {
	sessionID     string
	userID        string
	waClient      *whatsmeow.Client
	store         *PostgresStore
	logger        logger.LoggerInterface
	eventHandlers EventHandlers
	connected     bool
	mu            sync.RWMutex
	eventHandler  *EventHandler
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

// IncomingMessage represents a message received from WhatsApp
type IncomingMessage struct {
	MessageID   string
	From        string
	To          string
	Content     string
	MessageType string
	MediaURL    string
	Caption     string
	Timestamp   time.Time
}

// Connect establishes a WebSocket connection to WhatsApp
func (c *Client) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connected {
		c.logger.Warn("client already connected")
		return nil
	}

	// Connect to WhatsApp
	if err := c.waClient.Connect(); err != nil {
		c.logger.Error("failed to connect to WhatsApp", logger.Fields{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to connect: %w", err)
	}

	c.connected = true
	c.logger.Info("client connected successfully")

	return nil
}

// Disconnect gracefully disconnects from WhatsApp
func (c *Client) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		c.logger.Warn("client already disconnected")
		return nil
	}

	// Disconnect from WhatsApp
	c.waClient.Disconnect()

	c.connected = false
	c.logger.Info("client disconnected successfully")

	return nil
}

// GetQRCode generates and returns a QR code as base64-encoded PNG
func (c *Client) GetQRCode(ctx context.Context, timeout time.Duration) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Check if already logged in
	if c.waClient.IsLoggedIn() {
		return "", fmt.Errorf("already logged in")
	}

	// Create a channel to receive QR code
	qrChan, err := c.waClient.GetQRChannel(ctx)
	if err != nil {
		c.logger.Error("failed to get QR channel", logger.Fields{
			"error": err.Error(),
		})
		return "", fmt.Errorf("failed to get QR channel: %w", err)
	}

	// Connect if not connected
	if !c.connected {
		if err := c.waClient.Connect(); err != nil {
			c.logger.Error("failed to connect for QR code", logger.Fields{
				"error": err.Error(),
			})
			return "", fmt.Errorf("failed to connect: %w", err)
		}
		c.connected = true
	}

	// Wait for QR code with timeout
	select {
	case evt := <-qrChan:
		switch evt.Event {
		case "code":
			// Return the QR code string directly
			// The service layer will handle encoding it to an image
			c.logger.Info("QR code generated successfully")
			return evt.Code, nil

		case "success":
			c.logger.Info("already logged in")
			return "", fmt.Errorf("already logged in")

		default:
			c.logger.Warn("unexpected QR event", logger.Fields{
				"event": evt.Event,
			})
			return "", fmt.Errorf("unexpected QR event: %s", evt.Event)
		}

	case <-time.After(timeout):
		c.logger.Error("QR code generation timeout")
		return "", fmt.Errorf("QR code generation timeout")

	case <-ctx.Done():
		c.logger.Error("QR code generation cancelled")
		return "", ctx.Err()
	}
}

// SendTextMessage sends a text message to a recipient
func (c *Client) SendTextMessage(ctx context.Context, to, content string) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.waClient.IsLoggedIn() {
		return "", fmt.Errorf("not logged in")
	}

	// Parse recipient JID
	recipient, err := c.parseJID(to)
	if err != nil {
		c.logger.Error("failed to parse recipient JID", logger.Fields{
			"to":    to,
			"error": err.Error(),
		})
		return "", fmt.Errorf("invalid recipient: %w", err)
	}

	// Create message
	msg := &waProto.Message{
		Conversation: &content,
	}

	// Send message
	resp, err := c.waClient.SendMessage(ctx, recipient, msg)
	if err != nil {
		c.logger.Error("failed to send text message", logger.Fields{
			"to":    to,
			"error": err.Error(),
		})
		return "", fmt.Errorf("failed to send message: %w", err)
	}

	c.logger.Info("text message sent successfully", logger.Fields{
		"to":         to,
		"message_id": resp.ID,
		"timestamp":  resp.Timestamp,
	})

	return resp.ID, nil
}

// SendMediaMessage sends a media message to a recipient
func (c *Client) SendMediaMessage(ctx context.Context, to string, media MediaMessage) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.waClient.IsLoggedIn() {
		return "", fmt.Errorf("not logged in")
	}

	// Parse recipient JID
	recipient, err := c.parseJID(to)
	if err != nil {
		c.logger.Error("failed to parse recipient JID", logger.Fields{
			"to":    to,
			"error": err.Error(),
		})
		return "", fmt.Errorf("invalid recipient: %w", err)
	}

	// Upload media
	uploaded, err := c.waClient.Upload(ctx, media.Data, whatsmeow.MediaType(media.MimeType))
	if err != nil {
		c.logger.Error("failed to upload media", logger.Fields{
			"to":    to,
			"type":  media.Type,
			"error": err.Error(),
		})
		return "", fmt.Errorf("failed to upload media: %w", err)
	}

	// Create message based on media type
	var msg *waProto.Message
	switch media.Type {
	case MediaTypeImage:
		msg = &waProto.Message{
			ImageMessage: &waProto.ImageMessage{
				URL:           &uploaded.URL,
				DirectPath:    &uploaded.DirectPath,
				MediaKey:      uploaded.MediaKey,
				Mimetype:      &media.MimeType,
				FileEncSHA256: uploaded.FileEncSHA256,
				FileSHA256:    uploaded.FileSHA256,
				FileLength:    &uploaded.FileLength,
				Caption:       &media.Caption,
			},
		}

	case MediaTypeVideo:
		msg = &waProto.Message{
			VideoMessage: &waProto.VideoMessage{
				URL:           &uploaded.URL,
				DirectPath:    &uploaded.DirectPath,
				MediaKey:      uploaded.MediaKey,
				Mimetype:      &media.MimeType,
				FileEncSHA256: uploaded.FileEncSHA256,
				FileSHA256:    uploaded.FileSHA256,
				FileLength:    &uploaded.FileLength,
				Caption:       &media.Caption,
			},
		}

	case MediaTypeAudio:
		msg = &waProto.Message{
			AudioMessage: &waProto.AudioMessage{
				URL:           &uploaded.URL,
				DirectPath:    &uploaded.DirectPath,
				MediaKey:      uploaded.MediaKey,
				Mimetype:      &media.MimeType,
				FileEncSHA256: uploaded.FileEncSHA256,
				FileSHA256:    uploaded.FileSHA256,
				FileLength:    &uploaded.FileLength,
			},
		}

	case MediaTypeDocument:
		msg = &waProto.Message{
			DocumentMessage: &waProto.DocumentMessage{
				URL:           &uploaded.URL,
				DirectPath:    &uploaded.DirectPath,
				MediaKey:      uploaded.MediaKey,
				Mimetype:      &media.MimeType,
				FileEncSHA256: uploaded.FileEncSHA256,
				FileSHA256:    uploaded.FileSHA256,
				FileLength:    &uploaded.FileLength,
				FileName:      &media.FileName,
			},
		}

	default:
		return "", fmt.Errorf("unsupported media type: %s", media.Type)
	}

	// Send message
	resp, err := c.waClient.SendMessage(ctx, recipient, msg)
	if err != nil {
		c.logger.Error("failed to send media message", logger.Fields{
			"to":    to,
			"type":  media.Type,
			"error": err.Error(),
		})
		return "", fmt.Errorf("failed to send message: %w", err)
	}

	c.logger.Info("media message sent successfully", logger.Fields{
		"to":         to,
		"type":       media.Type,
		"message_id": resp.ID,
	})

	return resp.ID, nil
}

// IsConnected returns whether the client is connected
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.connected && c.waClient.IsConnected()
}

// IsLoggedIn returns whether the client is logged in
func (c *Client) IsLoggedIn() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.waClient.IsLoggedIn()
}

// GetConnectedPhone returns the connected phone number
func (c *Client) GetConnectedPhone() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.waClient.IsLoggedIn() {
		return ""
	}

	jid := c.waClient.Store.ID
	if jid == nil {
		return ""
	}

	return jid.User
}

// RegisterEventHandlers registers event handlers for the WhatsApp client
func (c *Client) RegisterEventHandlers() {
	c.eventHandler = NewEventHandler(c.sessionID, c.userID, c.logger, c.eventHandlers)
	c.waClient.AddEventHandler(c.eventHandler.HandleEvent)

	c.logger.Info("event handlers registered")
}

// parseJID parses a phone number into a WhatsApp JID
func (c *Client) parseJID(phone string) (types.JID, error) {
	// Remove any non-numeric characters except +
	// Phone should be in E.164 format (e.g., +1234567890)
	
	// Check if it's already a JID
	if jid, err := types.ParseJID(phone); err == nil {
		return jid, nil
	}

	// Otherwise, assume it's a phone number and create a JID
	// Remove + if present
	if len(phone) > 0 && phone[0] == '+' {
		phone = phone[1:]
	}

	// Create JID for regular WhatsApp user
	return types.NewJID(phone, types.DefaultUserServer), nil
}

// GetSessionID returns the session ID
func (c *Client) GetSessionID() string {
	return c.sessionID
}

// GetUserID returns the user ID
func (c *Client) GetUserID() string {
	return c.userID
}

// Logout logs out from WhatsApp and removes device data
func (c *Client) Logout(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.waClient.IsLoggedIn() {
		return fmt.Errorf("not logged in")
	}

	// Logout from WhatsApp
	if err := c.waClient.Logout(ctx); err != nil {
		c.logger.Error("failed to logout", logger.Fields{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to logout: %w", err)
	}

	c.connected = false
	c.logger.Info("logged out successfully")

	return nil
}
