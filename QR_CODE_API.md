# WhatsApp QR Code API

## Overview
The QR code endpoint can now return either a PNG image (default) or JSON data, making it easy to integrate WhatsApp authentication into web and mobile applications.

## Endpoint
```
GET /api/v1/whatsapp/sessions/{session_id}/qr
```

## Authentication
- Requires: Bearer token
- Permission: `whatsapp:read` or session ownership

## Query Parameters

| Parameter | Type   | Default | Description                                      | Range/Values        |
|-----------|--------|---------|--------------------------------------------------|---------------------|
| `format`  | string | `image` | Response format                                  | `image`, `json`     |
| `size`    | int    | `256`   | QR code image size in pixels (image format only) | 128-1024           |
| `timeout` | int    | `60`    | QR code timeout in seconds                       | 1-300 (5 minutes)  |

## Response Formats

### 1. Image Format (Default) - File Download

**Request:**
```bash
GET /api/v1/whatsapp/sessions/{session_id}/qr
# or explicitly
GET /api/v1/whatsapp/sessions/{session_id}/qr?format=image&size=512
```

**Response:**
- **Content-Type:** `image/png`
- **Body:** Binary PNG image data (automatic download)

**Response Headers:**
```
Content-Type: image/png
Content-Disposition: attachment; filename="whatsapp-qr-{session_id}.png"
Content-Length: 12345
Cache-Control: no-cache, no-store, must-revalidate
X-QR-Expires-At: 2025-10-10T02:31:00Z
X-Session-ID: e950f799-344e-4372-a1de-386cbd38bad8
X-Session-Status: disconnected
```

**Behavior:**
- Browser will **automatically download** the file
- File name: `whatsapp-qr-{session_id}.png`
- Works perfectly in **Swagger UI** - just click "Download file"

**Direct Link (Auto Download):**
```html
<a href="/api/v1/whatsapp/sessions/{session_id}/qr?format=image&size=300"
   download="whatsapp-qr.png">
   Download QR Code
</a>
```

**JavaScript Fetch & Display:**
```javascript
// Fetch and display QR code image
const response = await fetch('/api/v1/whatsapp/sessions/{session_id}/qr', {
    headers: {
        'Authorization': `Bearer ${token}`
    }
});

const blob = await response.blob();
const imageUrl = URL.createObjectURL(blob);

// Display in img tag
document.getElementById('qr-image').src = imageUrl;

// Or trigger download
const a = document.createElement('a');
a.href = imageUrl;
a.download = 'whatsapp-qr.png';
a.click();

// Check expiry from headers
const expiresAt = response.headers.get('X-QR-Expires-At');
console.log('QR Code expires at:', expiresAt);
```

### 2. JSON Format

**Request:**
```bash
GET /api/v1/whatsapp/sessions/{session_id}/qr?format=json
```

**Response:**
```json
{
    "success": true,
    "message": "QR code generated successfully",
    "data": {
        "session_id": "e950f799-344e-4372-a1de-386cbd38bad8",
        "qr_code": "2@ABC123XYZ...",
        "expires_at": "2025-10-10T02:31:00Z",
        "status": "disconnected"
    }
}
```

**Use Case:** When you need to generate the QR code image client-side or use a different QR library.

## Usage Examples

### cURL Examples

**Get QR Code as Image:**
```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
     "http://localhost:8080/api/v1/whatsapp/sessions/{session_id}/qr" \
     -o whatsapp-qr.png
```

**Get QR Code as Image with Custom Size:**
```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
     "http://localhost:8080/api/v1/whatsapp/sessions/{session_id}/qr?format=image&size=512" \
     -o whatsapp-qr-large.png
```

**Get QR Code Data as JSON:**
```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
     "http://localhost:8080/api/v1/whatsapp/sessions/{session_id}/qr?format=json"
```

### React Example

```jsx
import React, { useEffect, useState } from 'react';

function WhatsAppQRCode({ sessionId, token }) {
    const [qrUrl, setQrUrl] = useState('');
    const [expiresAt, setExpiresAt] = useState('');
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);

    useEffect(() => {
        const fetchQRCode = async () => {
            try {
                const response = await fetch(
                    `/api/v1/whatsapp/sessions/${sessionId}/qr?format=image&size=300`,
                    {
                        headers: {
                            'Authorization': `Bearer ${token}`
                        }
                    }
                );

                if (!response.ok) {
                    throw new Error('Failed to fetch QR code');
                }

                const blob = await response.blob();
                const url = URL.createObjectURL(blob);
                setQrUrl(url);
                setExpiresAt(response.headers.get('X-QR-Expires-At'));
                setLoading(false);
            } catch (err) {
                setError(err.message);
                setLoading(false);
            }
        };

        fetchQRCode();

        // Cleanup
        return () => {
            if (qrUrl) {
                URL.revokeObjectURL(qrUrl);
            }
        };
    }, [sessionId, token]);

    if (loading) return <div>Loading QR Code...</div>;
    if (error) return <div>Error: {error}</div>;

    return (
        <div>
            <h3>Scan QR Code with WhatsApp</h3>
            <img src={qrUrl} alt="WhatsApp QR Code" />
            <p>Expires at: {new Date(expiresAt).toLocaleString()}</p>
        </div>
    );
}

export default WhatsAppQRCode;
```

### Vue.js Example

```vue
<template>
  <div class="whatsapp-qr">
    <h3>Scan QR Code with WhatsApp</h3>
    <div v-if="loading">Loading QR Code...</div>
    <div v-else-if="error">Error: {{ error }}</div>
    <div v-else>
      <img :src="qrUrl" alt="WhatsApp QR Code" />
      <p>Expires at: {{ expiryTime }}</p>
    </div>
  </div>
</template>

<script>
export default {
  props: {
    sessionId: String,
    token: String
  },
  data() {
    return {
      qrUrl: '',
      expiresAt: '',
      loading: true,
      error: null
    };
  },
  computed: {
    expiryTime() {
      return this.expiresAt ? new Date(this.expiresAt).toLocaleString() : '';
    }
  },
  async mounted() {
    try {
      const response = await fetch(
        `/api/v1/whatsapp/sessions/${this.sessionId}/qr?format=image&size=300`,
        {
          headers: {
            'Authorization': `Bearer ${this.token}`
          }
        }
      );

      if (!response.ok) {
        throw new Error('Failed to fetch QR code');
      }

      const blob = await response.blob();
      this.qrUrl = URL.createObjectURL(blob);
      this.expiresAt = response.headers.get('X-QR-Expires-At');
      this.loading = false;
    } catch (err) {
      this.error = err.message;
      this.loading = false;
    }
  },
  beforeUnmount() {
    if (this.qrUrl) {
      URL.revokeObjectURL(this.qrUrl);
    }
  }
};
</script>
```

## Error Responses

All error responses follow the standard format:

```json
{
    "success": false,
    "error": {
        "code": "ERROR_CODE",
        "message": "Human readable error message",
        "details": null
    }
}
```

### Common Error Codes

| Status | Code                          | Description                              |
|--------|-------------------------------|------------------------------------------|
| 400    | `VALIDATION_ERROR`            | Invalid parameters                       |
| 401    | `UNAUTHORIZED`                | Missing or invalid token                 |
| 403    | `FORBIDDEN`                   | No permission to access session          |
| 404    | `NOT_FOUND`                   | Session not found                        |
| 409    | `SESSION_ALREADY_CONNECTED`   | Session is already connected to WhatsApp |
| 500    | `INTERNAL_ERROR`              | Server error                             |

## Best Practices

1. **Auto-refresh QR Codes:** QR codes expire after the specified timeout (default 60 seconds). Implement auto-refresh logic.

2. **Handle Expiry:** Use the `X-QR-Expires-At` header to display a countdown timer.

3. **Error Handling:** Always handle errors gracefully and provide clear user feedback.

4. **Image Size:** Use appropriate sizes based on display:
   - Mobile: 256-300px
   - Desktop: 300-512px
   - High-DPI displays: 512px or larger

5. **Caching:** QR codes should not be cached (Cache-Control headers are set appropriately).

6. **Resource Cleanup:** Remember to revoke blob URLs when components unmount to prevent memory leaks.

## Security Notes

- QR codes contain sensitive authentication data
- QR codes expire after the specified timeout
- Always use HTTPS in production
- Validate the Bearer token on every request
- QR codes are session-specific and cannot be reused across sessions
