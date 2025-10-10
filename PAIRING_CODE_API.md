# WhatsApp Pairing Code API

## Overview

Pairing Code adalah alternatif dari QR Code untuk menghubungkan WhatsApp. User memasukkan **8-digit code** secara manual di WhatsApp mobile app instead of scanning QR code.

Pairing Code is an alternative to QR Code for linking WhatsApp. Users enter an **8-digit code** manually in the WhatsApp mobile app instead of scanning a QR code.

---

## Endpoint

```
GET /api/v1/whatsapp/sessions/{session_id}/pairing-code
```

## Authentication

- Requires: Bearer token
- Permission: `whatsapp:read` or session ownership

## Query Parameters

| Parameter     | Type   | Required | Default | Description                                    |
|---------------|--------|----------|---------|------------------------------------------------|
| `phone_number`| string | **Yes**  | -       | Phone number with country code (e.g., 628123456789) |
| `timeout`     | int    | No       | `60`    | Code timeout in seconds (1-300)                |

---

## Phone Number Format

**Format:** Country code + number (NO + or spaces)

### Examples:

| Country    | Phone Number    | Correct Format   |
|------------|-----------------|------------------|
| Indonesia  | +62 812 3456 789| `628123456789`   |
| USA        | +1 555 123 4567 | `15551234567`    |
| UK         | +44 20 1234 5678| `442012345678`   |
| India      | +91 98765 43210 | `919876543210`   |

### ❌ Invalid Formats:
```
+628123456789  ← NO plus sign
62 812 3456 789  ← NO spaces
62-812-3456-789  ← NO dashes
```

### ✅ Valid Format:
```
628123456789  ← Just digits with country code
```

---

## Response Format

### Success Response (200 OK)

```json
{
  "success": true,
  "message": "Pairing code generated successfully",
  "data": {
    "session_id": "e950f799-344e-4372-a1de-386cbd38bad8",
    "pairing_code": "12345678",
    "phone_number": "628123456789",
    "expires_at": "2025-10-10T10:31:00Z",
    "status": "disconnected"
  }
}
```

### Response Fields

| Field          | Type     | Description                                    |
|----------------|----------|------------------------------------------------|
| `session_id`   | string   | Session identifier                             |
| `pairing_code` | string   | 8-digit pairing code                           |
| `phone_number` | string   | Phone number used for pairing                  |
| `expires_at`   | datetime | Expiry time (default 60 seconds from now)      |
| `status`       | string   | Current session status (disconnected/connected)|

---

## Usage Examples

### cURL Example

```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
     "http://localhost:8080/api/v1/whatsapp/sessions/{session_id}/pairing-code?phone_number=628123456789"
```

**Response:**
```json
{
  "success": true,
  "message": "Pairing code generated successfully",
  "data": {
    "session_id": "abc-123",
    "pairing_code": "87654321",
    "phone_number": "628123456789",
    "expires_at": "2025-10-10T10:31:00Z",
    "status": "disconnected"
  }
}
```

### PowerShell Example

```powershell
$token = "YOUR_TOKEN"
$sessionId = "YOUR_SESSION_ID"
$phoneNumber = "628123456789"

$headers = @{
    "Authorization" = "Bearer $token"
}

$url = "http://localhost:8080/api/v1/whatsapp/sessions/$sessionId/pairing-code?phone_number=$phoneNumber"

$response = Invoke-RestMethod -Uri $url -Headers $headers -Method GET

Write-Host "Pairing Code: $($response.data.pairing_code)"
```

### JavaScript (Fetch) Example

```javascript
const token = 'YOUR_TOKEN';
const sessionId = 'YOUR_SESSION_ID';
const phoneNumber = '628123456789';

const response = await fetch(
  `http://localhost:8080/api/v1/whatsapp/sessions/${sessionId}/pairing-code?phone_number=${phoneNumber}`,
  {
    headers: {
      'Authorization': `Bearer ${token}`
    }
  }
);

const data = await response.json();
console.log('Pairing Code:', data.data.pairing_code);
```

### React Example

```jsx
import React, { useState } from 'react';

function PairingCodeGenerator({ sessionId, token }) {
  const [phoneNumber, setPhoneNumber] = useState('');
  const [pairingCode, setPairingCode] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  const generateCode = async () => {
    setLoading(true);
    setError(null);

    try {
      const response = await fetch(
        `/api/v1/whatsapp/sessions/${sessionId}/pairing-code?phone_number=${phoneNumber}`,
        {
          headers: {
            'Authorization': `Bearer ${token}`
          }
        }
      );

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.error.message);
      }

      setPairingCode(data.data.pairing_code);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="pairing-code-generator">
      <h3>Link WhatsApp with Pairing Code</h3>
      
      <div>
        <input
          type="tel"
          placeholder="Phone number (e.g., 628123456789)"
          value={phoneNumber}
          onChange={(e) => setPhoneNumber(e.target.value)}
        />
        <button onClick={generateCode} disabled={loading || !phoneNumber}>
          {loading ? 'Generating...' : 'Generate Code'}
        </button>
      </div>

      {pairingCode && (
        <div className="code-display">
          <h4>Enter this code in WhatsApp:</h4>
          <div className="pairing-code">{pairingCode}</div>
          <p>Code expires in 60 seconds</p>
        </div>
      )}

      {error && <div className="error">{error}</div>}

      <div className="instructions">
        <h4>How to link:</h4>
        <ol>
          <li>Open WhatsApp on your phone</li>
          <li>Tap <strong>Menu</strong> → <strong>Linked Devices</strong></li>
          <li>Tap <strong>Link a Device</strong></li>
          <li>Tap <strong>Link with phone number instead</strong></li>
          <li>Enter the 8-digit code above</li>
        </ol>
      </div>
    </div>
  );
}

export default PairingCodeGenerator;
```

---

## How to Use (Step-by-Step)

### 1. Generate Pairing Code via API

```bash
curl -H "Authorization: Bearer TOKEN" \
     "http://localhost:8080/api/v1/whatsapp/sessions/{id}/pairing-code?phone_number=628123456789"
```

### 2. Copy the 8-Digit Code

```json
{
  "data": {
    "pairing_code": "87654321"  ← Copy this
  }
}
```

### 3. Open WhatsApp on Mobile

1. Open WhatsApp app
2. Tap **Menu** (⋮) → **Linked Devices**
3. Tap **Link a Device**
4. Tap **Link with phone number instead**

### 4. Enter the Code

- Enter the 8-digit code: `87654321`
- Tap **Next**

### 5. Done!

WhatsApp akan otomatis connect dan session status berubah jadi **"connected"**.

---

## Error Responses

### 400 - Phone Number Required

```json
{
  "success": false,
  "error": {
    "code": "WHATSAPP_INVALID_PHONE_NUMBER",
    "message": "Invalid phone number format (E.164 required)",
    "details": {
      "phone_number": ""
    }
  }
}
```

**Solution:** Provide phone_number parameter

### 400 - Invalid Phone Format

```json
{
  "success": false,
  "error": {
    "code": "WHATSAPP_INVALID_PHONE_NUMBER",
    "message": "Invalid phone number format (E.164 required)",
    "details": {
      "phone_number": "+62812..."
    }
  }
}
```

**Solution:** Remove `+` sign, spaces, and dashes

### 409 - Session Already Connected

```json
{
  "success": false,
  "error": {
    "code": "WHATSAPP_SESSION_NOT_CONNECTED",
    "message": "Session is already connected"
  }
}
```

**Solution:** Disconnect session first

```bash
curl -X POST -H "Authorization: Bearer TOKEN" \
     "http://localhost:8080/api/v1/whatsapp/sessions/{id}/disconnect"
```

### 500 - Code Generation Failed

```json
{
  "success": false,
  "error": {
    "code": "WHATSAPP_QR_CODE_GENERATION_FAILED",
    "message": "Failed to generate QR code"
  }
}
```

**Solutions:**
- Check internet connection
- Check WhatsApp servers status
- Try again after a few seconds
- Check application logs

---

## QR Code vs Pairing Code

| Feature            | QR Code                      | Pairing Code              |
|--------------------|------------------------------|---------------------------|
| **Method**         | Scan with camera             | Enter code manually       |
| **Code Type**      | Visual QR image              | 8-digit number            |
| **Expiry**         | 60 seconds (default)         | 60 seconds (default)      |
| **Use Case**       | Quick setup, desktop         | No camera / troubleshooting|
| **Accessibility**  | Requires camera              | Keyboard only             |
| **Network**        | Device must see screen       | Just enter digits         |

### When to Use Pairing Code?

✅ **Good for:**
- Camera tidak berfungsi / Camera not working
- Screen sharing / remote access scenarios
- Accessibility needs (visually impaired users)
- Mobile-to-mobile linking
- Testing/automation

✅ **Good for QR Code:**
- Quick setup
- Desktop applications
- Normal consumer use cases
- Camera available

---

## Security Notes

1. **Pairing codes expire after 60 seconds** (default)
2. **One-time use** - Code cannot be reused
3. **Phone number required** - Links to specific phone number
4. **Secure transmission** - Always use HTTPS in production
5. **Rate limiting** - Implement rate limits to prevent abuse

---

## Troubleshooting

### Problem: "Code doesn't work in WhatsApp"

**Checks:**
1. Code expired? (max 60 seconds)
2. Correct phone number format?
3. Correct digit sequence?
4. Session already connected?

**Solution:**
```bash
# Generate fresh code
curl -H "Authorization: Bearer TOKEN" \
     "http://localhost:8080/api/v1/whatsapp/sessions/{id}/pairing-code?phone_number=628123456789"
```

### Problem: "Invalid phone number"

**Check format:**
```
❌ Wrong: +628123456789
❌ Wrong: 62 812 3456 789
❌ Wrong: 0812 3456 789

✅ Correct: 628123456789
```

### Problem: "Session already connected"

**Disconnect first:**
```bash
curl -X POST -H "Authorization: Bearer TOKEN" \
     "http://localhost:8080/api/v1/whatsapp/sessions/{id}/disconnect"
```

### Problem: "Code generation timeout"

**Increase timeout:**
```bash
curl -H "Authorization: Bearer TOKEN" \
     "http://localhost:8080/api/v1/whatsapp/sessions/{id}/pairing-code?phone_number=628123456789&timeout=120"
```

---

## Testing

### Full Flow Test

```bash
# 1. Login
TOKEN=$(curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"password123"}' \
  | jq -r '.data.access_token')

# 2. Create session
SESSION=$(curl -X POST http://localhost:8080/api/v1/whatsapp/sessions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"session_name":"Test"}' \
  | jq -r '.data.id')

# 3. Generate pairing code
curl -H "Authorization: Bearer $TOKEN" \
     "http://localhost:8080/api/v1/whatsapp/sessions/$SESSION/pairing-code?phone_number=628123456789" \
     | jq .

# 4. Enter code in WhatsApp mobile app

# 5. Check session status
curl -H "Authorization: Bearer $TOKEN" \
     "http://localhost:8080/api/v1/whatsapp/sessions/$SESSION/status" \
     | jq .
```

---

## API Comparison

### Pairing Code Endpoint

```
GET /api/v1/whatsapp/sessions/{id}/pairing-code?phone_number=628123456789
```

**Returns:** 8-digit code string

### QR Code Endpoint

```
GET /api/v1/whatsapp/sessions/{id}/qr
```

**Returns:** QR code (image or JSON)

---

## Best Practices

1. **Always validate phone number** before sending to API
2. **Display countdown timer** (60 seconds)
3. **Auto-refresh** if code expires
4. **Clear error messages** for users
5. **Provide instructions** for entering code
6. **Implement retry logic** for failed requests
7. **Log successful pairings** for audit trail

---

## Summary

✅ **Pairing Code API provides:**
- Alternative to QR code scanning
- Phone number-based linking
- 8-digit code generation
- 60-second expiry (configurable)
- REST API integration
- Same security as QR code

✅ **Perfect for:**
- Mobile-only users
- Accessibility needs
- Camera-less scenarios
- Remote linking
- Automated testing

✅ **Easy integration:**
- Single API call
- Simple response format
- Clear error messages
- Comprehensive documentation
