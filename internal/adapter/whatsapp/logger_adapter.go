package whatsapp

import (
	"fmt"

	"golang-chi-hexagonal-base/internal/pkg/logger"
	waLog "go.mau.fi/whatsmeow/util/log"
)

// LoggerAdapter adapts our logger to the whatsmeow logger interface
type LoggerAdapter struct {
	logger logger.LoggerInterface
}

// NewLoggerAdapter creates a new logger adapter
func NewLoggerAdapter(log logger.LoggerInterface) waLog.Logger {
	return &LoggerAdapter{logger: log}
}

// Errorf logs an error message
func (l *LoggerAdapter) Errorf(msg string, args ...interface{}) {
	l.logger.Error(fmt.Sprintf(msg, args...))
}

// Warnf logs a warning message
func (l *LoggerAdapter) Warnf(msg string, args ...interface{}) {
	l.logger.Warn(fmt.Sprintf(msg, args...))
}

// Infof logs an info message
func (l *LoggerAdapter) Infof(msg string, args ...interface{}) {
	l.logger.Info(fmt.Sprintf(msg, args...))
}

// Debugf logs a debug message
func (l *LoggerAdapter) Debugf(msg string, args ...interface{}) {
	l.logger.Debug(fmt.Sprintf(msg, args...))
}

// Sub creates a sub-logger with a specific module name
func (l *LoggerAdapter) Sub(module string) waLog.Logger {
	return &LoggerAdapter{
		logger: l.logger.WithFields(logger.Fields{"module": module}),
	}
}
