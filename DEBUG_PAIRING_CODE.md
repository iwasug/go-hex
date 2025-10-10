# Debug Pairing Code Connection Issue

## Steps to Debug

### 1. Check Application Logs

Jalankan server dan perhatikan logs saat generate pairing code:

```bash
.\bin\server.exe
```

**Look for:**
- `"requesting pairing code"` - Request started
- `"failed to request pairing code"` - Error occurred
- `"pairing code generated successfully"` - Success

### 2. Test Pairing Code Generation

```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
     "http://localhost:8080/api/v1/whatsapp/sessions/{SESSION_ID}/pairing-code?phone_number=628123456789" \
     -v
```

**Check response:**
- Status code 200?
- `pairing_code` field present?
- `pairing_code` not empty?

### 3. Common Issues

#### Issue 1: Phone Number Format

**Wrong:**
```
+628123456789  ← Has + sign
62 812 3456 789  ← Has spaces
0812 3456 789  ← Missing country code
```

**Correct:**
```
628123456789  ← Just digits with country code
```

#### Issue 2: Session Already Connected

```bash
# Check session status
curl -H "Authorization: Bearer TOKEN" \
     "http://localhost:8080/api/v1/whatsapp/sessions/{SESSION_ID}/status"

# If connected, disconnect first
curl -X POST -H "Authorization: Bearer TOKEN" \
     "http://localhost:8080/api/v1/whatsapp/sessions/{SESSION_ID}/disconnect"
```

#### Issue 3: WhatsApp Connection Failed

Check logs for:
```
ERROR: failed to connect for pairing code
```

**Possible causes:**
- Internet connection issue
- WhatsApp servers down
- Firewall blocking

#### Issue 4: Device Initialization Failed

Check logs for:
```
ERROR: failed to create WhatsApp client
ERROR: device not found
```

**Solution:** Delete and recreate session

### 4. Test with QR Code First

Jika pairing code tidak work, test dengan QR code dulu:

```bash
# Generate QR code
curl -H "Authorization: Bearer TOKEN" \
     "http://localhost:8080/api/v1/whatsapp/sessions/{SESSION_ID}/qr?format=json"
```

If QR code works but pairing code doesn't, issue is specific to pairing implementation.

### 5. Check Phone Number in WhatsApp

**Important:** Phone number yang digunakan harus sama dengan nomor WhatsApp yang akan di-link!

Example:
- API call: `phone_number=628123456789`
- WhatsApp app: Must use same number 0812-3456-789

### 6. Verify API Response

```bash
curl -H "Authorization: Bearer TOKEN" \
     "http://localhost:8080/api/v1/whatsapp/sessions/{SESSION_ID}/pairing-code?phone_number=628123456789" | jq .
```

**Expected:**
```json
{
  "success": true,
  "data": {
    "pairing_code": "12345678",  ← Should be 8 digits
    "phone_number": "628123456789",
    "expires_at": "...",
    "status": "disconnected"
  }
}
```

**If error:**
```json
{
  "success": false,
  "error": {
    "code": "...",
    "message": "..."  ← Check this message
  }
}
```

## Common Error Messages

### "Invalid phone number format"

**Problem:** Phone number format salah

**Solution:**
```bash
# Wrong
curl "...?phone_number=+628123456789"

# Correct
curl "...?phone_number=628123456789"
```

### "Session is already connected"

**Problem:** Session sudah connected

**Solution:**
```bash
# Disconnect first
curl -X POST -H "Authorization: Bearer TOKEN" \
     "http://localhost:8080/api/v1/whatsapp/sessions/{SESSION_ID}/disconnect"

# Then try again
```

### "Failed to request pairing code"

**Check logs for details:**
```
ERROR: failed to request pairing code, error: ...
```

**Common causes:**
1. WhatsApp API changed
2. Network issue
3. Invalid phone number format (internal)
4. Device not initialized properly

### "Pairing code is empty"

**Problem:** WhatsApp API returned empty code

**Possible causes:**
1. Phone number not registered on WhatsApp
2. Phone number format incorrect
3. WhatsApp servers issue

## Testing Checklist

- [ ] Server running without errors
- [ ] Session created successfully
- [ ] Session status = "disconnected"
- [ ] Phone number format correct (no +, no spaces)
- [ ] Phone number has WhatsApp account
- [ ] Bearer token valid
- [ ] Internet connection working
- [ ] No firewall blocking
- [ ] Application logs show no errors

## Report Issue

If still not working, provide:

1. **Error message** from API response
2. **Application logs** (especially lines with "pairing" or "error")
3. **Phone number format** used (hide actual number)
4. **Session status** before requesting code
5. **cURL command** used

Example report:
```
Error: "Failed to request pairing code"

Logs:
ERROR: failed to request pairing code, error: some error details

Phone format: 628123456789 (Indonesia)
Session status: disconnected
Command: curl -H "Authorization: Bearer xxx" "http://localhost:8080/.../pairing-code?phone_number=628123456789"
```
