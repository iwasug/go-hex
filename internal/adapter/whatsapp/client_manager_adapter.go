package whatsapp

import (
	"context"

	"golang-chi-hexagonal-base/internal/domain/service"
)

// ClientManagerAdapter adapts the whatsapp.ClientManager to the service.WhatsAppClientManager interface
type ClientManagerAdapter struct {
	manager *ClientManager
}

// NewClientManagerAdapter creates a new ClientManagerAdapter
func NewClientManagerAdapter(manager *ClientManager) *ClientManagerAdapter {
	return &ClientManagerAdapter{
		manager: manager,
	}
}

// CreateClient creates a new WhatsApp client for a session
func (a *ClientManagerAdapter) CreateClient(ctx context.Context, sessionID, userID string) error {
	return a.manager.CreateClient(ctx, sessionID, userID)
}

// GetClient retrieves a WhatsApp client by session ID
// Returns service.WhatsAppClient interface
func (a *ClientManagerAdapter) GetClient(sessionID string) (service.WhatsAppClient, error) {
	client, err := a.manager.GetClient(sessionID)
	if err != nil {
		return nil, err
	}
	// Wrap client with adapter to satisfy service.WhatsAppClient interface
	return NewClientAdapter(client), nil
}

// DeleteClient removes a WhatsApp client and performs cleanup
func (a *ClientManagerAdapter) DeleteClient(ctx context.Context, sessionID string) error {
	return a.manager.DeleteClient(ctx, sessionID)
}

// RestoreClients restores all connected sessions on application startup
func (a *ClientManagerAdapter) RestoreClients(ctx context.Context) error {
	return a.manager.RestoreClients(ctx)
}
