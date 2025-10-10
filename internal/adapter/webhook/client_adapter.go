package webhook

import (
	"context"

	"golang-chi-hexagonal-base/internal/domain/service"
	"golang-chi-hexagonal-base/internal/pkg/logger"
	"golang-chi-hexagonal-base/internal/pkg/webhook"
)

// ClientAdapter adapts the webhook.Client to the service.WebhookClient interface
type ClientAdapter struct {
	client *webhook.Client
}

// NewClientAdapter creates a new webhook client adapter
func NewClientAdapter(config webhook.Config, log logger.LoggerInterface) *ClientAdapter {
	return &ClientAdapter{
		client: webhook.NewClient(config, log),
	}
}

// SendWebhook sends a webhook using the underlying webhook client
func (a *ClientAdapter) SendWebhook(ctx context.Context, req service.WebhookRequest) *service.WebhookResponse {
	// Convert service request to webhook package request
	webhookReq := webhook.WebhookRequest{
		URL:    req.URL,
		Secret: req.Secret,
		Payload: webhook.WebhookPayload{
			Event:     req.Payload.Event,
			SessionID: req.Payload.SessionID,
			Timestamp: req.Payload.Timestamp,
			Data:      req.Payload.Data,
		},
	}

	// Send webhook
	webhookResp := a.client.SendWebhook(ctx, webhookReq)

	// Convert webhook package response to service response
	return &service.WebhookResponse{
		Success:      webhookResp.Success,
		StatusCode:   webhookResp.StatusCode,
		ResponseBody: webhookResp.ResponseBody,
		Error:        webhookResp.Error,
		Attempts:     webhookResp.Attempts,
		Duration:     webhookResp.Duration,
	}
}
