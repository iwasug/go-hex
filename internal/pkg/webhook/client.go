package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"

	"golang-chi-hexagonal-base/internal/pkg/logger"
)

// Client handles webhook HTTP delivery with retry logic
type Client struct {
	httpClient *http.Client
	logger     logger.LoggerInterface
	config     Config
}

// Config holds webhook client configuration
type Config struct {
	Timeout       time.Duration // HTTP request timeout
	MaxRetries    int           // Maximum number of retry attempts
	RetryDelay    time.Duration // Initial retry delay (exponential backoff)
	MaxRetryDelay time.Duration // Maximum retry delay
}

// DefaultConfig returns default webhook configuration
func DefaultConfig() Config {
	return Config{
		Timeout:       10 * time.Second,
		MaxRetries:    3,
		RetryDelay:    5 * time.Second,
		MaxRetryDelay: 60 * time.Second,
	}
}

// NewClient creates a new webhook client
func NewClient(config Config, log logger.LoggerInterface) *Client {
	if config.Timeout == 0 {
		config.Timeout = 10 * time.Second
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = 5 * time.Second
	}
	if config.MaxRetryDelay == 0 {
		config.MaxRetryDelay = 60 * time.Second
	}

	return &Client{
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		logger: log,
		config: config,
	}
}

// WebhookPayload represents the data sent to webhook endpoints
type WebhookPayload struct {
	Event     string                 `json:"event"`
	SessionID string                 `json:"session_id"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// WebhookRequest represents a webhook delivery request
type WebhookRequest struct {
	URL     string
	Secret  string // Optional HMAC secret for authentication
	Payload WebhookPayload
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

// SendWebhook sends a webhook with retry logic and exponential backoff
func (c *Client) SendWebhook(ctx context.Context, req WebhookRequest) *WebhookResponse {
	startTime := time.Now()
	response := &WebhookResponse{
		Success:  false,
		Attempts: 0,
	}

	// Marshal payload to JSON
	payloadBytes, err := json.Marshal(req.Payload)
	if err != nil {
		response.Error = fmt.Errorf("failed to marshal payload: %w", err)
		c.logger.Error("Failed to marshal webhook payload", logger.Fields{
			"error":      err.Error(),
			"session_id": req.Payload.SessionID,
			"event":      req.Payload.Event,
		})
		return response
	}

	// Attempt delivery with retries
	for attempt := 1; attempt <= c.config.MaxRetries; attempt++ {
		response.Attempts = attempt

		// Create HTTP request
		httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, req.URL, bytes.NewReader(payloadBytes))
		if err != nil {
			response.Error = fmt.Errorf("failed to create request: %w", err)
			c.logger.Error("Failed to create webhook request", logger.Fields{
				"error":      err.Error(),
				"url":        req.URL,
				"attempt":    attempt,
				"session_id": req.Payload.SessionID,
			})
			
			// Don't retry on request creation errors
			break
		}

		// Set headers
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("User-Agent", "WhatsApp-Webhook-Client/1.0")
		httpReq.Header.Set("X-Webhook-Event", req.Payload.Event)
		httpReq.Header.Set("X-Webhook-Timestamp", req.Payload.Timestamp.Format(time.RFC3339))

		// Add HMAC signature if secret is provided
		if req.Secret != "" {
			signature := c.generateHMACSignature(payloadBytes, req.Secret)
			httpReq.Header.Set("X-Webhook-Signature", signature)
		}

		// Send request
		c.logger.Info("Sending webhook", logger.Fields{
			"url":        req.URL,
			"event":      req.Payload.Event,
			"session_id": req.Payload.SessionID,
			"attempt":    attempt,
		})

		httpResp, err := c.httpClient.Do(httpReq)
		if err != nil {
			response.Error = fmt.Errorf("request failed: %w", err)
			c.logger.Warn("Webhook request failed", logger.Fields{
				"error":      err.Error(),
				"url":        req.URL,
				"attempt":    attempt,
				"session_id": req.Payload.SessionID,
			})

			// Retry on network errors
			if attempt < c.config.MaxRetries {
				delay := c.calculateBackoff(attempt)
				c.logger.Info("Retrying webhook after delay", logger.Fields{
					"delay_seconds": delay.Seconds(),
					"attempt":       attempt,
					"max_retries":   c.config.MaxRetries,
				})
				
				select {
				case <-time.After(delay):
					continue
				case <-ctx.Done():
					response.Error = ctx.Err()
					break
				}
			}
			continue
		}

		// Read response body
		defer httpResp.Body.Close()
		bodyBytes, _ := io.ReadAll(httpResp.Body)
		response.StatusCode = httpResp.StatusCode
		response.ResponseBody = string(bodyBytes)

		// Check if successful (2xx status code)
		if httpResp.StatusCode >= 200 && httpResp.StatusCode < 300 {
			response.Success = true
			response.Error = nil
			response.Duration = time.Since(startTime)

			c.logger.Info("Webhook delivered successfully", logger.Fields{
				"url":         req.URL,
				"status_code": httpResp.StatusCode,
				"attempt":     attempt,
				"duration_ms": response.Duration.Milliseconds(),
				"session_id":  req.Payload.SessionID,
			})
			return response
		}

		// Log failed attempt
		c.logger.Warn("Webhook delivery failed", logger.Fields{
			"url":           req.URL,
			"status_code":   httpResp.StatusCode,
			"response_body": response.ResponseBody,
			"attempt":       attempt,
			"session_id":    req.Payload.SessionID,
		})

		response.Error = fmt.Errorf("webhook returned status %d", httpResp.StatusCode)

		// Retry on 5xx errors and 429 (rate limit)
		shouldRetry := (httpResp.StatusCode >= 500 && httpResp.StatusCode < 600) || httpResp.StatusCode == 429
		if shouldRetry && attempt < c.config.MaxRetries {
			delay := c.calculateBackoff(attempt)
			c.logger.Info("Retrying webhook after delay", logger.Fields{
				"delay_seconds": delay.Seconds(),
				"attempt":       attempt,
				"max_retries":   c.config.MaxRetries,
			})

			select {
			case <-time.After(delay):
				continue
			case <-ctx.Done():
				response.Error = ctx.Err()
				break
			}
		} else {
			// Don't retry on 4xx errors (except 429)
			break
		}
	}

	response.Duration = time.Since(startTime)

	// Log final failure after all retries
	if !response.Success {
		c.logger.Error("Webhook delivery failed after all retries", logger.Fields{
			"url":         req.URL,
			"attempts":    response.Attempts,
			"max_retries": c.config.MaxRetries,
			"error":       response.Error.Error(),
			"session_id":  req.Payload.SessionID,
			"event":       req.Payload.Event,
		})
	}

	return response
}

// generateHMACSignature generates an HMAC-SHA256 signature for webhook authentication
func (c *Client) generateHMACSignature(payload []byte, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	return hex.EncodeToString(h.Sum(nil))
}

// calculateBackoff calculates the delay for exponential backoff
func (c *Client) calculateBackoff(attempt int) time.Duration {
	// Exponential backoff: delay * 2^(attempt-1)
	delay := float64(c.config.RetryDelay) * math.Pow(2, float64(attempt-1))
	
	// Cap at max retry delay
	if delay > float64(c.config.MaxRetryDelay) {
		delay = float64(c.config.MaxRetryDelay)
	}
	
	return time.Duration(delay)
}

// VerifyWebhookSignature verifies an HMAC signature from a webhook request
// This can be used by webhook receivers to verify authenticity
func VerifyWebhookSignature(payload []byte, signature, secret string) bool {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	expectedSignature := hex.EncodeToString(h.Sum(nil))
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}
