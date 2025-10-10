# Quick Test - Pairing Code

## Test dengan cURL

### Step 1: Login
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"admin@example.com\",\"password\":\"password123\"}"
```

Copy `access_token` dari response.

### Step 2: Create Session
```bash
curl -X POST http://localhost:8080/api/v1/whatsapp/sessions \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"session_name\":\"Test Pairing\"}"
```

Copy `id` dari response.

### Step 3: Generate Pairing Code
```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
     "http://localhost:8080/api/v1/whatsapp/sessions/YOUR_SESSION_ID/pairing-code?phone_number=628123456789"
```

**Replace:**
- `YOUR_TOKEN` = access token from step 1
- `YOUR_SESSION_ID` = session id from step 2  
- `628123456789` = your phone number (country code + number, NO + sign)

### Expected Response:
```json
{
  "success": true,
  "message": "Pairing code generated successfully",
  "data": {
    "session_id": "...",
    "pairing_code": "12345678",
    "phone_number": "628123456789",
    "expires_at": "2025-10-10T...",
    "status": "disconnected"
  }
}
```

### Step 4: Enter Code in WhatsApp

1. Open WhatsApp on your phone
2. Tap Menu → Linked Devices
3. Tap "Link a Device"
4. Tap "Link with phone number instead"
5. Enter the 8-digit code
6. Done!

## Common Issues

### Error: "Invalid phone number format"

**Problem:** Phone number format salah

**Fix:**
```bash
# Wrong
?phone_number=+628123456789
?phone_number=62 812 3456 789

# Correct
?phone_number=628123456789
```

### Error: "Session is already connected"

**Fix:**
```bash
# Disconnect first
curl -X POST -H "Authorization: Bearer YOUR_TOKEN" \
     "http://localhost:8080/api/v1/whatsapp/sessions/YOUR_SESSION_ID/disconnect"

# Then try again
```

### Error: "Pairing code is empty"

**Possible causes:**
1. Phone number not registered on WhatsApp
2. Phone number format incorrect
3. Network issue

**Check:**
```bash
# Test if phone number format is valid
echo "628123456789" | wc -c  # Should be 12-15 characters
```

### Server Logs Show Error

Check server console untuk error messages:
- `"failed to connect for pairing code"` = Connection issue
- `"failed to request pairing code"` = WhatsApp API error  
- `"pairing code is empty"` = Phone not on WhatsApp or format wrong

## Full PowerShell Example

```powershell
# Variables
$baseUrl = "http://localhost:8080/api/v1"
$email = "admin@example.com"
$password = "password123"
$phoneNumber = "628123456789"

# 1. Login
$loginResponse = Invoke-RestMethod -Uri "$baseUrl/auth/login" `
    -Method POST `
    -ContentType "application/json" `
    -Body (@{email=$email; password=$password} | ConvertTo-Json)

$token = $loginResponse.data.access_token
Write-Host "✓ Logged in"

# 2. Create Session
$sessionResponse = Invoke-RestMethod -Uri "$baseUrl/whatsapp/sessions" `
    -Method POST `
    -Headers @{Authorization="Bearer $token"} `
    -ContentType "application/json" `
    -Body (@{session_name="Test Pairing"} | ConvertTo-Json)

$sessionId = $sessionResponse.data.id
Write-Host "✓ Session created: $sessionId"

# 3. Generate Pairing Code
$pairingResponse = Invoke-RestMethod -Uri "$baseUrl/whatsapp/sessions/$sessionId/pairing-code?phone_number=$phoneNumber" `
    -Headers @{Authorization="Bearer $token"}

$code = $pairingResponse.data.pairing_code
Write-Host ""
Write-Host "========================"
Write-Host "PAIRING CODE: $code"
Write-Host "========================"
Write-Host ""
Write-Host "Enter this code in WhatsApp:"
Write-Host "1. Open WhatsApp"
Write-Host "2. Menu → Linked Devices"
Write-Host "3. Link a Device"
Write-Host "4. Link with phone number instead"
Write-Host "5. Enter: $code"
```

## Check if Working

After entering code in WhatsApp, check session status:

```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
     "http://localhost:8080/api/v1/whatsapp/sessions/YOUR_SESSION_ID/status"
```

**Expected response:**
```json
{
  "data": {
    "status": "connected",  ← Should be "connected" now
    "connected_phone": "628123456789@s.whatsapp.net"
  }
}
```

## Still Not Working?

See `DEBUG_PAIRING_CODE.md` for detailed troubleshooting.

Or provide:
1. Error message from API
2. Server logs (especially errors)
3. Phone number format (hide actual digits)
4. Session status before requesting code
