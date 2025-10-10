package media

import (
	"strings"
	"testing"
)

func TestNewValidator(t *testing.T) {
	// Test with custom max size
	validator := NewValidator(ValidationConfig{MaxSizeMB: 10})
	if validator.config.MaxSizeMB != 10 {
		t.Errorf("Expected MaxSizeMB 10, got %d", validator.config.MaxSizeMB)
	}

	// Test with default max size
	validator = NewValidator(ValidationConfig{})
	if validator.config.MaxSizeMB != 16 {
		t.Errorf("Expected default MaxSizeMB 16, got %d", validator.config.MaxSizeMB)
	}
}

func TestValidateSize(t *testing.T) {
	validator := NewValidator(ValidationConfig{MaxSizeMB: 1})

	tests := []struct {
		name      string
		sizeBytes int
		wantError bool
	}{
		{"valid size", 500000, false},
		{"max size", 1024 * 1024, false},
		{"zero size", 0, true},
		{"negative size", -1, true},
		{"exceeds max", 2 * 1024 * 1024, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateSize(tt.sizeBytes)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateSize() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestValidateMimeType(t *testing.T) {
	validator := NewValidator(ValidationConfig{MaxSizeMB: 16})

	tests := []struct {
		name      string
		mimeType  string
		wantError bool
	}{
		{"valid image", "image/jpeg", false},
		{"valid video", "video/mp4", false},
		{"valid audio", "audio/mpeg", false},
		{"valid document", "application/pdf", false},
		{"empty mime type", "", true},
		{"invalid mime type", "application/x-malware", true},
		{"unsupported type", "text/html", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateMimeType(tt.mimeType)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateMimeType() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestValidateFilename(t *testing.T) {
	validator := NewValidator(ValidationConfig{MaxSizeMB: 16})

	tests := []struct {
		name      string
		filename  string
		wantError bool
	}{
		{"valid filename", "document.pdf", false},
		{"filename with spaces", "my document.pdf", false},
		{"empty filename", "", true},
		{"path traversal", "../etc/passwd", true},
		{"with forward slash", "path/to/file.pdf", true},
		{"with backslash", "path\\to\\file.pdf", true},
		{"too long", strings.Repeat("a", 256) + ".pdf", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateFilename(tt.filename)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateFilename() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestValidateFile(t *testing.T) {
	validator := NewValidator(ValidationConfig{MaxSizeMB: 1})

	tests := []struct {
		name      string
		data      []byte
		mimeType  string
		filename  string
		wantError bool
	}{
		{
			name:      "valid file",
			data:      []byte("test data"),
			mimeType:  "image/jpeg",
			filename:  "test.jpg",
			wantError: false,
		},
		{
			name:      "invalid size",
			data:      make([]byte, 2*1024*1024),
			mimeType:  "image/jpeg",
			filename:  "test.jpg",
			wantError: true,
		},
		{
			name:      "invalid mime type",
			data:      []byte("test data"),
			mimeType:  "text/html",
			filename:  "test.html",
			wantError: true,
		},
		{
			name:      "invalid filename",
			data:      []byte("test data"),
			mimeType:  "image/jpeg",
			filename:  "../test.jpg",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateFile(tt.data, tt.mimeType, tt.filename)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateFile() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestDetermineMediaType(t *testing.T) {
	tests := []struct {
		mimeType     string
		expectedType MediaType
		wantError    bool
	}{
		{"image/jpeg", MediaTypeImage, false},
		{"image/png", MediaTypeImage, false},
		{"video/mp4", MediaTypeVideo, false},
		{"audio/mpeg", MediaTypeAudio, false},
		{"application/pdf", MediaTypeDocument, false},
		{"text/plain", MediaTypeDocument, false},
		{"application/x-unknown", "", true},
		{"", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.mimeType, func(t *testing.T) {
			mediaType, err := DetermineMediaType(tt.mimeType)
			if (err != nil) != tt.wantError {
				t.Errorf("DetermineMediaType() error = %v, wantError %v", err, tt.wantError)
			}
			if !tt.wantError && mediaType != tt.expectedType {
				t.Errorf("DetermineMediaType() = %v, expected %v", mediaType, tt.expectedType)
			}
		})
	}
}

func TestIsImageMimeType(t *testing.T) {
	tests := []struct {
		mimeType string
		expected bool
	}{
		{"image/jpeg", true},
		{"image/png", true},
		{"image/gif", true},
		{"video/mp4", false},
		{"application/pdf", false},
	}

	for _, tt := range tests {
		t.Run(tt.mimeType, func(t *testing.T) {
			result := IsImageMimeType(tt.mimeType)
			if result != tt.expected {
				t.Errorf("IsImageMimeType(%s) = %v, expected %v", tt.mimeType, result, tt.expected)
			}
		})
	}
}

func TestIsVideoMimeType(t *testing.T) {
	tests := []struct {
		mimeType string
		expected bool
	}{
		{"video/mp4", true},
		{"video/mpeg", true},
		{"video/quicktime", true},
		{"image/jpeg", false},
		{"application/pdf", false},
	}

	for _, tt := range tests {
		t.Run(tt.mimeType, func(t *testing.T) {
			result := IsVideoMimeType(tt.mimeType)
			if result != tt.expected {
				t.Errorf("IsVideoMimeType(%s) = %v, expected %v", tt.mimeType, result, tt.expected)
			}
		})
	}
}

func TestIsAudioMimeType(t *testing.T) {
	tests := []struct {
		mimeType string
		expected bool
	}{
		{"audio/mpeg", true},
		{"audio/mp3", true},
		{"audio/ogg", true},
		{"video/mp4", false},
		{"application/pdf", false},
	}

	for _, tt := range tests {
		t.Run(tt.mimeType, func(t *testing.T) {
			result := IsAudioMimeType(tt.mimeType)
			if result != tt.expected {
				t.Errorf("IsAudioMimeType(%s) = %v, expected %v", tt.mimeType, result, tt.expected)
			}
		})
	}
}

func TestIsDocumentMimeType(t *testing.T) {
	tests := []struct {
		mimeType string
		expected bool
	}{
		{"application/pdf", true},
		{"text/plain", true},
		{"application/msword", true},
		{"image/jpeg", false},
		{"video/mp4", false},
	}

	for _, tt := range tests {
		t.Run(tt.mimeType, func(t *testing.T) {
			result := IsDocumentMimeType(tt.mimeType)
			if result != tt.expected {
				t.Errorf("IsDocumentMimeType(%s) = %v, expected %v", tt.mimeType, result, tt.expected)
			}
		})
	}
}

func TestIsAllowedMimeType(t *testing.T) {
	tests := []struct {
		mimeType string
		expected bool
	}{
		{"image/jpeg", true},
		{"video/mp4", true},
		{"audio/mpeg", true},
		{"application/pdf", true},
		{"text/html", false},
		{"application/x-executable", false},
	}

	for _, tt := range tests {
		t.Run(tt.mimeType, func(t *testing.T) {
			result := IsAllowedMimeType(tt.mimeType)
			if result != tt.expected {
				t.Errorf("IsAllowedMimeType(%s) = %v, expected %v", tt.mimeType, result, tt.expected)
			}
		})
	}
}

func TestDetectMimeType(t *testing.T) {
	tests := []struct {
		filename     string
		expectedType string
	}{
		{"document.pdf", "application/pdf"},
		{"image.jpg", "image/jpeg"},
		{"image.png", "image/png"},
		{"video.mp4", "video/mp4"},
		{"noextension", "application/octet-stream"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := DetectMimeType(tt.filename)
			if result != tt.expectedType {
				t.Errorf("DetectMimeType(%s) = %s, expected %s", tt.filename, result, tt.expectedType)
			}
		})
	}
}

func TestGetFileExtension(t *testing.T) {
	tests := []struct {
		mimeType  string
		extension string
	}{
		{"image/jpeg", ".jpg"},
		{"image/png", ".png"},
		{"video/mp4", ".mp4"},
		{"application/pdf", ".pdf"},
		{"text/plain", ".txt"},
		{"unknown/type", ".bin"},
	}

	for _, tt := range tests {
		t.Run(tt.mimeType, func(t *testing.T) {
			result := GetFileExtension(tt.mimeType)
			if result != tt.extension {
				t.Errorf("GetFileExtension(%s) = %s, expected %s", tt.mimeType, result, tt.extension)
			}
		})
	}
}
