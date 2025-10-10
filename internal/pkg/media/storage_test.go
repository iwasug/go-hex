package media

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewStorage(t *testing.T) {
	// Create temp directory for testing
	tempDir := t.TempDir()

	storage, err := NewStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	if storage.basePath != tempDir {
		t.Errorf("Expected basePath %s, got %s", tempDir, storage.basePath)
	}

	// Verify directory was created
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		t.Error("Base directory was not created")
	}
}

func TestSaveFile(t *testing.T) {
	tempDir := t.TempDir()
	storage, _ := NewStorage(tempDir)

	sessionID := "test-session-123"
	mediaType := MediaTypeImage
	data := []byte("test image data")
	filename := "test.jpg"

	relativePath, err := storage.SaveFile(sessionID, mediaType, data, filename)
	if err != nil {
		t.Fatalf("Failed to save file: %v", err)
	}

	// Verify file was created
	fullPath := filepath.Join(tempDir, relativePath)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		t.Error("File was not created")
	}

	// Verify file content
	savedData, err := os.ReadFile(fullPath)
	if err != nil {
		t.Fatalf("Failed to read saved file: %v", err)
	}

	if string(savedData) != string(data) {
		t.Errorf("Expected data %s, got %s", string(data), string(savedData))
	}

	// Verify relative path structure
	expectedPrefix := filepath.Join(sessionID, string(mediaType))
	if !filepath.HasPrefix(relativePath, expectedPrefix) {
		t.Errorf("Expected path to start with %s, got %s", expectedPrefix, relativePath)
	}
}

func TestGetFile(t *testing.T) {
	tempDir := t.TempDir()
	storage, _ := NewStorage(tempDir)

	sessionID := "test-session-123"
	mediaType := MediaTypeImage
	data := []byte("test image data")
	filename := "test.jpg"

	// Save file first
	relativePath, _ := storage.SaveFile(sessionID, mediaType, data, filename)

	// Retrieve file
	retrievedData, err := storage.GetFile(relativePath)
	if err != nil {
		t.Fatalf("Failed to get file: %v", err)
	}

	if string(retrievedData) != string(data) {
		t.Errorf("Expected data %s, got %s", string(data), string(retrievedData))
	}
}

func TestGetFileNotFound(t *testing.T) {
	tempDir := t.TempDir()
	storage, _ := NewStorage(tempDir)

	_, err := storage.GetFile("nonexistent/file.jpg")
	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}
}

func TestGetFilePathTraversal(t *testing.T) {
	tempDir := t.TempDir()
	storage, _ := NewStorage(tempDir)

	// Try to access file outside base path
	_, err := storage.GetFile("../../etc/passwd")
	if err == nil {
		t.Error("Expected error for path traversal attempt, got nil")
	}
}

func TestDeleteFile(t *testing.T) {
	tempDir := t.TempDir()
	storage, _ := NewStorage(tempDir)

	sessionID := "test-session-123"
	mediaType := MediaTypeImage
	data := []byte("test image data")
	filename := "test.jpg"

	// Save file first
	relativePath, _ := storage.SaveFile(sessionID, mediaType, data, filename)

	// Delete file
	err := storage.DeleteFile(relativePath)
	if err != nil {
		t.Fatalf("Failed to delete file: %v", err)
	}

	// Verify file was deleted
	fullPath := filepath.Join(tempDir, relativePath)
	if _, err := os.Stat(fullPath); !os.IsNotExist(err) {
		t.Error("File was not deleted")
	}
}

func TestDeleteSessionFiles(t *testing.T) {
	tempDir := t.TempDir()
	storage, _ := NewStorage(tempDir)

	sessionID := "test-session-123"
	
	// Save multiple files
	storage.SaveFile(sessionID, MediaTypeImage, []byte("image1"), "test1.jpg")
	storage.SaveFile(sessionID, MediaTypeImage, []byte("image2"), "test2.jpg")
	storage.SaveFile(sessionID, MediaTypeVideo, []byte("video1"), "test1.mp4")

	// Delete all session files
	err := storage.DeleteSessionFiles(sessionID)
	if err != nil {
		t.Fatalf("Failed to delete session files: %v", err)
	}

	// Verify session directory was deleted
	sessionDir := filepath.Join(tempDir, sessionID)
	if _, err := os.Stat(sessionDir); !os.IsNotExist(err) {
		t.Error("Session directory was not deleted")
	}
}

func TestGetFileInfo(t *testing.T) {
	tempDir := t.TempDir()
	storage, _ := NewStorage(tempDir)

	sessionID := "test-session-123"
	mediaType := MediaTypeImage
	data := []byte("test image data")
	filename := "test.jpg"

	// Save file first
	relativePath, _ := storage.SaveFile(sessionID, mediaType, data, filename)

	// Get file info
	info, err := storage.GetFileInfo(relativePath)
	if err != nil {
		t.Fatalf("Failed to get file info: %v", err)
	}

	if info.Path != relativePath {
		t.Errorf("Expected path %s, got %s", relativePath, info.Path)
	}

	if info.Size != int64(len(data)) {
		t.Errorf("Expected size %d, got %d", len(data), info.Size)
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"normal.jpg", "normal.jpg"},
		{"file/with/slashes.jpg", "file_with_slashes.jpg"},
		{"file\\with\\backslashes.jpg", "file_with_backslashes.jpg"},
		{"file..with..dots.jpg", "file__with__dots.jpg"},
	}

	for _, tt := range tests {
		result := sanitizeFilename(tt.input)
		if result != tt.expected {
			t.Errorf("sanitizeFilename(%s) = %s, expected %s", tt.input, result, tt.expected)
		}
	}
}

func TestCalculateChecksum(t *testing.T) {
	tempDir := t.TempDir()
	storage, _ := NewStorage(tempDir)

	data := []byte("test data")
	checksum1 := storage.CalculateChecksum(data)
	checksum2 := storage.CalculateChecksum(data)

	// Same data should produce same checksum
	if checksum1 != checksum2 {
		t.Error("Same data produced different checksums")
	}

	// Different data should produce different checksum
	differentData := []byte("different data")
	checksum3 := storage.CalculateChecksum(differentData)
	if checksum1 == checksum3 {
		t.Error("Different data produced same checksum")
	}
}
