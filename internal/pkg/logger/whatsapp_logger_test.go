package logger

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestNewWhatsAppLogger(t *testing.T) {
	buf := &bytes.Buffer{}
	baseLogger := New(INFO, JSONFormat, buf)
	waLogger := NewWhatsAppLogger(baseLogger)

	if waLogger == nil {
		t.Fatal("Expected WhatsAppLogger to be created")
	}

	// Test that component field is set
	waLogger.LogSessionCreated("session-1", "user-1", "test-session")

	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}

	fields, ok := logEntry["fields"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected fields to be present")
	}

	if fields["component"] != "whatsapp" {
		t.Errorf("Expected component to be 'whatsapp', got %v", fields["component"])
	}
}

func TestWhatsAppLogger_LogSessionCreated(t *testing.T) {
	buf := &bytes.Buffer{}
	baseLogger := New(INFO, JSONFormat, buf)
	waLogger := NewWhatsAppLogger(baseLogger)

	sessionID := "session-123"
	userID := "user-456"
	sessionName := "my-session"

	waLogger.LogSessionCreated(sessionID, userID, sessionName)

	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}

	if logEntry["level"] != "INFO" {
		t.Errorf("Expected level INFO, got %v", logEntry["level"])
	}

	fields, ok := logEntry["fields"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected fields to be present")
	}

	if fields["event"] != "session_created" {
		t.Errorf("Expected event 'session_created', got %v", fields["event"])
	}

	if fields["session_id"] != sessionID {
		t.Errorf("Expected session_id %s, got %v", sessionID, fields["session_id"])
	}

	if fields["user_id"] != userID {
		t.Errorf("Expected user_id %s, got %v", userID, fields["user_id"])
	}

	if fields["session_name"] != sessionName {
		t.Errorf("Expected session_name %s, got %v", sessionName, fields["session_name"])
	}
}

func TestWhatsAppLogger_LogSessionConnected(t *testing.T) {
	buf := &bytes.Buffer{}
	baseLogger := New(INFO, JSONFormat, buf)
	waLogger := NewWhatsAppLogger(baseLogger)

	sessionID := "session-123"
	phoneNumber := "+1234567890"

	waLogger.LogSessionConnected(sessionID, phoneNumber)

	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}

	fields, ok := logEntry["fields"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected fields to be present")
	}

	if fields["event"] != "session_connected" {
		t.Errorf("Expected event 'session_connected', got %v", fields["event"])
	}

	if fields["phone_number"] != phoneNumber {
		t.Errorf("Expected phone_number %s, got %v", phoneNumber, fields["phone_number"])
	}
}

func TestWhatsAppLogger_LogSessionDisconnected(t *testing.T) {
	buf := &bytes.Buffer{}
	baseLogger := New(WARN, JSONFormat, buf)
	waLogger := NewWhatsAppLogger(baseLogger)

	sessionID := "session-123"
	reason := "connection lost"

	waLogger.LogSessionDisconnected(sessionID, reason)

	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}

	if logEntry["level"] != "WARN" {
		t.Errorf("Expected level WARN, got %v", logEntry["level"])
	}

	fields, ok := logEntry["fields"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected fields to be present")
	}

	if fields["event"] != "session_disconnected" {
		t.Errorf("Expected event 'session_disconnected', got %v", fields["event"])
	}

	if fields["reason"] != reason {
		t.Errorf("Expected reason %s, got %v", reason, fields["reason"])
	}
}

func TestWhatsAppLogger_LogQRCodeGenerated(t *testing.T) {
	buf := &bytes.Buffer{}
	baseLogger := New(INFO, JSONFormat, buf)
	waLogger := NewWhatsAppLogger(baseLogger)

	sessionID := "session-123"
	expiryTime := time.Now().Add(60 * time.Second)

	waLogger.LogQRCodeGenerated(sessionID, expiryTime)

	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}

	fields, ok := logEntry["fields"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected fields to be present")
	}

	if fields["event"] != "qr_code_generated" {
		t.Errorf("Expected event 'qr_code_generated', got %v", fields["event"])
	}

	if fields["expiry_time"] == nil {
		t.Error("Expected expiry_time to be present")
	}
}

func TestWhatsAppLogger_LogMessageSent(t *testing.T) {
	buf := &bytes.Buffer{}
	baseLogger := New(INFO, JSONFormat, buf)
	waLogger := NewWhatsAppLogger(baseLogger)

	sessionID := "session-123"
	messageID := "msg-456"
	recipient := "+1234567890"
	duration := 150 * time.Millisecond

	waLogger.LogMessageSent(sessionID, messageID, recipient, duration)

	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}

	fields, ok := logEntry["fields"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected fields to be present")
	}

	if fields["event"] != "message_sent" {
		t.Errorf("Expected event 'message_sent', got %v", fields["event"])
	}

	if fields["message_id"] != messageID {
		t.Errorf("Expected message_id %s, got %v", messageID, fields["message_id"])
	}

	if fields["recipient"] != recipient {
		t.Errorf("Expected recipient %s, got %v", recipient, fields["recipient"])
	}

	durationMs, ok := fields["duration_ms"].(float64)
	if !ok || durationMs != 150 {
		t.Errorf("Expected duration_ms 150, got %v", fields["duration_ms"])
	}
}

func TestWhatsAppLogger_LogMessageSendFailed(t *testing.T) {
	buf := &bytes.Buffer{}
	baseLogger := New(ERROR, JSONFormat, buf)
	waLogger := NewWhatsAppLogger(baseLogger)

	sessionID := "session-123"
	messageID := "msg-456"
	recipient := "+1234567890"
	err := errors.New("network timeout")

	waLogger.LogMessageSendFailed(sessionID, messageID, recipient, err)

	var logEntry map[string]interface{}
	if jsonErr := json.Unmarshal(buf.Bytes(), &logEntry); jsonErr != nil {
		t.Fatalf("Failed to parse log output: %v", jsonErr)
	}

	if logEntry["level"] != "ERROR" {
		t.Errorf("Expected level ERROR, got %v", logEntry["level"])
	}

	fields, ok := logEntry["fields"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected fields to be present")
	}

	if fields["event"] != "message_send_failed" {
		t.Errorf("Expected event 'message_send_failed', got %v", fields["event"])
	}

	if fields["error"] != err.Error() {
		t.Errorf("Expected error %s, got %v", err.Error(), fields["error"])
	}
}

func TestWhatsAppLogger_LogWebhookDeliverySuccess(t *testing.T) {
	buf := &bytes.Buffer{}
	baseLogger := New(INFO, JSONFormat, buf)
	waLogger := NewWhatsAppLogger(baseLogger)

	sessionID := "session-123"
	messageID := "msg-456"
	webhookURL := "https://example.com/webhook"
	statusCode := 200
	duration := 250 * time.Millisecond

	waLogger.LogWebhookDeliverySuccess(sessionID, messageID, webhookURL, statusCode, duration)

	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}

	fields, ok := logEntry["fields"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected fields to be present")
	}

	if fields["event"] != "webhook_delivery_success" {
		t.Errorf("Expected event 'webhook_delivery_success', got %v", fields["event"])
	}

	if fields["webhook_url"] != webhookURL {
		t.Errorf("Expected webhook_url %s, got %v", webhookURL, fields["webhook_url"])
	}

	statusCodeFloat, ok := fields["status_code"].(float64)
	if !ok || int(statusCodeFloat) != statusCode {
		t.Errorf("Expected status_code %d, got %v", statusCode, fields["status_code"])
	}
}

func TestWhatsAppLogger_LogWebhookDeliveryFailed(t *testing.T) {
	buf := &bytes.Buffer{}
	baseLogger := New(WARN, JSONFormat, buf)
	waLogger := NewWhatsAppLogger(baseLogger)

	sessionID := "session-123"
	messageID := "msg-456"
	webhookURL := "https://example.com/webhook"
	attempt := 2
	statusCode := 500
	err := errors.New("internal server error")

	waLogger.LogWebhookDeliveryFailed(sessionID, messageID, webhookURL, attempt, statusCode, err)

	var logEntry map[string]interface{}
	if jsonErr := json.Unmarshal(buf.Bytes(), &logEntry); jsonErr != nil {
		t.Fatalf("Failed to parse log output: %v", jsonErr)
	}

	if logEntry["level"] != "WARN" {
		t.Errorf("Expected level WARN, got %v", logEntry["level"])
	}

	fields, ok := logEntry["fields"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected fields to be present")
	}

	if fields["event"] != "webhook_delivery_failed" {
		t.Errorf("Expected event 'webhook_delivery_failed', got %v", fields["event"])
	}

	attemptFloat, ok := fields["attempt"].(float64)
	if !ok || int(attemptFloat) != attempt {
		t.Errorf("Expected attempt %d, got %v", attempt, fields["attempt"])
	}
}

func TestWhatsAppLogger_LogReconnectionSuccess(t *testing.T) {
	buf := &bytes.Buffer{}
	baseLogger := New(INFO, JSONFormat, buf)
	waLogger := NewWhatsAppLogger(baseLogger)

	sessionID := "session-123"
	attempt := 3
	duration := 5 * time.Second

	waLogger.LogReconnectionSuccess(sessionID, attempt, duration)

	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}

	fields, ok := logEntry["fields"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected fields to be present")
	}

	if fields["event"] != "reconnection_success" {
		t.Errorf("Expected event 'reconnection_success', got %v", fields["event"])
	}

	attemptFloat, ok := fields["attempt"].(float64)
	if !ok || int(attemptFloat) != attempt {
		t.Errorf("Expected attempt %d, got %v", attempt, fields["attempt"])
	}

	durationMs, ok := fields["duration_ms"].(float64)
	if !ok || durationMs != 5000 {
		t.Errorf("Expected duration_ms 5000, got %v", fields["duration_ms"])
	}
}

func TestWhatsAppLogger_LogErrorWithContext(t *testing.T) {
	buf := &bytes.Buffer{}
	baseLogger := New(ERROR, JSONFormat, buf)
	waLogger := NewWhatsAppLogger(baseLogger)

	operation := "send_message"
	err := errors.New("network error")
	context := map[string]interface{}{
		"session_id": "session-123",
		"recipient":  "+1234567890",
		"retry":      3,
	}

	waLogger.LogErrorWithContext(operation, err, context)

	var logEntry map[string]interface{}
	if jsonErr := json.Unmarshal(buf.Bytes(), &logEntry); jsonErr != nil {
		t.Fatalf("Failed to parse log output: %v", jsonErr)
	}

	if logEntry["level"] != "ERROR" {
		t.Errorf("Expected level ERROR, got %v", logEntry["level"])
	}

	fields, ok := logEntry["fields"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected fields to be present")
	}

	if fields["event"] != "error" {
		t.Errorf("Expected event 'error', got %v", fields["event"])
	}

	if fields["operation"] != operation {
		t.Errorf("Expected operation %s, got %v", operation, fields["operation"])
	}

	if fields["error"] != err.Error() {
		t.Errorf("Expected error %s, got %v", err.Error(), fields["error"])
	}

	if fields["session_id"] != context["session_id"] {
		t.Errorf("Expected session_id from context, got %v", fields["session_id"])
	}

	retryFloat, ok := fields["retry"].(float64)
	if !ok || int(retryFloat) != 3 {
		t.Errorf("Expected retry 3, got %v", fields["retry"])
	}
}

func TestWhatsAppLogger_WithSessionContext(t *testing.T) {
	buf := &bytes.Buffer{}
	baseLogger := New(INFO, JSONFormat, buf)
	waLogger := NewWhatsAppLogger(baseLogger)

	sessionID := "session-123"
	userID := "user-456"

	contextLogger := waLogger.WithSessionContext(sessionID, userID)
	contextLogger.LogMessageSending(sessionID, "msg-1", "+1234567890", "text")

	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}

	fields, ok := logEntry["fields"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected fields to be present")
	}

	if fields["session_id"] != sessionID {
		t.Errorf("Expected session_id %s, got %v", sessionID, fields["session_id"])
	}

	if fields["user_id"] != userID {
		t.Errorf("Expected user_id %s, got %v", userID, fields["user_id"])
	}
}

func TestWhatsAppLogger_TextFormat(t *testing.T) {
	buf := &bytes.Buffer{}
	baseLogger := New(INFO, TextFormat, buf)
	waLogger := NewWhatsAppLogger(baseLogger)

	waLogger.LogSessionCreated("session-1", "user-1", "test-session")

	output := buf.String()
	if !strings.Contains(output, "INFO") {
		t.Error("Expected log output to contain INFO level")
	}

	if !strings.Contains(output, "WhatsApp session created") {
		t.Error("Expected log output to contain message")
	}

	if !strings.Contains(output, "component=whatsapp") {
		t.Error("Expected log output to contain component field")
	}

	if !strings.Contains(output, "session_id=session-1") {
		t.Error("Expected log output to contain session_id field")
	}
}

func TestWhatsAppLogger_LogLevels(t *testing.T) {
	tests := []struct {
		name      string
		logLevel  Level
		logFunc   func(*WhatsAppLogger)
		shouldLog bool
	}{
		{
			name:     "DEBUG level logs debug",
			logLevel: DEBUG,
			logFunc: func(l *WhatsAppLogger) {
				l.LogMessageStored("s1", "m1", "inbound")
			},
			shouldLog: true,
		},
		{
			name:     "INFO level skips debug",
			logLevel: INFO,
			logFunc: func(l *WhatsAppLogger) {
				l.LogMessageStored("s1", "m1", "inbound")
			},
			shouldLog: false,
		},
		{
			name:     "INFO level logs info",
			logLevel: INFO,
			logFunc: func(l *WhatsAppLogger) {
				l.LogSessionCreated("s1", "u1", "test")
			},
			shouldLog: true,
		},
		{
			name:     "ERROR level logs error",
			logLevel: ERROR,
			logFunc: func(l *WhatsAppLogger) {
				l.LogMessageSendFailed("s1", "m1", "+123", errors.New("test"))
			},
			shouldLog: true,
		},
		{
			name:     "ERROR level skips info",
			logLevel: ERROR,
			logFunc: func(l *WhatsAppLogger) {
				l.LogSessionCreated("s1", "u1", "test")
			},
			shouldLog: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			baseLogger := New(tt.logLevel, JSONFormat, buf)
			waLogger := NewWhatsAppLogger(baseLogger)

			tt.logFunc(waLogger)

			hasOutput := buf.Len() > 0
			if hasOutput != tt.shouldLog {
				t.Errorf("Expected shouldLog=%v, but hasOutput=%v", tt.shouldLog, hasOutput)
			}
		})
	}
}
