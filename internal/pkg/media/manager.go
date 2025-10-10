package media

import (
	"fmt"
	"io"
)

// Manager handles media file operations with validation
type Manager struct {
	storage   *Storage
	validator *Validator
}

// NewManager creates a new Manager instance
func NewManager(basePath string, maxSizeMB int) (*Manager, error) {
	// Create storage
	storage, err := NewStorage(basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage: %w", err)
	}

	// Create validator
	validator := NewValidator(ValidationConfig{
		MaxSizeMB: maxSizeMB,
	})

	return &Manager{
		storage:   storage,
		validator: validator,
	}, nil
}

// SaveMedia saves a media file with validation
func (m *Manager) SaveMedia(sessionID string, data []byte, mimeType string, filename string) (string, error) {
	// Validate file
	if err := m.validator.ValidateFile(data, mimeType, filename); err != nil {
		return "", fmt.Errorf("validation failed: %w", err)
	}

	// Determine media type
	mediaType, err := DetermineMediaType(mimeType)
	if err != nil {
		return "", fmt.Errorf("failed to determine media type: %w", err)
	}

	// Save file
	relativePath, err := m.storage.SaveFile(sessionID, mediaType, data, filename)
	if err != nil {
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	return relativePath, nil
}

// SaveMediaFromReader saves a media file from an io.Reader with validation
func (m *Manager) SaveMediaFromReader(sessionID string, reader io.Reader, mimeType string, filename string) (string, error) {
	// Validate MIME type and filename first
	if err := m.validator.ValidateMimeType(mimeType); err != nil {
		return "", fmt.Errorf("validation failed: %w", err)
	}

	if err := m.validator.ValidateFilename(filename); err != nil {
		return "", fmt.Errorf("validation failed: %w", err)
	}

	// Determine media type
	mediaType, err := DetermineMediaType(mimeType)
	if err != nil {
		return "", fmt.Errorf("failed to determine media type: %w", err)
	}

	// Calculate max size in bytes
	maxSizeBytes := int64(m.validator.config.MaxSizeMB * 1024 * 1024)

	// Save file with size limit
	relativePath, err := m.storage.SaveFileFromReader(sessionID, mediaType, reader, filename, maxSizeBytes)
	if err != nil {
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	return relativePath, nil
}

// GetMedia retrieves a media file
func (m *Manager) GetMedia(relativePath string) ([]byte, error) {
	return m.storage.GetFile(relativePath)
}

// DeleteMedia deletes a media file
func (m *Manager) DeleteMedia(relativePath string) error {
	return m.storage.DeleteFile(relativePath)
}

// DeleteSessionMedia deletes all media files for a session
func (m *Manager) DeleteSessionMedia(sessionID string) error {
	return m.storage.DeleteSessionFiles(sessionID)
}

// GetMediaInfo returns information about a media file
func (m *Manager) GetMediaInfo(relativePath string) (*FileInfo, error) {
	return m.storage.GetFileInfo(relativePath)
}

// ValidateMediaFile validates a media file without saving it
func (m *Manager) ValidateMediaFile(data []byte, mimeType string, filename string) error {
	return m.validator.ValidateFile(data, mimeType, filename)
}

// GetMaxSizeMB returns the maximum file size in MB
func (m *Manager) GetMaxSizeMB() int {
	return m.validator.config.MaxSizeMB
}

// GetMaxSizeBytes returns the maximum file size in bytes
func (m *Manager) GetMaxSizeBytes() int {
	return m.validator.config.MaxSizeMB * 1024 * 1024
}
