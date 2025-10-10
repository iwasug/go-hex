package whatsapp

import (
	"context"
	"fmt"
	"time"
)

// ClientAdapter adapts the whatsapp.Client to the service.WhatsAppClient interface
type ClientAdapter struct {
	client *Client
}

// NewClientAdapter creates a new ClientAdapter
func NewClientAdapter(client *Client) *ClientAdapter {
	return &ClientAdapter{
		client: client,
	}
}

// Connect establishes a WebSocket connection to WhatsApp
func (a *ClientAdapter) Connect() error {
	return a.client.Connect()
}

// Disconnect gracefully disconnects from WhatsApp
func (a *ClientAdapter) Disconnect() error {
	return a.client.Disconnect()
}

// GetQRCode generates and returns a QR code as base64-encoded PNG
func (a *ClientAdapter) GetQRCode(ctx context.Context, timeout time.Duration) (string, error) {
	return a.client.GetQRCode(ctx, timeout)
}

// SendTextMessage sends a text message to a recipient
func (a *ClientAdapter) SendTextMessage(ctx context.Context, to, content string) (string, error) {
	return a.client.SendTextMessage(ctx, to, content)
}

// SendMediaMessage sends a media message to a recipient
// Accepts interface{} and converts to whatsapp.MediaMessage
func (a *ClientAdapter) SendMediaMessage(ctx context.Context, to string, media interface{}) (string, error) {
	// Convert interface{} to MediaMessage
	var mediaMsg MediaMessage
	
	switch m := media.(type) {
	case MediaMessage:
		mediaMsg = m
	case map[string]interface{}:
		// Handle map conversion if needed
		mediaMsg = MediaMessage{
			Type:     MediaType(fmt.Sprint(m["Type"])),
			Data:     m["Data"].([]byte),
			MimeType: fmt.Sprint(m["MimeType"]),
			Caption:  fmt.Sprint(m["Caption"]),
			FileName: fmt.Sprint(m["FileName"]),
		}
	default:
		return "", fmt.Errorf("unsupported media type: %T", media)
	}
	
	return a.client.SendMediaMessage(ctx, to, mediaMsg)
}

// IsConnected returns whether the client is connected
func (a *ClientAdapter) IsConnected() bool {
	return a.client.IsConnected()
}

// IsLoggedIn returns whether the client is logged in
func (a *ClientAdapter) IsLoggedIn() bool {
	return a.client.IsLoggedIn()
}

// GetConnectedPhone returns the connected phone number
func (a *ClientAdapter) GetConnectedPhone() string {
	return a.client.GetConnectedPhone()
}

// GetSessionID returns the session ID
func (a *ClientAdapter) GetSessionID() string {
	return a.client.GetSessionID()
}

// GetUserID returns the user ID
func (a *ClientAdapter) GetUserID() string {
	return a.client.GetUserID()
}
