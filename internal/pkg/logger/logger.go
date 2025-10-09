package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

// Level represents the log level
type Level int

const (
	// DEBUG level for detailed debugging information
	DEBUG Level = iota
	// INFO level for general information
	INFO
	// WARN level for warning messages
	WARN
	// ERROR level for error messages
	ERROR
)

// String returns the string representation of the log level
func (l Level) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// ParseLevel parses a string into a Level
func ParseLevel(level string) Level {
	switch level {
	case "debug":
		return DEBUG
	case "info":
		return INFO
	case "warn":
		return WARN
	case "error":
		return ERROR
	default:
		return INFO
	}
}

// Format represents the log output format
type Format int

const (
	// JSONFormat outputs logs in JSON format
	JSONFormat Format = iota
	// TextFormat outputs logs in plain text format
	TextFormat
)

// LoggerInterface defines the logging interface
type LoggerInterface interface {
	Debug(message string, fields ...Fields)
	Info(message string, fields ...Fields)
	Warn(message string, fields ...Fields)
	Error(message string, fields ...Fields)
	WithFields(fields Fields) LoggerInterface
}

// Logger is a structured logger
type Logger struct {
	level  Level
	format Format
	output io.Writer
}

// New creates a new Logger instance
func New(level Level, format Format, output io.Writer) *Logger {
	if output == nil {
		output = os.Stdout
	}
	return &Logger{
		level:  level,
		format: format,
		output: output,
	}
}

// NewFromConfig creates a new Logger from configuration strings
func NewFromConfig(levelStr, formatStr string) *Logger {
	level := ParseLevel(levelStr)
	
	format := JSONFormat
	if formatStr == "text" {
		format = TextFormat
	}
	
	return New(level, format, os.Stdout)
}

// Fields represents additional context fields for logging
type Fields map[string]interface{}

// logEntry represents a single log entry
type logEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

// log writes a log entry
func (l *Logger) log(level Level, message string, fields Fields) {
	// Skip if log level is below configured level
	if level < l.level {
		return
	}

	entry := logEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     level.String(),
		Message:   message,
		Fields:    fields,
	}

	var output string
	if l.format == JSONFormat {
		output = l.formatJSON(entry)
	} else {
		output = l.formatText(entry)
	}

	fmt.Fprintln(l.output, output)
}

// formatJSON formats the log entry as JSON
func (l *Logger) formatJSON(entry logEntry) string {
	data, err := json.Marshal(entry)
	if err != nil {
		// Fallback to simple format if JSON marshaling fails
		return fmt.Sprintf(`{"timestamp":"%s","level":"%s","message":"failed to marshal log entry: %v"}`,
			entry.Timestamp, entry.Level, err)
	}
	return string(data)
}

// formatText formats the log entry as plain text
func (l *Logger) formatText(entry logEntry) string {
	output := fmt.Sprintf("[%s] %s: %s", entry.Timestamp, entry.Level, entry.Message)
	
	if len(entry.Fields) > 0 {
		output += " |"
		for key, value := range entry.Fields {
			output += fmt.Sprintf(" %s=%v", key, value)
		}
	}
	
	return output
}

// Debug logs a debug message
func (l *Logger) Debug(message string, fields ...Fields) {
	l.log(DEBUG, message, mergeFields(fields...))
}

// Info logs an info message
func (l *Logger) Info(message string, fields ...Fields) {
	l.log(INFO, message, mergeFields(fields...))
}

// Warn logs a warning message
func (l *Logger) Warn(message string, fields ...Fields) {
	l.log(WARN, message, mergeFields(fields...))
}

// Error logs an error message
func (l *Logger) Error(message string, fields ...Fields) {
	l.log(ERROR, message, mergeFields(fields...))
}

// WithFields returns a new logger with additional fields
func (l *Logger) WithFields(fields Fields) LoggerInterface {
	// Create a new logger that wraps the original with default fields
	return &contextLogger{
		logger:        l,
		defaultFields: fields,
	}
}

// contextLogger wraps a Logger with default fields
type contextLogger struct {
	logger        *Logger
	defaultFields Fields
}

// Debug logs a debug message with default fields
func (c *contextLogger) Debug(message string, fields ...Fields) {
	allFields := mergeFields(c.defaultFields)
	for _, f := range fields {
		for k, v := range f {
			allFields[k] = v
		}
	}
	c.logger.Debug(message, allFields)
}

// Info logs an info message with default fields
func (c *contextLogger) Info(message string, fields ...Fields) {
	allFields := mergeFields(c.defaultFields)
	for _, f := range fields {
		for k, v := range f {
			allFields[k] = v
		}
	}
	c.logger.Info(message, allFields)
}

// Warn logs a warning message with default fields
func (c *contextLogger) Warn(message string, fields ...Fields) {
	allFields := mergeFields(c.defaultFields)
	for _, f := range fields {
		for k, v := range f {
			allFields[k] = v
		}
	}
	c.logger.Warn(message, allFields)
}

// Error logs an error message with default fields
func (c *contextLogger) Error(message string, fields ...Fields) {
	allFields := mergeFields(c.defaultFields)
	for _, f := range fields {
		for k, v := range f {
			allFields[k] = v
		}
	}
	c.logger.Error(message, allFields)
}

// WithFields returns a new logger with additional fields
func (c *contextLogger) WithFields(fields Fields) LoggerInterface {
	allFields := mergeFields(c.defaultFields, fields)
	return c.logger.WithFields(allFields)
}

// mergeFields merges multiple Fields maps into one
func mergeFields(fieldsList ...Fields) Fields {
	result := make(Fields)
	for _, fields := range fieldsList {
		for k, v := range fields {
			result[k] = v
		}
	}
	return result
}

// SetLevel sets the log level
func (l *Logger) SetLevel(level Level) {
	l.level = level
}

// GetLevel returns the current log level
func (l *Logger) GetLevel() Level {
	return l.level
}
