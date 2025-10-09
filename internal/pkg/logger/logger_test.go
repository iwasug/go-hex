package logger

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestParseLevel(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Level
	}{
		{"debug level", "debug", DEBUG},
		{"info level", "info", INFO},
		{"warn level", "warn", WARN},
		{"error level", "error", ERROR},
		{"unknown defaults to info", "unknown", INFO},
		{"empty defaults to info", "", INFO},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseLevel(tt.input)
			if result != tt.expected {
				t.Errorf("ParseLevel(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestLevelString(t *testing.T) {
	tests := []struct {
		level    Level
		expected string
	}{
		{DEBUG, "DEBUG"},
		{INFO, "INFO"},
		{WARN, "WARN"},
		{ERROR, "ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.level.String()
			if result != tt.expected {
				t.Errorf("Level.String() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestNewFromConfig(t *testing.T) {
	tests := []struct {
		name         string
		levelStr     string
		formatStr    string
		expectedLvl  Level
		expectedFmt  Format
	}{
		{"json format", "info", "json", INFO, JSONFormat},
		{"text format", "debug", "text", DEBUG, TextFormat},
		{"default format", "warn", "", WARN, JSONFormat},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewFromConfig(tt.levelStr, tt.formatStr)
			if logger.level != tt.expectedLvl {
				t.Errorf("level = %v, want %v", logger.level, tt.expectedLvl)
			}
			if logger.format != tt.expectedFmt {
				t.Errorf("format = %v, want %v", logger.format, tt.expectedFmt)
			}
		})
	}
}

func TestLoggerJSONFormat(t *testing.T) {
	var buf bytes.Buffer
	logger := New(DEBUG, JSONFormat, &buf)

	tests := []struct {
		name     string
		logFunc  func(string, ...Fields)
		message  string
		fields   Fields
		level    string
	}{
		{
			name:    "debug log",
			logFunc: logger.Debug,
			message: "debug message",
			fields:  Fields{"key": "value"},
			level:   "DEBUG",
		},
		{
			name:    "info log",
			logFunc: logger.Info,
			message: "info message",
			fields:  Fields{"user_id": "123"},
			level:   "INFO",
		},
		{
			name:    "warn log",
			logFunc: logger.Warn,
			message: "warning message",
			fields:  Fields{"warning": true},
			level:   "WARN",
		},
		{
			name:    "error log",
			logFunc: logger.Error,
			message: "error message",
			fields:  Fields{"error": "something went wrong"},
			level:   "ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			tt.logFunc(tt.message, tt.fields)

			output := buf.String()
			if output == "" {
				t.Fatal("expected log output, got empty string")
			}

			// Parse JSON output
			var entry logEntry
			if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &entry); err != nil {
				t.Fatalf("failed to parse JSON output: %v\nOutput: %s", err, output)
			}

			// Verify fields
			if entry.Level != tt.level {
				t.Errorf("level = %q, want %q", entry.Level, tt.level)
			}
			if entry.Message != tt.message {
				t.Errorf("message = %q, want %q", entry.Message, tt.message)
			}
			if entry.Timestamp == "" {
				t.Error("timestamp is empty")
			}

			// Verify custom fields
			for key, expectedValue := range tt.fields {
				if actualValue, ok := entry.Fields[key]; !ok {
					t.Errorf("field %q not found in output", key)
				} else if actualValue != expectedValue {
					t.Errorf("field %q = %v, want %v", key, actualValue, expectedValue)
				}
			}
		})
	}
}

func TestLoggerTextFormat(t *testing.T) {
	var buf bytes.Buffer
	logger := New(INFO, TextFormat, &buf)

	logger.Info("test message", Fields{"key": "value"})

	output := buf.String()
	if !strings.Contains(output, "INFO") {
		t.Error("output should contain INFO level")
	}
	if !strings.Contains(output, "test message") {
		t.Error("output should contain message")
	}
	if !strings.Contains(output, "key=value") {
		t.Error("output should contain fields")
	}
}

func TestLoggerLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	logger := New(WARN, JSONFormat, &buf)

	// These should not be logged
	logger.Debug("debug message")
	logger.Info("info message")

	if buf.Len() > 0 {
		t.Errorf("expected no output for DEBUG and INFO when level is WARN, got: %s", buf.String())
	}

	// These should be logged
	logger.Warn("warn message")
	if buf.Len() == 0 {
		t.Error("expected output for WARN message")
	}

	buf.Reset()
	logger.Error("error message")
	if buf.Len() == 0 {
		t.Error("expected output for ERROR message")
	}
}

func TestLoggerWithFields(t *testing.T) {
	var buf bytes.Buffer
	logger := New(INFO, JSONFormat, &buf)

	// Create logger with default fields
	contextLogger := logger.WithFields(Fields{
		"request_id": "123",
		"user_id":    "456",
	})

	contextLogger.Info("test message", Fields{"action": "create"})

	output := buf.String()
	var entry logEntry
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &entry); err != nil {
		t.Fatalf("failed to parse JSON output: %v", err)
	}

	// Verify all fields are present
	expectedFields := map[string]interface{}{
		"request_id": "123",
		"user_id":    "456",
		"action":     "create",
	}

	for key, expectedValue := range expectedFields {
		if actualValue, ok := entry.Fields[key]; !ok {
			t.Errorf("field %q not found in output", key)
		} else if actualValue != expectedValue {
			t.Errorf("field %q = %v, want %v", key, actualValue, expectedValue)
		}
	}
}

func TestLoggerWithFieldsChaining(t *testing.T) {
	var buf bytes.Buffer
	logger := New(INFO, JSONFormat, &buf)

	// Chain WithFields calls
	contextLogger := logger.WithFields(Fields{"request_id": "123"})
	nestedLogger := contextLogger.WithFields(Fields{"user_id": "456"})

	nestedLogger.Info("test message")

	output := buf.String()
	var entry logEntry
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &entry); err != nil {
		t.Fatalf("failed to parse JSON output: %v", err)
	}

	// Verify both fields are present
	if entry.Fields["request_id"] != "123" {
		t.Error("request_id field not found or incorrect")
	}
	if entry.Fields["user_id"] != "456" {
		t.Error("user_id field not found or incorrect")
	}
}

func TestLoggerNoFields(t *testing.T) {
	var buf bytes.Buffer
	logger := New(INFO, JSONFormat, &buf)

	logger.Info("message without fields")

	output := buf.String()
	var entry logEntry
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &entry); err != nil {
		t.Fatalf("failed to parse JSON output: %v", err)
	}

	if entry.Message != "message without fields" {
		t.Errorf("message = %q, want %q", entry.Message, "message without fields")
	}
}

func TestSetLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := New(INFO, JSONFormat, &buf)

	// Initially at INFO level, debug should not log
	logger.Debug("debug message")
	if buf.Len() > 0 {
		t.Error("expected no output for DEBUG at INFO level")
	}

	// Change to DEBUG level
	logger.SetLevel(DEBUG)
	logger.Debug("debug message")
	if buf.Len() == 0 {
		t.Error("expected output for DEBUG after setting level to DEBUG")
	}
}

func TestGetLevel(t *testing.T) {
	logger := New(WARN, JSONFormat, nil)
	if logger.GetLevel() != WARN {
		t.Errorf("GetLevel() = %v, want %v", logger.GetLevel(), WARN)
	}
}

func TestMergeFields(t *testing.T) {
	fields1 := Fields{"a": 1, "b": 2}
	fields2 := Fields{"c": 3, "b": 4} // b should override
	fields3 := Fields{"d": 5}

	result := mergeFields(fields1, fields2, fields3)

	expected := Fields{
		"a": 1,
		"b": 4, // overridden by fields2
		"c": 3,
		"d": 5,
	}

	if len(result) != len(expected) {
		t.Errorf("result length = %d, want %d", len(result), len(expected))
	}

	for key, expectedValue := range expected {
		if actualValue, ok := result[key]; !ok {
			t.Errorf("field %q not found in result", key)
		} else if actualValue != expectedValue {
			t.Errorf("field %q = %v, want %v", key, actualValue, expectedValue)
		}
	}
}

func TestLoggerMultipleFields(t *testing.T) {
	var buf bytes.Buffer
	logger := New(INFO, JSONFormat, &buf)

	logger.Info("test message",
		Fields{"field1": "value1"},
		Fields{"field2": "value2"},
		Fields{"field3": "value3"},
	)

	output := buf.String()
	var entry logEntry
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &entry); err != nil {
		t.Fatalf("failed to parse JSON output: %v", err)
	}

	expectedFields := map[string]string{
		"field1": "value1",
		"field2": "value2",
		"field3": "value3",
	}

	for key, expectedValue := range expectedFields {
		if actualValue, ok := entry.Fields[key]; !ok {
			t.Errorf("field %q not found in output", key)
		} else if actualValue != expectedValue {
			t.Errorf("field %q = %v, want %v", key, actualValue, expectedValue)
		}
	}
}
