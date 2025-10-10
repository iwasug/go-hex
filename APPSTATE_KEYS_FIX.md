# App State Sync Keys Fix

## Problem Resolved

**Panic:**
```
panic: runtime error: invalid memory address or nil pointer dereference
at go.mau.fi/whatsmeow.(*Client).handleAppStateSyncKeyShare
message.go:741
```

**Root Cause:**  
Device.AppStateKeys was NIL - PostgresStore had partial implementation of AppStateSyncKeyStore interface with incorrect signature and missing method.

## Solution Implemented

### 1. Fixed AppStateSyncKeyStore Interface

**Methods Implemented:**
- ✅ `PutAppStateSyncKey()` - Store app state sync key (signature fixed - removed pointer)
- ✅ `GetAppStateSyncKey()` - Retrieve app state sync key
- ✅ `GetLatestAppStateSyncKeyID()` - Get latest sync key ID (NEW method)

### 2. Set Device.AppStateKeys Field

```go
device := &store.Device{
    ...
    Identities:   s,
    Sessions:     s,
    PreKeys:      s,
    SenderKeys:   s,
    LIDs:         s,
    AppStateKeys: s, // ← ADDED
}
```

### 3. Fixed Method Signature

**Before:**
```go
func (s *PostgresStore) PutAppStateSyncKey(ctx context.Context, keyID []byte, key *store.AppStateSyncKey) error
```

**After:**
```go
func (s *PostgresStore) PutAppStateSyncKey(ctx context.Context, keyID []byte, key store.AppStateSyncKey) error
// Removed pointer - WhatsApp expects value type
```

## What is App State Sync Key?

**App State Sync Keys** are encryption keys used by WhatsApp multi-device to synchronize app state across devices:

- Contact list
- Chat settings
- Pinned chats
- Muted chats
- Archived chats
- User preferences

When you pair a new device, WhatsApp shares these keys to sync your account state.

## Database Structure

**Table:** `whatsapp_app_state_sync_keys` (already exists from migration 017)

```sql
CREATE TABLE whatsapp_app_state_sync_keys (
    id SERIAL PRIMARY KEY,
    session_id UUID REFERENCES whatsapp_sessions,
    key_id BYTEA NOT NULL,
    key_data BYTEA NOT NULL,
    timestamp BIGINT NOT NULL,
    fingerprint BYTEA,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    UNIQUE(session_id, key_id)
);
```

**Fields:**
- `key_id` - Unique identifier for sync key
- `key_data` - Encrypted key data
- `timestamp` - Key creation timestamp
- `fingerprint` - Key fingerprint for verification

## Files Modified

1. ✅ `internal/adapter/whatsapp/postgres_store.go`
   - Fixed `PutAppStateSyncKey` signature (removed pointer)
   - Added `GetLatestAppStateSyncKeyID` method
   - Set `Device.AppStateKeys = s`

## How to Test

### Step 1: Restart Server

```powershell
cd C:\Users\iwa\iwasug\go-hex

# Stop current server (Ctrl+C)

# Start new server
.\bin\server.exe
```

### Step 2: Generate Pairing Code

```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
     "http://localhost:8080/api/v1/whatsapp/sessions/{SESSION_ID}/pairing-code?phone_number=628123456789"
```

### Step 3: Enter Code in WhatsApp

1. Open WhatsApp on mobile
2. Menu → Linked Devices → Link a Device
3. Link with phone number instead
4. Enter the 8-digit code

### Step 4: Wait for Connection

Watch server logs for:
```
INFO: handling connected event
INFO: connected event handled successfully
DEBUG: Received app state sync key XXXXX
```

### Step 5: Verify Success

**Expected logs (NO PANIC):**
```
✅ Connected event
✅ App state sync keys received
✅ Keys stored to database
✅ Session fully operational
```

**Verify in database:**
```sql
SELECT 
    session_id,
    key_id,
    timestamp,
    created_at
FROM whatsapp_app_state_sync_keys
ORDER BY created_at DESC
LIMIT 5;
```

Should show newly created sync keys!

## Progress Summary

**Store Interfaces Implemented:**

| Interface                | Status | Notes                          |
|--------------------------|--------|--------------------------------|
| IdentityStore            | ✅     | Identity keys storage          |
| SessionStore             | ✅     | Signal protocol sessions       |
| PreKeyStore              | ✅     | Pre-keys for encryption        |
| SenderKeyStore           | ✅     | Group message encryption       |
| LIDStore                 | ✅     | Linked device ID mapping       |
| AppStateSyncKeyStore     | ✅     | **JUST FIXED**                 |
| AppStateStore            | ❓     | May be needed for full sync    |
| ContactStore             | ❓     | May be needed                  |
| MsgSecretStore           | ❓     | May be needed                  |
| PrivacyTokenStore        | ❓     | Optional                       |
| ChatSettingsStore        | ❓     | Optional                       |

## Expected Behavior After Fix

**Pairing Flow:**

1. ✅ Generate pairing code
2. ✅ Enter in WhatsApp
3. ✅ WhatsApp accepts code
4. ✅ Connection established
5. ✅ LID-PN mapping stored
6. ✅ **App state sync keys received** ← FIXED
7. ✅ **Sync keys stored to database** ← FIXED
8. ✅ Session fully connected

**If panic still occurs**, it means we need to implement additional store interfaces (AppStateStore, ContactStore, etc.)

## Troubleshooting

### Server Still Crashes?

Check the **new panic location**:
```bash
# Look for different panic message in logs
# Example:
# panic at handleAppStateSync
# panic at handleContacts
# etc.
```

If different panic, it means we need to implement another store interface.

### No Sync Keys in Database?

```sql
-- Check if keys are being stored
SELECT COUNT(*) FROM whatsapp_app_state_sync_keys;

-- If 0, check logs for errors:
-- "Failed to store app state sync key"
```

### Connection Successful but No Messages?

May need to implement:
- **ContactStore** - For contact list sync
- **AppStateStore** - For full app state sync
- **MsgSecretStore** - For message secrets

## Build Status

✅ **BUILD SUCCESS**  
✅ **AppStateSyncKeyStore complete**  
✅ **All 3 methods implemented**  
✅ **Device.AppStateKeys set**  
✅ **Table already exists (migration 017)**  
✅ **Ready to test**

## Next Steps

1. **Restart server** with new binary
2. **Test pairing code** again
3. **Watch logs** for app state sync messages
4. **Check database** for stored sync keys
5. **If still panic**, implement next required store interface

## Technical Details

### AppStateSyncKeyStore Methods

```go
// Store a sync key
PutAppStateSyncKey(ctx, keyID, AppStateSyncKey{
    Data:        []byte{...},
    Fingerprint: []byte{...},
    Timestamp:   1234567890,
})

// Get a sync key
key := GetAppStateSyncKey(ctx, keyID)

// Get latest key ID
latestID := GetLatestAppStateSyncKeyID(ctx)
```

### How WhatsApp Uses These Keys

1. **After pairing**, WhatsApp server sends app state sync keys
2. **Client stores** keys to database via `PutAppStateSyncKey`
3. **For each sync operation**, client retrieves keys via `GetAppStateSyncKey`
4. **To check sync status**, client gets latest key via `GetLatestAppStateSyncKeyID`
5. **Keys are used** to encrypt/decrypt app state patches

### Key Rotation

WhatsApp rotates sync keys periodically for security:
- Old keys kept for historical data
- New keys generated for fresh sync operations
- `GetLatestAppStateSyncKeyID` ensures using current key

## Summary

✅ **Implemented:** Complete AppStateSyncKeyStore interface  
✅ **Fixed:** Method signature (value instead of pointer)  
✅ **Added:** GetLatestAppStateSyncKeyID method  
✅ **Set:** Device.AppStateKeys field  
✅ **Ready:** Table already exists in database  

**Result:** Pairing should progress past app state sync without panic! 🚀

**Next:** Restart server → Test pairing → May need more store interfaces for full functionality.
