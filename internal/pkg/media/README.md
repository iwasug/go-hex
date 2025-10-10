# Media File Handling Package

This package provides comprehensive media file handling for the WhatsApp integration, including storage, validation, and management of media files (images, videos, audio, documents).

## Features

- **File Storage**: Organized directory structure for media files by session and type
- **Validation**: File size, MIME type, and filename validation
- **Security**: Path traversal protection and sanitization
- **Unique Naming**: UUID-based unique filenames to prevent conflicts
- **Session Cleanup**: Automatic cleanup of all media files when a session is deleted

## Components

### Storage (`storage.go`)

Handles the physical storage of media files on disk.

**Key Features:**
- Creates organized directory structure: `basePath/sessionID/mediaType/filename`
- Generates unique filenames using UUID and timestamp
- Provides path traversal protection
- Supports file checksum calculation

**Example Usage:**
```go
storage, err := media.NewStorage("./data/whatsapp/media")
if err != nil {
    log.Fatal(err)
}

// Save a file
relativePath, err := storage.SaveFile(
    "session-123",
    media.MediaTypeImage,
    imageData,
    "photo.jpg",
)

// Retrieve a file
data, err := storage.GetFile(relativePath)

// Delete session files
err = storage.DeleteSessionFiles("session-123")
```

### Validator (`validator.go`)

Validates media files before storage.

**Validation Rules:**
- **File Size**: Configurable maximum size (default: 16MB)
- **MIME Types**: Only allows specific media types
- **Filenames**: Prevents path traversal and validates length

**Supported MIME Types:**

**Images:**
- image/jpeg, image/jpg, image/png, image/gif, image/webp, image/bmp, image/tiff

**Videos:**
- video/mp4, video/mpeg, video/3gpp, video/quicktime, video/x-matroska, video/webm

**Audio:**
- audio/aac, audio/mp4, audio/mpeg, audio/mp3, audio/amr, audio/ogg, audio/opus, audio/wav

**Documents:**
- application/pdf, application/msword, application/vnd.ms-excel
- application/vnd.openxmlformats-officedocument.* (Office formats)
- text/plain, text/csv
- application/zip, application/json, application/xml

**Example Usage:**
```go
validator := media.NewValidator(media.ValidationConfig{
    MaxSizeMB: 16,
})

// Validate a file
err := validator.ValidateFile(fileData, "image/jpeg", "photo.jpg")

// Determine media type from MIME type
mediaType, err := media.DetermineMediaType("image/jpeg")
```

### Manager (`manager.go`)

High-level interface combining storage and validation.

**Example Usage:**
```go
manager, err := media.NewManager("./data/whatsapp/media", 16)
if err != nil {
    log.Fatal(err)
}

// Save media with validation
relativePath, err := manager.SaveMedia(
    "session-123",
    imageData,
    "image/jpeg",
    "photo.jpg",
)

// Get media
data, err := manager.GetMedia(relativePath)

// Delete session media
err = manager.DeleteSessionMedia("session-123")
```

## Directory Structure

Media files are organized as follows:

```
basePath/
├── session-id-1/
│   ├── image/
│   │   ├── 1234567890_abc12345_photo.jpg
│   │   └── 1234567891_def67890_screenshot.png
│   ├── video/
│   │   └── 1234567892_ghi12345_video.mp4
│   ├── audio/
│   │   └── 1234567893_jkl67890_voice.ogg
│   └── document/
│       └── 1234567894_mno12345_report.pdf
└── session-id-2/
    └── image/
        └── 1234567895_pqr12345_image.jpg
```

## Integration with WhatsApp Service

The media manager is integrated with the WhatsApp service through an adapter:

```go
// Create media manager
mediaManager, err := media.NewManager(config.WhatsApp.MediaStoragePath, config.WhatsApp.MediaMaxSizeMB)

// Create adapter
mediaAdapter := mediaadapter.NewManagerAdapter(mediaManager)

// Pass to WhatsApp service
whatsappService := service.NewWhatsAppService(
    sessionRepo,
    messageRepo,
    userRepo,
    clientManager,
    webhookClient,
    mediaAdapter, // Media manager
)
```

## Configuration

Media handling is configured through environment variables:

```env
# Maximum media file size in MB
WHATSAPP_MEDIA_MAX_SIZE_MB=16

# Media storage path
WHATSAPP_MEDIA_STORAGE_PATH=./data/whatsapp/media
```

## Security Considerations

1. **Path Traversal Protection**: All file paths are validated to prevent access outside the base directory
2. **Filename Sanitization**: Dangerous characters are removed from filenames
3. **MIME Type Validation**: Only allowed media types can be stored
4. **File Size Limits**: Configurable maximum file size prevents disk exhaustion
5. **Unique Naming**: UUID-based naming prevents filename conflicts and overwrites

## Testing

The package includes comprehensive tests:

```bash
go test ./internal/pkg/media/... -v
```

Tests cover:
- Storage operations (save, retrieve, delete)
- Validation (size, MIME type, filename)
- Security (path traversal, sanitization)
- Manager operations (integration tests)

## Error Handling

All operations return descriptive errors:

```go
// File not found
_, err := storage.GetFile("nonexistent/file.jpg")
// Error: "file not found: nonexistent/file.jpg"

// File too large
err := validator.ValidateSize(20 * 1024 * 1024) // 20MB
// Error: "file size (20971520 bytes) exceeds maximum allowed size of 16MB"

// Invalid MIME type
err := validator.ValidateMimeType("text/html")
// Error: "MIME type text/html is not allowed"

// Path traversal attempt
_, err := storage.GetFile("../../etc/passwd")
// Error: "access denied: path outside storage directory"
```

## Performance Considerations

1. **Disk I/O**: File operations are synchronous; consider async operations for high-volume scenarios
2. **Memory Usage**: Large files are read into memory; consider streaming for very large files
3. **Directory Structure**: Organized by session and type for efficient cleanup and management
4. **Checksums**: SHA256 checksums available for file integrity verification

## Future Enhancements

Potential improvements:
- Cloud storage support (S3, Azure Blob, GCS)
- Image thumbnail generation
- Video transcoding
- Compression for large files
- Streaming support for large files
- Metadata extraction (EXIF, duration, etc.)
