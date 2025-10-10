# Task 19: Media File Handling - Quick Reference

## Package Location
`internal/pkg/media/`

## Quick Start

### Initialize Media Manager

```go
import "golang-chi-hexagonal-base/internal/pkg/media"

// Create manager with base path and max size
manager, err := media.NewManager("./data/whatsapp/media", 16)
if err != nil {
    log.Fatal(err)
}
```

### Save Media File

```go
// Save from byte array
relativePath, err := manager.SaveMedia(
    "session-123",           // Session ID
    imageData,               // File data
    "image/jpeg",            // MIME type
    "photo.jpg",             // Original filename
)

// Save from io.Reader
relativePath, err := manager.SaveMediaFromReader(
    "session-123",
    fileReader,
    "video/mp4",
    "video.mp4",
)
```

### Retrieve Media File

```go
data, err := manager.GetMedia(relativePath)
```

### Delete Media

```go
// Delete single file
err := manager.DeleteMedia(relativePath)

// Delete all session media
err := manager.DeleteSessionMedia("session-123")
```

### Validate Without Saving

```go
err := manager.ValidateMediaFile(data, "image/jpeg", "photo.jpg")
```

## Supported Media Types

### Images (7 types)
- `image/jpeg`, `image/jpg`, `image/png`, `image/gif`
- `image/webp`, `image/bmp`, `image/tiff`

### Videos (7 types)
- `video/mp4`, `video/mpeg`, `video/3gpp`, `video/quicktime`
- `video/x-matroska`, `video/webm`, `video/x-msvideo`

### Audio (9 types)
- `audio/aac`, `audio/mp4`, `audio/mpeg`, `audio/mp3`
- `audio/amr`, `audio/ogg`, `audio/opus`, `audio/wav`, `audio/webm`

### Documents (17 types)
- PDF: `application/pdf`
- MS Office: `application/msword`, `application/vnd.ms-excel`, `application/vnd.ms-powerpoint`
- OpenXML: `application/vnd.openxmlformats-officedocument.*`
- Text: `text/plain`, `text/csv`
- Archives: `application/zip`, `application/x-zip-compressed`, `application/x-rar-compressed`, `application/x-7z-compressed`
- Data: `application/json`, `application/xml`, `text/xml`

## Validation Rules

### File Size
- Default: 16MB
- Configurable via `WHATSAPP_MEDIA_MAX_SIZE_MB`
- Validated before saving

### MIME Type
- Whitelist-based validation
- Only allowed types can be saved
- Automatic media type determination

### Filename
- Max length: 255 characters
- No path separators (`/`, `\`)
- No path traversal (`..`)
- Automatic sanitization

## Directory Structure

```
basePath/
└── {sessionID}/
    ├── image/
    │   └── {timestamp}_{uuid}_{filename}
    ├── video/
    │   └── {timestamp}_{uuid}_{filename}
    ├── audio/
    │   └── {timestamp}_{uuid}_{filename}
    └── document/
        └── {timestamp}_{uuid}_{filename}
```

## Configuration

### Environment Variables

```env
WHATSAPP_MEDIA_MAX_SIZE_MB=16
WHATSAPP_MEDIA_STORAGE_PATH=./data/whatsapp/media
```

### Config Struct

```go
type WhatsAppConfig struct {
    MediaMaxSizeMB   int    // Maximum file size in MB
    MediaStoragePath string // Base path for media storage
}
```

## Integration with WhatsApp Service

### Service Interface

```go
type MediaManager interface {
    DeleteSessionMedia(sessionID string) error
}
```

### Adapter

```go
import mediaadapter "golang-chi-hexagonal-base/internal/adapter/media"

mediaAdapter := mediaadapter.NewManagerAdapter(mediaManager)
```

### Service Initialization

```go
whatsappService := service.NewWhatsAppService(
    sessionRepo,
    messageRepo,
    userRepo,
    clientManager,
    webhookClient,
    mediaAdapter, // Media manager
)
```

## Common Operations

### Determine Media Type from MIME

```go
import "golang-chi-hexagonal-base/internal/pkg/media"

mediaType, err := media.DetermineMediaType("image/jpeg")
// Returns: media.MediaTypeImage
```

### Check if MIME Type is Allowed

```go
isAllowed := media.IsAllowedMimeType("image/jpeg")
// Returns: true
```

### Detect MIME Type from Filename

```go
mimeType := media.DetectMimeType("photo.jpg")
// Returns: "image/jpeg"
```

### Get File Extension from MIME Type

```go
ext := media.GetFileExtension("image/jpeg")
// Returns: ".jpg"
```

## Error Handling

### Common Errors

```go
// File not found
_, err := manager.GetMedia("nonexistent/file.jpg")
// Error: "file not found: nonexistent/file.jpg"

// File too large
err := manager.SaveMedia(sessionID, largeData, "image/jpeg", "large.jpg")
// Error: "validation failed: file size (20971520 bytes) exceeds maximum allowed size of 16MB"

// Invalid MIME type
err := manager.SaveMedia(sessionID, data, "text/html", "page.html")
// Error: "validation failed: MIME type text/html is not allowed"

// Path traversal attempt
_, err := manager.GetMedia("../../etc/passwd")
// Error: "access denied: path outside storage directory"
```

## Security Features

1. **Path Traversal Protection** - All paths validated
2. **Filename Sanitization** - Dangerous characters removed
3. **MIME Type Validation** - Whitelist-based
4. **File Size Limits** - Prevents disk exhaustion
5. **Unique Naming** - UUID-based, prevents conflicts

## Testing

```bash
# Run media package tests
go test ./internal/pkg/media/... -v

# Run all tests
go test ./... -v
```

## Performance Notes

- Files loaded into memory (suitable for 16MB limit)
- Synchronous I/O operations
- Efficient directory structure for cleanup
- SHA256 checksums available for integrity

## Helper Functions

### Calculate Checksum

```go
checksum := manager.storage.CalculateChecksum(data)
// Returns: SHA256 hex string
```

### Get File Info

```go
info, err := manager.GetMediaInfo(relativePath)
// Returns: FileInfo{Path, Size, ModifiedTime}
```

### Get Size Limits

```go
maxMB := manager.GetMaxSizeMB()      // Returns: 16
maxBytes := manager.GetMaxSizeBytes() // Returns: 16777216
```

## Cleanup on Session Deletion

Media files are automatically cleaned up when a session is deleted:

```go
// In WhatsAppService.DeleteSession()
if s.mediaManager != nil {
    err = s.mediaManager.DeleteSessionMedia(sessionID)
    // Errors logged but don't prevent session deletion
}
```

## Best Practices

1. **Always validate** before saving
2. **Use temp directories** for tests
3. **Handle errors gracefully** - media cleanup failures shouldn't block operations
4. **Check file sizes** before loading into memory
5. **Use unique filenames** - provided automatically by the manager

## Related Files

- `internal/pkg/media/storage.go` - Storage implementation
- `internal/pkg/media/validator.go` - Validation logic
- `internal/pkg/media/manager.go` - High-level interface
- `internal/adapter/media/manager_adapter.go` - Service adapter
- `internal/pkg/media/README.md` - Detailed documentation
