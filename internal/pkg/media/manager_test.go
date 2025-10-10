package media

import (
	"bytes"
	"testing"
)

func TestNewManager(t *testing.T) {
	tempDir := t.TempDir()

	manager, err := NewManager(tempDir, 10)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	if manager.storage == nil {
		t.Error("Storage is nil")
	}

	if manager.validator == nil {
		t.Error("Validator is nil")
	}

	if manager.GetMaxSizeMB() != 10 {
		t.Errorf("Expected MaxSizeMB 10, got %d", manager.GetMaxSizeMB())
	}
}

func TestSaveMedia(t *testing.T) {
	tempDir := t.TempDir()
	manager, _ := NewManager(tempDir, 16)

	sessionID := "test-session-123"
	data := []byte("test image data")
	mimeType := "image/jpeg"
	filename := "test.jpg"

	relativePath, err := manager.SaveMedia(sessionID, data, mimeType, filename)
	if err != nil {
		t.Fatalf("Failed to save media: %v", err)
	}

	if relativePath == "" {
		t.Error("Expected non-empty relative path")
	}

	// Verify file can be retrieved
	retrievedData, err := manager.GetMedia(relativePath)
	if err != nil {
		t.Fatalf("Failed to get media: %v", err)
	}

	if string(retrievedData) != string(data) {
		t.Errorf("Expected data %s, got %s", string(data), string(retrievedData))
	}
}

func TestSaveMediaValidationFailure(t *testing.T) {
	tempDir := t.TempDir()
	manager, _ := NewManager(tempDir, 1)

	sessionID := "test-session-123"
	
	tests := []struct {
		name     string
		data     []byte
		mimeType string
		filename string
	}{
		{
			name:     "file too large",
			data:     make([]byte, 2*1024*1024),
			mimeType: "image/jpeg",
			filename: "test.jpg",
		},
		{
			name:     "invalid mime type",
			data:     []byte("test"),
			mimeType: "text/html",
			filename: "test.html",
		},
		{
			name:     "invalid filename",
			data:     []byte("test"),
			mimeType: "image/jpeg",
			filename: "../test.jpg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := manager.SaveMedia(sessionID, tt.data, tt.mimeType, tt.filename)
			if err == nil {
				t.Error("Expected error, got nil")
			}
		})
	}
}

func TestSaveMediaFromReader(t *testing.T) {
	tempDir := t.TempDir()
	manager, _ := NewManager(tempDir, 16)

	sessionID := "test-session-123"
	data := []byte("test image data")
	reader := bytes.NewReader(data)
	mimeType := "image/jpeg"
	filename := "test.jpg"

	relativePath, err := manager.SaveMediaFromReader(sessionID, reader, mimeType, filename)
	if err != nil {
		t.Fatalf("Failed to save media from reader: %v", err)
	}

	if relativePath == "" {
		t.Error("Expected non-empty relative path")
	}

	// Verify file can be retrieved
	retrievedData, err := manager.GetMedia(relativePath)
	if err != nil {
		t.Fatalf("Failed to get media: %v", err)
	}

	if string(retrievedData) != string(data) {
		t.Errorf("Expected data %s, got %s", string(data), string(retrievedData))
	}
}

func TestDeleteMedia(t *testing.T) {
	tempDir := t.TempDir()
	manager, _ := NewManager(tempDir, 16)

	sessionID := "test-session-123"
	data := []byte("test image data")
	mimeType := "image/jpeg"
	filename := "test.jpg"

	// Save media first
	relativePath, _ := manager.SaveMedia(sessionID, data, mimeType, filename)

	// Delete media
	err := manager.DeleteMedia(relativePath)
	if err != nil {
		t.Fatalf("Failed to delete media: %v", err)
	}

	// Verify file was deleted
	_, err = manager.GetMedia(relativePath)
	if err == nil {
		t.Error("Expected error when getting deleted file, got nil")
	}
}

func TestDeleteSessionMedia(t *testing.T) {
	tempDir := t.TempDir()
	manager, _ := NewManager(tempDir, 16)

	sessionID := "test-session-123"

	// Save multiple media files
	manager.SaveMedia(sessionID, []byte("image1"), "image/jpeg", "test1.jpg")
	manager.SaveMedia(sessionID, []byte("image2"), "image/png", "test2.png")
	manager.SaveMedia(sessionID, []byte("video1"), "video/mp4", "test1.mp4")

	// Delete all session media
	err := manager.DeleteSessionMedia(sessionID)
	if err != nil {
		t.Fatalf("Failed to delete session media: %v", err)
	}

	// Verify all files were deleted by trying to delete again (should not error)
	err = manager.DeleteSessionMedia(sessionID)
	if err != nil {
		t.Errorf("Expected no error when deleting non-existent session, got %v", err)
	}
}

func TestGetMediaInfo(t *testing.T) {
	tempDir := t.TempDir()
	manager, _ := NewManager(tempDir, 16)

	sessionID := "test-session-123"
	data := []byte("test image data")
	mimeType := "image/jpeg"
	filename := "test.jpg"

	// Save media first
	relativePath, _ := manager.SaveMedia(sessionID, data, mimeType, filename)

	// Get media info
	info, err := manager.GetMediaInfo(relativePath)
	if err != nil {
		t.Fatalf("Failed to get media info: %v", err)
	}

	if info.Path != relativePath {
		t.Errorf("Expected path %s, got %s", relativePath, info.Path)
	}

	if info.Size != int64(len(data)) {
		t.Errorf("Expected size %d, got %d", len(data), info.Size)
	}
}

func TestValidateMediaFile(t *testing.T) {
	tempDir := t.TempDir()
	manager, _ := NewManager(tempDir, 1)

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
			name:      "file too large",
			data:      make([]byte, 2*1024*1024),
			mimeType:  "image/jpeg",
			filename:  "test.jpg",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.ValidateMediaFile(tt.data, tt.mimeType, tt.filename)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateMediaFile() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestGetMaxSizeBytes(t *testing.T) {
	tempDir := t.TempDir()
	manager, _ := NewManager(tempDir, 10)

	expectedBytes := 10 * 1024 * 1024
	if manager.GetMaxSizeBytes() != expectedBytes {
		t.Errorf("Expected %d bytes, got %d", expectedBytes, manager.GetMaxSizeBytes())
	}
}
