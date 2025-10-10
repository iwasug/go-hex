package webhook

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"golang-chi-hexagonal-base/internal/pkg/logger"
)

func TestNewClient(t *testing.T) {
	log := logger.New(logger.INFO, logger.JSONFormat, nil)
	
	t.Run("with default config", func(t *testing.T) {
		config := DefaultConfig()
		client := NewClient(config, log)
		
		if client == nil {
			t.Fatal("expected client to be created")
		}
		if client.config.Timeout != 10*time.Second {
			t.Errorf("expected timeout 10s, got %v", client.config.Timeout)
		}
		if client.config.MaxRetries != 3 {
			t.Errorf("expected max retries 3, got %d", client.config.MaxRetries)
		}
	})
	
	t.Run("with custom config", func(t *testing.T) {
		config := Config{
			Timeout:    5 * time.Second,
			MaxRetries: 5,
			RetryDelay: 2 * time.Second,
		}
		client := NewClient(config, log)
		
		if client.config.Timeout != 5*time.Second {
			t.Errorf("expected timeout 5s, got %v", client.config.Timeout)
		}
		if client.config.MaxRetries != 5 {
			t.Errorf("expected max retries 5, got %d", client.config.MaxRetries)
		}
	})
}

func TestSendWebhook_Success(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}
		
		// Read and verify payload
		body, _ := io.ReadAll(r.Body)
		var payload WebhookPayload
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Errorf("failed to unmarshal payload: %v", err)
		}
		
		if payload.Event != "message.received" {
			t.Errorf("expected event message.received, got %s", payload.Event)
		}
		
		// Return success
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()
	
	// Create client
	log := logger.New(logger.INFO, logger.JSONFormat, nil)
	config := Config{
		Timeout:    5 * time.Second,
		MaxRetries: 3,
		RetryDelay: 1 * time.Second,
	}
	client := NewClient(config, log)
	
	// Send webhook
	req := WebhookRequest{
		URL: server.URL,
		Payload: WebhookPayload{
			Event:     "message.received",
			SessionID: "test-session-123",
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"message": "Hello, World!",
			},
		},
	}
	
	ctx := context.Background()
	response := client.SendWebhook(ctx, req)
	
	// Verify response
	if !response.Success {
		t.Errorf("expected success, got failure: %v", response.Error)
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", response.StatusCode)
	}
	if response.Attempts != 1 {
		t.Errorf("expected 1 attempt, got %d", response.Attempts)
	}
}

func TestSendWebhook_WithHMACSignature(t *testing.T) {
	secret := "test-secret-key"
	var receivedSignature string
	
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedSignature = r.Header.Get("X-Webhook-Signature")
		
		// Read payload
		body, _ := io.ReadAll(r.Body)
		
		// Verify signature
		if !VerifyWebhookSignature(body, receivedSignature, secret) {
			t.Error("signature verification failed")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	
	// Create client
	log := logger.New(logger.INFO, logger.JSONFormat, nil)
	client := NewClient(DefaultConfig(), log)
	
	// Send webhook with secret
	req := WebhookRequest{
		URL:    server.URL,
		Secret: secret,
		Payload: WebhookPayload{
			Event:     "message.received",
			SessionID: "test-session-123",
			Timestamp: time.Now(),
			Data:      map[string]interface{}{"test": "data"},
		},
	}
	
	ctx := context.Background()
	response := client.SendWebhook(ctx, req)
	
	// Verify response
	if !response.Success {
		t.Errorf("expected success, got failure: %v", response.Error)
	}
	if receivedSignature == "" {
		t.Error("expected signature to be sent")
	}
}

func TestSendWebhook_RetryOnServerError(t *testing.T) {
	attempts := 0
	
	// Create test server that fails twice then succeeds
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	
	// Create client
	log := logger.New(logger.INFO, logger.JSONFormat, nil)
	config := Config{
		Timeout:    5 * time.Second,
		MaxRetries: 3,
		RetryDelay: 100 * time.Millisecond, // Short delay for testing
	}
	client := NewClient(config, log)
	
	// Send webhook
	req := WebhookRequest{
		URL: server.URL,
		Payload: WebhookPayload{
			Event:     "test.event",
			SessionID: "test-session",
			Timestamp: time.Now(),
			Data:      map[string]interface{}{},
		},
	}
	
	ctx := context.Background()
	response := client.SendWebhook(ctx, req)
	
	// Verify response
	if !response.Success {
		t.Errorf("expected success after retries, got failure: %v", response.Error)
	}
	if response.Attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", response.Attempts)
	}
	if attempts != 3 {
		t.Errorf("expected server to receive 3 requests, got %d", attempts)
	}
}

func TestSendWebhook_FailAfterMaxRetries(t *testing.T) {
	// Create test server that always fails
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()
	
	// Create client
	log := logger.New(logger.INFO, logger.JSONFormat, nil)
	config := Config{
		Timeout:    5 * time.Second,
		MaxRetries: 3,
		RetryDelay: 100 * time.Millisecond,
	}
	client := NewClient(config, log)
	
	// Send webhook
	req := WebhookRequest{
		URL: server.URL,
		Payload: WebhookPayload{
			Event:     "test.event",
			SessionID: "test-session",
			Timestamp: time.Now(),
			Data:      map[string]interface{}{},
		},
	}
	
	ctx := context.Background()
	response := client.SendWebhook(ctx, req)
	
	// Verify response
	if response.Success {
		t.Error("expected failure after max retries")
	}
	if response.Attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", response.Attempts)
	}
	if response.Error == nil {
		t.Error("expected error to be set")
	}
}

func TestSendWebhook_NoRetryOn4xxError(t *testing.T) {
	attempts := 0
	
	// Create test server that returns 400
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()
	
	// Create client
	log := logger.New(logger.INFO, logger.JSONFormat, nil)
	config := Config{
		Timeout:    5 * time.Second,
		MaxRetries: 3,
		RetryDelay: 100 * time.Millisecond,
	}
	client := NewClient(config, log)
	
	// Send webhook
	req := WebhookRequest{
		URL: server.URL,
		Payload: WebhookPayload{
			Event:     "test.event",
			SessionID: "test-session",
			Timestamp: time.Now(),
			Data:      map[string]interface{}{},
		},
	}
	
	ctx := context.Background()
	response := client.SendWebhook(ctx, req)
	
	// Verify response - should not retry on 4xx
	if response.Success {
		t.Error("expected failure")
	}
	if response.Attempts != 1 {
		t.Errorf("expected 1 attempt (no retry on 4xx), got %d", response.Attempts)
	}
	if attempts != 1 {
		t.Errorf("expected server to receive 1 request, got %d", attempts)
	}
}

func TestSendWebhook_ContextCancellation(t *testing.T) {
	// Create test server with delay
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	
	// Create client
	log := logger.New(logger.INFO, logger.JSONFormat, nil)
	config := Config{
		Timeout:    5 * time.Second,
		MaxRetries: 3,
		RetryDelay: 1 * time.Second,
	}
	client := NewClient(config, log)
	
	// Send webhook with cancelled context
	req := WebhookRequest{
		URL: server.URL,
		Payload: WebhookPayload{
			Event:     "test.event",
			SessionID: "test-session",
			Timestamp: time.Now(),
			Data:      map[string]interface{}{},
		},
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	
	response := client.SendWebhook(ctx, req)
	
	// Verify response
	if response.Success {
		t.Error("expected failure due to context cancellation")
	}
}

func TestCalculateBackoff(t *testing.T) {
	log := logger.New(logger.INFO, logger.JSONFormat, nil)
	config := Config{
		Timeout:       5 * time.Second,
		MaxRetries:    3,
		RetryDelay:    5 * time.Second,
		MaxRetryDelay: 60 * time.Second,
	}
	client := NewClient(config, log)
	
	tests := []struct {
		attempt  int
		expected time.Duration
	}{
		{1, 5 * time.Second},   // 5 * 2^0 = 5
		{2, 10 * time.Second},  // 5 * 2^1 = 10
		{3, 20 * time.Second},  // 5 * 2^2 = 20
		{4, 40 * time.Second},  // 5 * 2^3 = 40
		{5, 60 * time.Second},  // 5 * 2^4 = 80, capped at 60
		{6, 60 * time.Second},  // 5 * 2^5 = 160, capped at 60
	}
	
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			delay := client.calculateBackoff(tt.attempt)
			if delay != tt.expected {
				t.Errorf("attempt %d: expected %v, got %v", tt.attempt, tt.expected, delay)
			}
		})
	}
}

func TestVerifyWebhookSignature(t *testing.T) {
	secret := "test-secret-key"
	payload := []byte(`{"event":"test","data":"value"}`)
	
	// Generate valid signature
	log := logger.New(logger.INFO, logger.JSONFormat, nil)
	client := NewClient(DefaultConfig(), log)
	validSignature := client.generateHMACSignature(payload, secret)
	
	t.Run("valid signature", func(t *testing.T) {
		if !VerifyWebhookSignature(payload, validSignature, secret) {
			t.Error("expected signature to be valid")
		}
	})
	
	t.Run("invalid signature", func(t *testing.T) {
		invalidSignature := "invalid-signature"
		if VerifyWebhookSignature(payload, invalidSignature, secret) {
			t.Error("expected signature to be invalid")
		}
	})
	
	t.Run("wrong secret", func(t *testing.T) {
		wrongSecret := "wrong-secret"
		if VerifyWebhookSignature(payload, validSignature, wrongSecret) {
			t.Error("expected signature to be invalid with wrong secret")
		}
	})
	
	t.Run("modified payload", func(t *testing.T) {
		modifiedPayload := []byte(`{"event":"test","data":"modified"}`)
		if VerifyWebhookSignature(modifiedPayload, validSignature, secret) {
			t.Error("expected signature to be invalid with modified payload")
		}
	})
}
