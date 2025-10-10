# Debug QR Code Generation Issue

## Symptoms
QR code image tidak ter-generate (gambar kosong atau error)

## Debugging Steps

### 1. Check Application Logs

Setelah request QR code, check log aplikasi untuk melihat:

```bash
# Windows PowerShell
.\bin\server.exe 2>&1 | Select-String -Pattern "QR|whatsapp|error"
```

**Look for these log messages:**
- ✅ `"disconnecting existing connection for fresh QR code"` - OK
- ✅ `"waiting for QR code event"` - OK
- ✅ `"received QR event"` - OK
- ✅ `"QR code generated successfully"` - OK

**Error messages to watch:**
- ❌ `"QR code is empty"` - Problem with whatsmeow
- ❌ `"QR code generation timeout"` - WhatsApp connection issue
- ❌ `"failed to get QR channel"` - Client initialization problem
- ❌ `"failed to connect for QR code"` - Network/connection issue

### 2. Test with JSON Format First

Coba dulu dengan format JSON untuk memastikan QR code string berhasil di-generate:

```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
     "http://localhost:8080/api/v1/whatsapp/sessions/{session_id}/qr?format=json"
```

**Expected Response:**
```json
{
  "success": true,
  "message": "QR code generated successfully",
  "data": {
    "session_id": "...",
    "qr_code": "2@ABC123XYZ...",  // <-- Should have value!
    "expires_at": "...",
    "status": "disconnected"
  }
}
```

**If `qr_code` is empty or missing:**
- Problem is in WhatsApp client connection, not image generation

**If `qr_code` has value:**
- WhatsApp client OK, problem might be in qrcode.Encode()

### 3. Test Image Generation

Jika JSON format berhasil, test image format:

```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
     "http://localhost:8080/api/v1/whatsapp/sessions/{session_id}/qr?format=image&size=256" \
     -o test-qr.png
```

Check file size:
```bash
# Windows
(Get-Item test-qr.png).Length

# Should be > 0 bytes
```

### 4. Common Issues & Solutions

#### Issue 1: Session Already Connected
**Error:** `"Session is already connected"`

**Solution:**
```bash
# Disconnect session first
curl -X POST -H "Authorization: Bearer YOUR_TOKEN" \
     "http://localhost:8080/api/v1/whatsapp/sessions/{session_id}/disconnect"

# Then try QR code again
```

#### Issue 2: Device Not Found
**Error:** `"device not found"` or `"client not found"`

**Check Database:**
```sql
-- Check if device exists
SELECT id, session_id, registration_id 
FROM whatsapp_devices 
WHERE session_id = 'YOUR_SESSION_ID';

-- Check if session exists
SELECT id, name, status 
FROM whatsapp_sessions 
WHERE id = 'YOUR_SESSION_ID';
```

**Solution:**
- Device akan auto-create saat pertama kali connect
- Pastikan session exists

#### Issue 3: Timeout
**Error:** `"QR code generation timeout after 60s"`

**Possible Causes:**
- Database connection slow
- WhatsApp server unreachable
- Firewall blocking

**Solution:**
```bash
# Try with longer timeout
curl -H "Authorization: Bearer YOUR_TOKEN" \
     "http://localhost:8080/api/v1/whatsapp/sessions/{session_id}/qr?timeout=120&format=json"
```

#### Issue 4: Empty QR Code String
**Error:** `"QR code is empty"`

**Check Logs for:**
- `"received QR event"` with `code_length: 0`
- WhatsApp connection errors
- Store/device initialization errors

**Solution:**
```bash
# Recreate session
curl -X DELETE -H "Authorization: Bearer YOUR_TOKEN" \
     "http://localhost:8080/api/v1/whatsapp/sessions/{session_id}"

curl -X POST -H "Authorization: Bearer YOUR_TOKEN" \
     -H "Content-Type: application/json" \
     -d '{"name": "Test Session", "description": "Testing"}' \
     "http://localhost:8080/api/v1/whatsapp/sessions"
```

### 5. Check Environment

**Required:**
- ✅ PostgreSQL running
- ✅ Database migrations applied
- ✅ Internet connection (for WhatsApp servers)
- ✅ Port 8080 available

**Check PostgreSQL:**
```bash
# Test connection
psql -h localhost -U your_user -d your_db -c "SELECT 1;"
```

**Check Tables:**
```sql
\dt whatsapp*
-- Should show:
-- whatsapp_sessions
-- whatsapp_devices
-- whatsapp_messages
```

### 6. Full Debug Flow

**Step-by-step test:**

```bash
# 1. Create session
SESSION_RESPONSE=$(curl -X POST -H "Authorization: Bearer $TOKEN" \
     -H "Content-Type: application/json" \
     -d '{"name": "Debug Session", "description": "Testing QR"}' \
     "http://localhost:8080/api/v1/whatsapp/sessions")

echo $SESSION_RESPONSE

# 2. Extract session ID from response
SESSION_ID="..." # Get from response

# 3. Request QR (JSON format for debugging)
curl -v -H "Authorization: Bearer $TOKEN" \
     "http://localhost:8080/api/v1/whatsapp/sessions/$SESSION_ID/qr?format=json"

# 4. Check logs for errors

# 5. If JSON works, try image
curl -H "Authorization: Bearer $TOKEN" \
     "http://localhost:8080/api/v1/whatsapp/sessions/$SESSION_ID/qr?format=image" \
     -o qr-test.png

# 6. Verify image
file qr-test.png  # Should say "PNG image data"
```

### 7. Enhanced Logging

Saya sudah tambahkan enhanced logging di code. Check untuk messages ini:

**In client.go:**
- `"disconnecting existing connection for fresh QR code"`
- `"waiting for QR code event"` with timeout info
- `"received QR event"` with event type and code length
- `"QR code generated successfully"` with QR preview

**In handler:**
- `"QR code is empty"` - QR string validation
- `"Generated QR image is empty"` - Image generation validation
- `"Failed to generate QR code image: [error]"` - qrcode.Encode error

### 8. Expected Full Flow Logs

```
INFO: disconnecting existing connection for fresh QR code
INFO: waiting for QR code event, timeout=1m0s
INFO: received QR event, event=code, code_length=150
INFO: QR code generated successfully, qr_code_preview=2@ABcD1234567890123...
```

## Quick Test Commands

### Test 1: JSON Format
```bash
curl -H "Authorization: Bearer $TOKEN" \
     "http://localhost:8080/api/v1/whatsapp/sessions/{id}/qr?format=json" | jq .
```

### Test 2: Image Format
```bash
curl -H "Authorization: Bearer $TOKEN" \
     "http://localhost:8080/api/v1/whatsapp/sessions/{id}/qr?format=image" \
     -o qr.png && explorer qr.png
```

### Test 3: Check File Size
```powershell
$file = Get-Item qr.png
Write-Host "File size: $($file.Length) bytes"
if ($file.Length -gt 0) { 
    Write-Host "✅ QR image generated successfully"
} else { 
    Write-Host "❌ QR image is empty"
}
```

## Report Issues

Jika masih error, provide:
1. **Error message** dari API response
2. **Application logs** (especially lines with "QR" or "whatsapp")
3. **Database check** hasil query whatsapp_sessions dan whatsapp_devices
4. **JSON format response** (does it have qr_code value?)
5. **File size** dari downloaded PNG

Example report:
```
Error: "QR code is empty"

Logs:
- INFO: disconnecting existing connection for fresh QR code
- INFO: waiting for QR code event, timeout=1m0s
- INFO: received QR event, event=code, code_length=0  <-- PROBLEM
- ERROR: QR code is empty

JSON Response:
{
  "data": {
    "qr_code": "",  <-- EMPTY
    ...
  }
}

Database:
- Session exists: YES
- Device exists: NO  <-- MIGHT BE ISSUE
```
