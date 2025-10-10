# Manual QR Code Generation

Jika API tidak bisa generate QR code image otomatis, Anda bisa generate manual menggunakan tool ini.

If API cannot auto-generate QR code image, you can manually generate using this tool.

---

## Method 1: Menggunakan Tool Manual (Recommended)

### Step 1: Ambil QR Code String dari API

```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
     "http://localhost:8080/api/v1/whatsapp/sessions/{session_id}/qr?format=json"
```

**Response:**
```json
{
  "success": true,
  "data": {
    "qr_code": "2@ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefg1234567890...",
    "session_id": "...",
    "expires_at": "...",
    "status": "disconnected"
  }
}
```

**Copy value dari field `qr_code`**

### Step 2: Generate Image Menggunakan Tool

```bash
cd C:\Users\iwa\iwasug\go-hex

# Generate QR code
go run cmd/qr-manual/main.go -qr "2@ABC123..." -output whatsapp-qr.png -size 512
```

**Output:**
```
Generating QR code...
- Input length: 150 characters
- Output file: whatsapp-qr.png
- Size: 512x512 pixels

✅ QR code generated successfully!
   File: whatsapp-qr.png
   Size: 5432 bytes

Scan this QR code with WhatsApp to link your account.
```

### Step 3: Scan QR Code

1. Buka file `whatsapp-qr.png`
2. Buka WhatsApp di HP
3. Tap **Linked Devices**
4. Tap **Link a Device**
5. Scan QR code dari file

---

## Method 2: Online QR Generator

### Menggunakan Website

**Recommended Sites:**
1. **qr-code-generator.com** - https://www.qr-code-generator.com
2. **qrcode.tec-it.com** - https://qrcode.tec-it.com
3. **qr.io** - https://qr.io

**Steps:**
1. Get QR string dari API (format=json)
2. Paste string ke website
3. Generate & download QR image
4. Scan dengan WhatsApp

### ⚠️ Security Warning

**JANGAN GUNAKAN online generator untuk production!**
- QR code WhatsApp berisi credential
- Bisa disalahgunakan orang lain
- Hanya untuk testing/development

**DO NOT USE online generators for production!**
- WhatsApp QR codes contain credentials
- Can be misused by others
- Only for testing/development

---

## Method 3: Menggunakan Library Lain

### Python (qrcode)

```python
import qrcode

qr_string = "2@ABC123..."  # From API
qr = qrcode.make(qr_string)
qr.save("whatsapp-qr.png")
```

### Node.js (qrcode)

```javascript
const QRCode = require('qrcode')

const qrString = "2@ABC123..."  // From API
QRCode.toFile('whatsapp-qr.png', qrString, {
  width: 512
}, function (err) {
  if (err) throw err
  console.log('QR code generated!')
})
```

### Browser JavaScript (qrcode.js)

```html
<!DOCTYPE html>
<html>
<head>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/qrcodejs/1.0.0/qrcode.min.js"></script>
</head>
<body>
    <div id="qrcode"></div>
    
    <script>
        // Get QR string from API
        const qrString = "2@ABC123...";
        
        // Generate QR
        new QRCode(document.getElementById("qrcode"), {
            text: qrString,
            width: 512,
            height: 512
        });
    </script>
</body>
</html>
```

---

## Troubleshooting

### Problem: QR String Empty

**Check:**
```bash
# Get JSON response
curl -H "Authorization: Bearer TOKEN" \
     "http://localhost:8080/api/v1/whatsapp/sessions/{id}/qr?format=json" | jq .

# Look for qr_code field
```

**If empty:**
- Check application logs
- Check WhatsApp connection
- Check database (session and device tables)
- See `TEST_QR_DEBUG.md`

### Problem: QR Code Invalid

**Symptoms:**
- WhatsApp says "Invalid QR code"
- WhatsApp says "QR code expired"

**Solutions:**
1. **Generate fresh QR** - QR expires after 60 seconds
2. **Check QR string** - Pastikan tidak ada character corruption
3. **Verify format** - QR string biasanya diawali dengan "2@"

### Problem: Cannot Scan QR

**Check:**
1. **Image quality** - Generate dengan size minimal 256px
2. **Screen brightness** - Increase brightness
3. **Camera focus** - Hold phone steady
4. **QR expiry** - Generate new QR if > 60 seconds

---

## Quick Reference

### Get QR String
```bash
curl -H "Authorization: Bearer TOKEN" \
     "http://localhost:8080/api/v1/whatsapp/sessions/{id}/qr?format=json" \
     | jq -r '.data.qr_code'
```

### Generate Image (One-liner)
```bash
QR_STRING=$(curl -H "Authorization: Bearer $TOKEN" \
     "http://localhost:8080/api/v1/whatsapp/sessions/$SESSION_ID/qr?format=json" \
     | jq -r '.data.qr_code') && \
go run cmd/qr-manual/main.go -qr "$QR_STRING" -output whatsapp-qr.png
```

### Open Generated QR (Windows)
```bash
explorer whatsapp-qr.png
```

### Check QR Details
```bash
go run cmd/qr-manual/main.go -qr "$QR_STRING" -output test.png && \
(Get-Item test.png).Length
```

---

## Why Manual Generation Might Be Needed

**Possible Issues with Auto-Generation:**
1. ❌ WhatsApp client tidak connect
2. ❌ QR channel timeout
3. ❌ Database connection issue
4. ❌ Network/firewall blocking WhatsApp servers
5. ❌ Device initialization failed

**Manual generation bypass these by:**
- ✅ Only handling string → image conversion
- ✅ No dependency on WhatsApp connection
- ✅ No database queries
- ✅ Simple, fast, reliable

---

## Notes

- QR codes expire after **60 seconds** (default timeout)
- Generate new QR if expired
- Each session needs unique QR
- QR contains authentication data - keep secure
- For production, always use API auto-generation (when fixed)
- Manual generation is for debugging/fallback only

---

## Alternative: Fix API Instead

Jika sering butuh manual generation, lebih baik **fix API-nya**:

1. Check logs: `TEST_QR_DEBUG.md`
2. Verify WhatsApp connection
3. Check database schema
4. Verify network connectivity
5. Review client initialization

**API auto-generation should work!** Manual tool ini hanya untuk fallback/debugging.
