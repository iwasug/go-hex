package media

import (
	"fmt"
	"mime"
	"path/filepath"
	"strings"
)

// ValidationConfig holds configuration for media validation
type ValidationConfig struct {
	MaxSizeMB int // Maximum file size in megabytes
}

// Validator handles media file validation
type Validator struct {
	config ValidationConfig
}

// NewValidator creates a new Validator instance
func NewValidator(config ValidationConfig) *Validator {
	// Set default max size if not provided
	if config.MaxSizeMB <= 0 {
		config.MaxSizeMB = 16 // Default 16MB
	}

	return &Validator{
		config: config,
	}
}

// ValidateFile validates a media file
func (v *Validator) ValidateFile(data []byte, mimeType string, filename string) error {
	// Validate file size
	if err := v.ValidateSize(len(data)); err != nil {
		return err
	}

	// Validate MIME type
	if err := v.ValidateMimeType(mimeType); err != nil {
		return err
	}

	// Validate filename
	if err := v.ValidateFilename(filename); err != nil {
		return err
	}

	return nil
}

// ValidateSize validates file size
func (v *Validator) ValidateSize(sizeBytes int) error {
	maxSizeBytes := v.config.MaxSizeMB * 1024 * 1024
	
	if sizeBytes <= 0 {
		return fmt.Errorf("file size must be greater than 0")
	}

	if sizeBytes > maxSizeBytes {
		return fmt.Errorf("file size (%d bytes) exceeds maximum allowed size of %dMB", sizeBytes, v.config.MaxSizeMB)
	}

	return nil
}

// ValidateMimeType validates MIME type
func (v *Validator) ValidateMimeType(mimeType string) error {
	if mimeType == "" {
		return fmt.Errorf("MIME type is required")
	}

	// Check if MIME type is in allowed list
	if !IsAllowedMimeType(mimeType) {
		return fmt.Errorf("MIME type %s is not allowed", mimeType)
	}

	return nil
}

// ValidateFilename validates filename
func (v *Validator) ValidateFilename(filename string) error {
	if filename == "" {
		return fmt.Errorf("filename is required")
	}

	// Check for path traversal attempts
	if strings.Contains(filename, "..") {
		return fmt.Errorf("filename contains invalid characters")
	}

	if strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		return fmt.Errorf("filename cannot contain path separators")
	}

	// Check filename length
	if len(filename) > 255 {
		return fmt.Errorf("filename exceeds maximum length of 255 characters")
	}

	return nil
}

// DetermineMediaType determines the media type from MIME type
func DetermineMediaType(mimeType string) (MediaType, error) {
	// Normalize MIME type
	mimeType = strings.ToLower(strings.TrimSpace(mimeType))

	// Check each media type
	if IsImageMimeType(mimeType) {
		return MediaTypeImage, nil
	}
	if IsVideoMimeType(mimeType) {
		return MediaTypeVideo, nil
	}
	if IsAudioMimeType(mimeType) {
		return MediaTypeAudio, nil
	}
	if IsDocumentMimeType(mimeType) {
		return MediaTypeDocument, nil
	}

	return "", fmt.Errorf("unsupported MIME type: %s", mimeType)
}

// IsAllowedMimeType checks if a MIME type is allowed
func IsAllowedMimeType(mimeType string) bool {
	mimeType = strings.ToLower(strings.TrimSpace(mimeType))
	
	return IsImageMimeType(mimeType) ||
		IsVideoMimeType(mimeType) ||
		IsAudioMimeType(mimeType) ||
		IsDocumentMimeType(mimeType)
}

// IsImageMimeType checks if MIME type is an image
func IsImageMimeType(mimeType string) bool {
	allowedImageTypes := []string{
		"image/jpeg",
		"image/jpg",
		"image/png",
		"image/gif",
		"image/webp",
		"image/bmp",
		"image/tiff",
	}

	for _, allowed := range allowedImageTypes {
		if mimeType == allowed {
			return true
		}
	}

	return false
}

// IsVideoMimeType checks if MIME type is a video
func IsVideoMimeType(mimeType string) bool {
	allowedVideoTypes := []string{
		"video/mp4",
		"video/mpeg",
		"video/3gpp",
		"video/quicktime",
		"video/x-matroska",
		"video/webm",
		"video/x-msvideo",
	}

	for _, allowed := range allowedVideoTypes {
		if mimeType == allowed {
			return true
		}
	}

	return false
}

// IsAudioMimeType checks if MIME type is audio
func IsAudioMimeType(mimeType string) bool {
	allowedAudioTypes := []string{
		"audio/aac",
		"audio/mp4",
		"audio/mpeg",
		"audio/mp3",
		"audio/amr",
		"audio/ogg",
		"audio/opus",
		"audio/wav",
		"audio/webm",
	}

	for _, allowed := range allowedAudioTypes {
		if mimeType == allowed {
			return true
		}
	}

	return false
}

// IsDocumentMimeType checks if MIME type is a document
func IsDocumentMimeType(mimeType string) bool {
	allowedDocumentTypes := []string{
		"application/pdf",
		"application/msword",
		"application/vnd.ms-excel",
		"application/vnd.ms-powerpoint",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		"application/vnd.openxmlformats-officedocument.presentationml.presentation",
		"text/plain",
		"text/csv",
		"application/zip",
		"application/x-zip-compressed",
		"application/x-rar-compressed",
		"application/x-7z-compressed",
		"application/json",
		"application/xml",
		"text/xml",
	}

	for _, allowed := range allowedDocumentTypes {
		if mimeType == allowed {
			return true
		}
	}

	return false
}

// DetectMimeType attempts to detect MIME type from filename extension
func DetectMimeType(filename string) string {
	ext := filepath.Ext(filename)
	if ext == "" {
		return "application/octet-stream"
	}

	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		return "application/octet-stream"
	}

	return mimeType
}

// GetFileExtension returns the file extension for a MIME type
func GetFileExtension(mimeType string) string {
	// Common MIME type to extension mappings
	extensions := map[string]string{
		"image/jpeg":      ".jpg",
		"image/jpg":       ".jpg",
		"image/png":       ".png",
		"image/gif":       ".gif",
		"image/webp":      ".webp",
		"image/bmp":       ".bmp",
		"image/tiff":      ".tiff",
		"video/mp4":       ".mp4",
		"video/mpeg":      ".mpeg",
		"video/3gpp":      ".3gp",
		"video/quicktime": ".mov",
		"video/webm":      ".webm",
		"audio/aac":       ".aac",
		"audio/mp4":       ".m4a",
		"audio/mpeg":      ".mp3",
		"audio/mp3":       ".mp3",
		"audio/ogg":       ".ogg",
		"audio/opus":      ".opus",
		"audio/wav":       ".wav",
		"application/pdf": ".pdf",
		"text/plain":      ".txt",
		"text/csv":        ".csv",
		"application/zip": ".zip",
		"application/json": ".json",
		"application/xml":  ".xml",
	}

	if ext, ok := extensions[strings.ToLower(mimeType)]; ok {
		return ext
	}

	// Try to get extension from mime package
	exts, err := mime.ExtensionsByType(mimeType)
	if err == nil && len(exts) > 0 {
		return exts[0]
	}

	return ".bin"
}
