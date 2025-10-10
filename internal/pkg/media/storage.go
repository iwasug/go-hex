package media

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// MediaType represents the type of media file
type MediaType string

const (
	MediaTypeImage    MediaType = "image"
	MediaTypeVideo    MediaType = "video"
	MediaTypeAudio    MediaType = "audio"
	MediaTypeDocument MediaType = "document"
)

// Storage handles media file storage operations
type Storage struct {
	basePath string
}

// NewStorage creates a new Storage instance
func NewStorage(basePath string) (*Storage, error) {
	// Create base directory if it doesn't exist
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	return &Storage{
		basePath: basePath,
	}, nil
}

// SaveFile saves a media file to storage with unique naming
// Returns the relative file path from basePath
func (s *Storage) SaveFile(sessionID string, mediaType MediaType, data []byte, originalFilename string) (string, error) {
	// Create session directory structure: basePath/sessionID/mediaType/
	sessionDir := filepath.Join(s.basePath, sessionID, string(mediaType))
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create session directory: %w", err)
	}

	// Generate unique filename using UUID and timestamp
	uniqueID := uuid.New().String()
	timestamp := time.Now().Unix()
	
	// Extract file extension from original filename
	ext := filepath.Ext(originalFilename)
	if ext == "" {
		ext = ".bin" // Default extension if none provided
	}

	// Create unique filename: timestamp_uuid_originalname
	safeFilename := sanitizeFilename(originalFilename)
	uniqueFilename := fmt.Sprintf("%d_%s_%s", timestamp, uniqueID[:8], safeFilename)

	// Full file path
	filePath := filepath.Join(sessionDir, uniqueFilename)

	// Write file to disk
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	// Return relative path from basePath
	relativePath := filepath.Join(sessionID, string(mediaType), uniqueFilename)
	return relativePath, nil
}

// SaveFileFromReader saves a media file from an io.Reader
func (s *Storage) SaveFileFromReader(sessionID string, mediaType MediaType, reader io.Reader, originalFilename string, maxSizeBytes int64) (string, error) {
	// Create session directory structure
	sessionDir := filepath.Join(s.basePath, sessionID, string(mediaType))
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create session directory: %w", err)
	}

	// Generate unique filename
	uniqueID := uuid.New().String()
	timestamp := time.Now().Unix()
	
	ext := filepath.Ext(originalFilename)
	if ext == "" {
		ext = ".bin"
	}

	safeFilename := sanitizeFilename(originalFilename)
	uniqueFilename := fmt.Sprintf("%d_%s_%s", timestamp, uniqueID[:8], safeFilename)

	// Full file path
	filePath := filepath.Join(sessionDir, uniqueFilename)

	// Create file
	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Copy data with size limit
	written, err := io.CopyN(file, reader, maxSizeBytes+1)
	if err != nil && err != io.EOF {
		os.Remove(filePath) // Clean up on error
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	// Check if file exceeds size limit
	if written > maxSizeBytes {
		os.Remove(filePath) // Clean up oversized file
		return "", fmt.Errorf("file size exceeds maximum allowed size of %d bytes", maxSizeBytes)
	}

	// Return relative path
	relativePath := filepath.Join(sessionID, string(mediaType), uniqueFilename)
	return relativePath, nil
}

// GetFile retrieves a file from storage
func (s *Storage) GetFile(relativePath string) ([]byte, error) {
	filePath := filepath.Join(s.basePath, relativePath)
	
	// Security check: ensure path is within basePath
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("invalid file path: %w", err)
	}
	
	absBasePath, err := filepath.Abs(s.basePath)
	if err != nil {
		return nil, fmt.Errorf("invalid base path: %w", err)
	}
	
	if !strings.HasPrefix(absPath, absBasePath) {
		return nil, fmt.Errorf("access denied: path outside storage directory")
	}

	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", relativePath)
		}
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return data, nil
}

// DeleteFile deletes a file from storage
func (s *Storage) DeleteFile(relativePath string) error {
	filePath := filepath.Join(s.basePath, relativePath)
	
	// Security check: ensure path is within basePath
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("invalid file path: %w", err)
	}
	
	absBasePath, err := filepath.Abs(s.basePath)
	if err != nil {
		return fmt.Errorf("invalid base path: %w", err)
	}
	
	if !strings.HasPrefix(absPath, absBasePath) {
		return fmt.Errorf("access denied: path outside storage directory")
	}

	// Delete file
	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			return nil // File already deleted, not an error
		}
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// DeleteSessionFiles deletes all files for a session
func (s *Storage) DeleteSessionFiles(sessionID string) error {
	sessionDir := filepath.Join(s.basePath, sessionID)
	
	// Check if directory exists
	if _, err := os.Stat(sessionDir); os.IsNotExist(err) {
		return nil // Directory doesn't exist, nothing to delete
	}

	// Remove entire session directory
	if err := os.RemoveAll(sessionDir); err != nil {
		return fmt.Errorf("failed to delete session directory: %w", err)
	}

	return nil
}

// GetFileInfo returns information about a stored file
func (s *Storage) GetFileInfo(relativePath string) (*FileInfo, error) {
	filePath := filepath.Join(s.basePath, relativePath)
	
	// Security check
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("invalid file path: %w", err)
	}
	
	absBasePath, err := filepath.Abs(s.basePath)
	if err != nil {
		return nil, fmt.Errorf("invalid base path: %w", err)
	}
	
	if !strings.HasPrefix(absPath, absBasePath) {
		return nil, fmt.Errorf("access denied: path outside storage directory")
	}

	// Get file info
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", relativePath)
		}
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	return &FileInfo{
		Path:         relativePath,
		Size:         info.Size(),
		ModifiedTime: info.ModTime(),
	}, nil
}

// CalculateChecksum calculates SHA256 checksum of file data
func (s *Storage) CalculateChecksum(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// FileInfo represents information about a stored file
type FileInfo struct {
	Path         string
	Size         int64
	ModifiedTime time.Time
}

// sanitizeFilename removes potentially dangerous characters from filename
func sanitizeFilename(filename string) string {
	// Remove path separators and other dangerous characters
	filename = strings.ReplaceAll(filename, "/", "_")
	filename = strings.ReplaceAll(filename, "\\", "_")
	filename = strings.ReplaceAll(filename, "\x00", "_")
	
	// Replace consecutive dots with single underscore (but preserve single dots for extensions)
	for strings.Contains(filename, "..") {
		filename = strings.ReplaceAll(filename, "..", "__")
	}
	
	// Limit filename length (keep extension)
	maxLen := 200
	if len(filename) > maxLen {
		ext := filepath.Ext(filename)
		nameWithoutExt := strings.TrimSuffix(filename, ext)
		if len(nameWithoutExt) > maxLen-len(ext) {
			nameWithoutExt = nameWithoutExt[:maxLen-len(ext)]
		}
		filename = nameWithoutExt + ext
	}
	
	return filename
}
