package whatsapp

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/util/keys"
)

// PostgresStore implements multiple whatsmeow store interfaces using PostgreSQL
// It implements: IdentityStore, PreKeyStore, SessionStore, SenderKeyStore
type PostgresStore struct {
	db        *sql.DB
	sessionID string
}

// NewPostgresStore creates a new PostgreSQL store for a specific session
func NewPostgresStore(db *sql.DB, sessionID string) *PostgresStore {
	return &PostgresStore{
		db:        db,
		sessionID: sessionID,
	}
}

// NewDevice creates a new device for this store
// This returns a device that will use this PostgresStore for all storage operations
func (s *PostgresStore) NewDevice() *store.Device {
	ctx := context.Background()
	
	// Try to load existing device from database
	existingDevice, err := s.loadDeviceFromDB(ctx)
	if err == nil && existingDevice != nil {
		// Device exists, return it with store interfaces configured
		existingDevice.Identities = s
		existingDevice.Sessions = s
		existingDevice.PreKeys = s
		existingDevice.SenderKeys = s
		return existingDevice
	}
	
	// No existing device, generate cryptographic keys for new device
	// These are required for the WhatsApp handshake
	noiseKey := keys.NewKeyPair()
	identityKey := keys.NewKeyPair()
	
	// Generate a signed pre-key
	signedPreKey := identityKey.CreateSignedPreKey(1)
	
	// Generate a random registration ID
	registrationID := generateRegistrationID()
	
	// Create a device with all store interfaces properly configured
	device := &store.Device{
		ID:             nil, // Will be set after pairing
		NoiseKey:       noiseKey,
		IdentityKey:    identityKey,
		SignedPreKey:   signedPreKey,
		RegistrationID: registrationID,
		Identities:     s, // IdentityStore interface
		Sessions:       s, // SessionStore interface
		PreKeys:        s, // PreKeyStore interface
		SenderKeys:     s, // SenderKeyStore interface
		LIDs:           s, // LIDStore interface
		AppStateKeys:   s, // AppStateSyncKeyStore interface
	}
	
	return device
}

// loadDeviceFromDB loads an existing device from the database
func (s *PostgresStore) loadDeviceFromDB(ctx context.Context) (*store.Device, error) {
	query := `
		SELECT jid, registration_id, identity_key, signed_pre_key, 
		       signed_pre_key_id, signed_pre_key_signature, adv_secret_key,
		       platform, business_name, push_name
		FROM whatsapp_devices
		WHERE session_id = $1
	`

	var jidStr string
	var registrationID uint32
	var identityKeyBytes, signedPreKeyBytes []byte
	var signedPreKeyID uint32
	var signedPreKeySignature []byte
	var advSecretKey []byte
	var platform, businessName, pushName sql.NullString

	err := s.db.QueryRowContext(ctx, query, s.sessionID).Scan(
		&jidStr,
		&registrationID,
		&identityKeyBytes,
		&signedPreKeyBytes,
		&signedPreKeyID,
		&signedPreKeySignature,
		&advSecretKey,
		&platform,
		&businessName,
		&pushName,
	)

	if err == sql.ErrNoRows {
		return nil, nil // No existing device
	}
	if err != nil {
		return nil, fmt.Errorf("failed to load device from database: %w", err)
	}

	// Parse JID
	jid, err := types.ParseJID(jidStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JID: %w", err)
	}

	// Reconstruct identity key
	if len(identityKeyBytes) != 32 {
		return nil, fmt.Errorf("invalid identity key length: %d", len(identityKeyBytes))
	}
	var identityKeyPriv [32]byte
	copy(identityKeyPriv[:], identityKeyBytes)
	identityKey := keys.NewKeyPairFromPrivateKey(identityKeyPriv)

	// Reconstruct signed pre-key
	// Note: The database stores the private key, we need to reconstruct the full PreKey
	if len(signedPreKeyBytes) != 32 {
		return nil, fmt.Errorf("invalid signed pre-key length: %d", len(signedPreKeyBytes))
	}
	var signedPreKeyPriv [32]byte
	copy(signedPreKeyPriv[:], signedPreKeyBytes)
	signedPreKeyPair := keys.NewKeyPairFromPrivateKey(signedPreKeyPriv)
	
	signedPreKey := &keys.PreKey{
		KeyPair: *signedPreKeyPair,
		KeyID:   signedPreKeyID,
	}

	// Create device with loaded data
	device := &store.Device{
		ID:             &jid,
		NoiseKey:       keys.NewKeyPair(), // Noise key is ephemeral
		IdentityKey:    identityKey,
		SignedPreKey:   signedPreKey,
		RegistrationID: registrationID,
		AdvSecretKey:   advSecretKey,
		Platform:       platform.String,
		BusinessName:   businessName.String,
		PushName:       pushName.String,
	}

	return device, nil
}

// saveDeviceToDB saves the device to the database
func (s *PostgresStore) saveDeviceToDB(ctx context.Context, device *store.Device) error {
	if device.ID == nil {
		return nil // Device not yet paired, nothing to save
	}

	query := `
		INSERT INTO whatsapp_devices (
			jid, session_id, registration_id, identity_key, signed_pre_key,
			signed_pre_key_id, signed_pre_key_signature, adv_secret_key,
			platform, business_name, push_name, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (session_id) DO UPDATE SET
			jid = EXCLUDED.jid,
			registration_id = EXCLUDED.registration_id,
			identity_key = EXCLUDED.identity_key,
			signed_pre_key = EXCLUDED.signed_pre_key,
			signed_pre_key_id = EXCLUDED.signed_pre_key_id,
			signed_pre_key_signature = EXCLUDED.signed_pre_key_signature,
			adv_secret_key = EXCLUDED.adv_secret_key,
			platform = EXCLUDED.platform,
			business_name = EXCLUDED.business_name,
			push_name = EXCLUDED.push_name,
			updated_at = EXCLUDED.updated_at
	`

	now := time.Now()
	_, err := s.db.ExecContext(ctx, query,
		device.ID.String(),
		s.sessionID,
		device.RegistrationID,
		device.IdentityKey.Priv[:],
		device.SignedPreKey.Priv[:],
		device.SignedPreKey.KeyID,
		[]byte{}, // signature placeholder
		device.AdvSecretKey,
		device.Platform,
		device.BusinessName,
		device.PushName,
		now,
		now,
	)

	if err != nil {
		return fmt.Errorf("failed to save device to database: %w", err)
	}

	return nil
}

// generateRegistrationID generates a random registration ID
func generateRegistrationID() uint32 {
	// Generate a random uint32 for registration ID
	// WhatsApp uses this to identify device registrations
	var id uint32
	for id == 0 {
		// Make sure we don't get 0
		id = uint32(time.Now().UnixNano() & 0xFFFFFFFF)
	}
	return id
}

// Ensure PostgresStore implements the required interfaces
var (
	_ store.IdentityStore  = (*PostgresStore)(nil)
	_ store.PreKeyStore    = (*PostgresStore)(nil)
	_ store.SessionStore   = (*PostgresStore)(nil)
	_ store.SenderKeyStore = (*PostgresStore)(nil)
)

// ============================================================================
// IdentityStore Interface Implementation
// ============================================================================

// PutIdentity stores or updates an identity key
func (s *PostgresStore) PutIdentity(ctx context.Context, address string, key [32]byte) error {
	query := `
		INSERT INTO whatsapp_identity_keys (
			address, session_id, identity_key, trust_level, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (address) DO UPDATE SET
			identity_key = EXCLUDED.identity_key,
			updated_at = EXCLUDED.updated_at
	`

	now := time.Now()
	_, err := s.db.ExecContext(ctx, query,
		address,
		s.sessionID,
		key[:],
		0, // Default trust level
		now,
		now,
	)

	if err != nil {
		return fmt.Errorf("failed to put identity: %w", err)
	}

	return nil
}

// DeleteAllIdentities removes all identity keys for a phone number
func (s *PostgresStore) DeleteAllIdentities(ctx context.Context, phone string) error {
	query := `DELETE FROM whatsapp_identity_keys WHERE session_id = $1 AND address LIKE $2`

	_, err := s.db.ExecContext(ctx, query, s.sessionID, phone+"%")
	if err != nil {
		return fmt.Errorf("failed to delete all identities: %w", err)
	}

	return nil
}

// DeleteIdentity removes a specific identity key
func (s *PostgresStore) DeleteIdentity(ctx context.Context, address string) error {
	query := `DELETE FROM whatsapp_identity_keys WHERE session_id = $1 AND address = $2`

	_, err := s.db.ExecContext(ctx, query, s.sessionID, address)
	if err != nil {
		return fmt.Errorf("failed to delete identity: %w", err)
	}

	return nil
}

// IsTrustedIdentity checks if an identity key is trusted
func (s *PostgresStore) IsTrustedIdentity(ctx context.Context, address string, key [32]byte) (bool, error) {
	query := `
		SELECT identity_key
		FROM whatsapp_identity_keys
		WHERE session_id = $1 AND address = $2
	`

	var storedKey []byte
	err := s.db.QueryRowContext(ctx, query, s.sessionID, address).Scan(&storedKey)

	if err == sql.ErrNoRows {
		// No stored identity means we trust it (first contact)
		return true, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check trusted identity: %w", err)
	}

	// Compare keys
	trusted := len(storedKey) == 32 && string(storedKey) == string(key[:])
	return trusted, nil
}

// ============================================================================
// PreKeyStore Interface Implementation
// ============================================================================

// GetOrGenPreKeys retrieves or generates pre-keys
func (s *PostgresStore) GetOrGenPreKeys(ctx context.Context, count uint32) ([]*keys.PreKey, error) {
	// First try to get existing pre-keys
	query := `
		SELECT key_id, public_key, private_key
		FROM whatsapp_pre_keys
		WHERE session_id = $1
		ORDER BY key_id
		LIMIT $2
	`

	rows, err := s.db.QueryContext(ctx, query, s.sessionID, count)
	if err != nil {
		return nil, fmt.Errorf("failed to get pre-keys: %w", err)
	}
	defer rows.Close()

	var preKeys []*keys.PreKey
	for rows.Next() {
		var keyID uint32
		var publicKey, privateKey []byte
		err := rows.Scan(&keyID, &publicKey, &privateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to scan pre-key: %w", err)
		}

		preKey := &keys.PreKey{
			KeyID: keyID,
		}
		copy(preKey.Pub[:], publicKey)
		copy(preKey.Priv[:], privateKey)
		preKeys = append(preKeys, preKey)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating pre-keys: %w", err)
	}

	// If we don't have enough, generate more
	if uint32(len(preKeys)) < count {
		needed := count - uint32(len(preKeys))
		for i := uint32(0); i < needed; i++ {
			preKey, err := s.GenOnePreKey(ctx)
			if err != nil {
				return nil, err
			}
			preKeys = append(preKeys, preKey)
		}
	}

	return preKeys, nil
}

// GenOnePreKey generates a single pre-key
func (s *PostgresStore) GenOnePreKey(ctx context.Context) (*keys.PreKey, error) {
	// Get the next key ID
	var maxKeyID sql.NullInt32
	query := `SELECT MAX(key_id) FROM whatsapp_pre_keys WHERE session_id = $1`
	err := s.db.QueryRowContext(ctx, query, s.sessionID).Scan(&maxKeyID)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get max key ID: %w", err)
	}

	nextKeyID := uint32(1)
	if maxKeyID.Valid {
		nextKeyID = uint32(maxKeyID.Int32) + 1
	}

	// Generate new key pair
	preKey := keys.NewPreKey(nextKeyID)

	// Store it
	insertQuery := `
		INSERT INTO whatsapp_pre_keys (
			session_id, key_id, public_key, private_key, created_at
		) VALUES ($1, $2, $3, $4, $5)
	`

	_, err = s.db.ExecContext(ctx, insertQuery,
		s.sessionID,
		preKey.KeyID,
		preKey.Pub[:],
		preKey.Priv[:],
		time.Now(),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to store generated pre-key: %w", err)
	}

	return preKey, nil
}

// GetPreKey retrieves a specific pre-key by ID
func (s *PostgresStore) GetPreKey(ctx context.Context, id uint32) (*keys.PreKey, error) {
	query := `
		SELECT public_key, private_key
		FROM whatsapp_pre_keys
		WHERE session_id = $1 AND key_id = $2
	`

	var publicKey, privateKey []byte
	err := s.db.QueryRowContext(ctx, query, s.sessionID, id).Scan(&publicKey, &privateKey)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get pre-key: %w", err)
	}

	preKey := &keys.PreKey{KeyID: id}
	copy(preKey.Pub[:], publicKey)
	copy(preKey.Priv[:], privateKey)

	return preKey, nil
}

// RemovePreKey removes a pre-key by its ID
func (s *PostgresStore) RemovePreKey(ctx context.Context, id uint32) error {
	query := `DELETE FROM whatsapp_pre_keys WHERE session_id = $1 AND key_id = $2`

	_, err := s.db.ExecContext(ctx, query, s.sessionID, id)
	if err != nil {
		return fmt.Errorf("failed to remove pre-key: %w", err)
	}

	return nil
}

// MarkPreKeysAsUploaded marks pre-keys as uploaded up to a certain ID
func (s *PostgresStore) MarkPreKeysAsUploaded(ctx context.Context, upToID uint32) error {
	// This is a no-op for our implementation as we don't track upload status separately
	return nil
}

// UploadedPreKeyCount returns the count of uploaded pre-keys
func (s *PostgresStore) UploadedPreKeyCount(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM whatsapp_pre_keys WHERE session_id = $1`

	var count int
	err := s.db.QueryRowContext(ctx, query, s.sessionID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get pre-key count: %w", err)
	}

	return count, nil
}

// ============================================================================
// SessionStore Interface Implementation
// ============================================================================

// GetSession retrieves a session for a specific address
func (s *PostgresStore) GetSession(ctx context.Context, address string) ([]byte, error) {
	query := `
		SELECT session_data
		FROM whatsapp_session_keys
		WHERE session_id = $1 AND address = $2
	`

	var sessionData []byte
	err := s.db.QueryRowContext(ctx, query, s.sessionID, address).Scan(&sessionData)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return sessionData, nil
}

// HasSession checks if a session exists for a specific address
func (s *PostgresStore) HasSession(ctx context.Context, address string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM whatsapp_session_keys
			WHERE session_id = $1 AND address = $2
		)
	`

	var exists bool
	err := s.db.QueryRowContext(ctx, query, s.sessionID, address).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check session existence: %w", err)
	}

	return exists, nil
}

// GetManySessions retrieves multiple sessions at once
func (s *PostgresStore) GetManySessions(ctx context.Context, addresses []string) (map[string][]byte, error) {
	if len(addresses) == 0 {
		return make(map[string][]byte), nil
	}

	// Convert []string to pq.Array for PostgreSQL
	query := `
		SELECT address, session_data
		FROM whatsapp_session_keys
		WHERE session_id = $1 AND address = ANY($2)
	`

	// Create a PostgreSQL array string
	addressArray := "{"
	for i, addr := range addresses {
		if i > 0 {
			addressArray += ","
		}
		addressArray += addr
	}
	addressArray += "}"

	rows, err := s.db.QueryContext(ctx, query, s.sessionID, addressArray)
	if err != nil {
		return nil, fmt.Errorf("failed to get many sessions: %w", err)
	}
	defer rows.Close()

	sessions := make(map[string][]byte)
	for rows.Next() {
		var address string
		var sessionData []byte
		err := rows.Scan(&address, &sessionData)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}
		sessions[address] = sessionData
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating sessions: %w", err)
	}

	return sessions, nil
}

// PutSession stores or updates a session
func (s *PostgresStore) PutSession(ctx context.Context, address string, session []byte) error {
	query := `
		INSERT INTO whatsapp_session_keys (
			session_id, address, session_data, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (session_id, address) DO UPDATE SET
			session_data = EXCLUDED.session_data,
			updated_at = EXCLUDED.updated_at
	`

	now := time.Now()
	_, err := s.db.ExecContext(ctx, query,
		s.sessionID,
		address,
		session,
		now,
		now,
	)

	if err != nil {
		return fmt.Errorf("failed to put session: %w", err)
	}

	return nil
}

// PutManySessions stores multiple sessions at once
func (s *PostgresStore) PutManySessions(ctx context.Context, sessions map[string][]byte) error {
	if len(sessions) == 0 {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO whatsapp_session_keys (
			session_id, address, session_data, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (session_id, address) DO UPDATE SET
			session_data = EXCLUDED.session_data,
			updated_at = EXCLUDED.updated_at
	`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	now := time.Now()
	for address, sessionData := range sessions {
		_, err := stmt.ExecContext(ctx, s.sessionID, address, sessionData, now, now)
		if err != nil {
			return fmt.Errorf("failed to put session in batch: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteAllSessions removes all sessions for a phone number
func (s *PostgresStore) DeleteAllSessions(ctx context.Context, phone string) error {
	query := `DELETE FROM whatsapp_session_keys WHERE session_id = $1 AND address LIKE $2`

	_, err := s.db.ExecContext(ctx, query, s.sessionID, phone+"%")
	if err != nil {
		return fmt.Errorf("failed to delete all sessions: %w", err)
	}

	return nil
}

// DeleteSession removes a specific session
func (s *PostgresStore) DeleteSession(ctx context.Context, address string) error {
	query := `DELETE FROM whatsapp_session_keys WHERE session_id = $1 AND address = $2`

	_, err := s.db.ExecContext(ctx, query, s.sessionID, address)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

// MigratePNToLID migrates a phone number to a LID
func (s *PostgresStore) MigratePNToLID(ctx context.Context, pn, lid types.JID) error {
	// Update all sessions with the old phone number to use the new LID
	query := `
		UPDATE whatsapp_session_keys
		SET address = REPLACE(address, $2, $3)
		WHERE session_id = $1 AND address LIKE $2 || '%'
	`

	_, err := s.db.ExecContext(ctx, query, s.sessionID, pn.User, lid.User)
	if err != nil {
		return fmt.Errorf("failed to migrate PN to LID: %w", err)
	}

	return nil
}

// ============================================================================
// SenderKeyStore Interface Implementation
// ============================================================================

// PutSenderKey stores or updates a sender key for group encryption
func (s *PostgresStore) PutSenderKey(ctx context.Context, group, user string, session []byte) error {
	query := `
		INSERT INTO whatsapp_sender_keys (
			session_id, group_id, sender_id, sender_key, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (session_id, group_id, sender_id) DO UPDATE SET
			sender_key = EXCLUDED.sender_key,
			updated_at = EXCLUDED.updated_at
	`

	now := time.Now()
	_, err := s.db.ExecContext(ctx, query,
		s.sessionID,
		group,
		user,
		session,
		now,
		now,
	)

	if err != nil {
		return fmt.Errorf("failed to put sender key: %w", err)
	}

	return nil
}

// GetSenderKey retrieves a sender key for group encryption
func (s *PostgresStore) GetSenderKey(ctx context.Context, group, user string) ([]byte, error) {
	query := `
		SELECT sender_key
		FROM whatsapp_sender_keys
		WHERE session_id = $1 AND group_id = $2 AND sender_id = $3
	`

	var senderKeyData []byte
	err := s.db.QueryRowContext(ctx, query, s.sessionID, group, user).Scan(&senderKeyData)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get sender key: %w", err)
	}

	return senderKeyData, nil
}

// ============================================================================
// App State Sync Key Methods (Optional - for multi-device sync)
// ============================================================================

// GetAppStateSyncKey retrieves an app state sync key
func (s *PostgresStore) GetAppStateSyncKey(ctx context.Context, keyID []byte) (*store.AppStateSyncKey, error) {
	query := `
		SELECT key_data, timestamp, fingerprint
		FROM whatsapp_app_state_sync_keys
		WHERE session_id = $1 AND key_id = $2
	`

	var keyData []byte
	var timestamp int64
	var fingerprint sql.NullString

	err := s.db.QueryRowContext(ctx, query, s.sessionID, keyID).Scan(&keyData, &timestamp, &fingerprint)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get app state sync key: %w", err)
	}

	key := &store.AppStateSyncKey{
		Data:      keyData,
		Timestamp: timestamp,
	}

	if fingerprint.Valid {
		key.Fingerprint = []byte(fingerprint.String)
	}

	return key, nil
}

// PutAppStateSyncKey stores or updates an app state sync key
func (s *PostgresStore) PutAppStateSyncKey(ctx context.Context, keyID []byte, key store.AppStateSyncKey) error {
	query := `
		INSERT INTO whatsapp_app_state_sync_keys (
			session_id, key_id, key_data, timestamp, fingerprint, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (session_id, key_id) DO UPDATE SET
			key_data = EXCLUDED.key_data,
			timestamp = EXCLUDED.timestamp,
			fingerprint = EXCLUDED.fingerprint,
			updated_at = EXCLUDED.updated_at
	`

	now := time.Now()
	_, err := s.db.ExecContext(ctx, query,
		s.sessionID,
		keyID,
		key.Data,
		key.Timestamp,
		key.Fingerprint,
		now,
		now,
	)

	if err != nil {
		return fmt.Errorf("failed to put app state sync key: %w", err)
	}

	return nil
}

// GetAppStateSyncKeyCount returns the number of app state sync keys
func (s *PostgresStore) GetAppStateSyncKeyCount(ctx context.Context) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM whatsapp_app_state_sync_keys
		WHERE session_id = $1
	`

	var count int
	err := s.db.QueryRowContext(ctx, query, s.sessionID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get app state sync key count: %w", err)
	}

	return count, nil
}

// GetLatestAppStateSyncKeyID returns the ID of the latest app state sync key
func (s *PostgresStore) GetLatestAppStateSyncKeyID(ctx context.Context) ([]byte, error) {
	query := `
		SELECT key_id
		FROM whatsapp_app_state_sync_keys
		WHERE session_id = $1
		ORDER BY timestamp DESC
		LIMIT 1
	`

	var keyID []byte
	err := s.db.QueryRowContext(ctx, query, s.sessionID).Scan(&keyID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get latest app state sync key ID: %w", err)
	}

	return keyID, nil
}

// ====================
// LIDStore Implementation
// ====================

// PutManyLIDMappings stores multiple LID to PN mappings
func (s *PostgresStore) PutManyLIDMappings(ctx context.Context, mappings []store.LIDMapping) error {
	if len(mappings) == 0 {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO whatsapp_lid_mappings (session_id, lid, phone_number, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (session_id, lid) 
		DO UPDATE SET phone_number = EXCLUDED.phone_number, updated_at = EXCLUDED.updated_at
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	now := time.Now()
	for _, mapping := range mappings {
		_, err = stmt.ExecContext(ctx, s.sessionID, mapping.LID.String(), mapping.PN.String(), now, now)
		if err != nil {
			return fmt.Errorf("failed to insert LID mapping: %w", err)
		}
	}

	return tx.Commit()
}

// PutLIDMapping stores a single LID to PN mapping
func (s *PostgresStore) PutLIDMapping(ctx context.Context, lid, jid types.JID) error {
	query := `
		INSERT INTO whatsapp_lid_mappings (session_id, lid, phone_number, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (session_id, lid) 
		DO UPDATE SET phone_number = EXCLUDED.phone_number, updated_at = EXCLUDED.updated_at
	`
	
	now := time.Now()
	_, err := s.db.ExecContext(ctx, query, s.sessionID, lid.String(), jid.String(), now, now)
	if err != nil {
		return fmt.Errorf("failed to store LID mapping: %w", err)
	}

	return nil
}

// GetPNForLID retrieves the phone number for a given LID
func (s *PostgresStore) GetPNForLID(ctx context.Context, lid types.JID) (types.JID, error) {
	query := `SELECT phone_number FROM whatsapp_lid_mappings WHERE session_id = $1 AND lid = $2`
	
	var pnStr string
	err := s.db.QueryRowContext(ctx, query, s.sessionID, lid.String()).Scan(&pnStr)
	if err == sql.ErrNoRows {
		return types.EmptyJID, nil
	} else if err != nil {
		return types.EmptyJID, fmt.Errorf("failed to get PN for LID: %w", err)
	}

	pn, err := types.ParseJID(pnStr)
	if err != nil {
		return types.EmptyJID, fmt.Errorf("failed to parse JID: %w", err)
	}

	return pn, nil
}

// GetLIDForPN retrieves the LID for a given phone number
func (s *PostgresStore) GetLIDForPN(ctx context.Context, pn types.JID) (types.JID, error) {
	query := `SELECT lid FROM whatsapp_lid_mappings WHERE session_id = $1 AND phone_number = $2`
	
	var lidStr string
	err := s.db.QueryRowContext(ctx, query, s.sessionID, pn.String()).Scan(&lidStr)
	if err == sql.ErrNoRows {
		return types.EmptyJID, nil
	} else if err != nil {
		return types.EmptyJID, fmt.Errorf("failed to get LID for PN: %w", err)
	}

	lid, err := types.ParseJID(lidStr)
	if err != nil {
		return types.EmptyJID, fmt.Errorf("failed to parse JID: %w", err)
	}

	return lid, nil
}

// GetManyLIDsForPNs retrieves LIDs for multiple phone numbers
func (s *PostgresStore) GetManyLIDsForPNs(ctx context.Context, pns []types.JID) (map[types.JID]types.JID, error) {
	if len(pns) == 0 {
		return make(map[types.JID]types.JID), nil
	}

	// Build IN clause with placeholders
	pnStrings := make([]interface{}, len(pns)+1)
	pnStrings[0] = s.sessionID
	for i, pn := range pns {
		pnStrings[i+1] = pn.String()
	}

	// Create placeholders $2, $3, $4, ...
	placeholders := "`"
	for i := range pns {
		if i > 0 {
			placeholders += ", `"
		}
		placeholders += fmt.Sprintf("$%d", i+2)
	}

	query := fmt.Sprintf(`
		SELECT lid, phone_number 
		FROM whatsapp_lid_mappings 
		WHERE session_id = $1 AND phone_number IN (%s)
	`, placeholders)

	rows, err := s.db.QueryContext(ctx, query, pnStrings...)
	if err != nil {
		return nil, fmt.Errorf("failed to query LID mappings: %w", err)
	}
	defer rows.Close()

	result := make(map[types.JID]types.JID)
	for rows.Next() {
		var lidStr, pnStr string
		if err := rows.Scan(&lidStr, &pnStr); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		lid, err := types.ParseJID(lidStr)
		if err != nil {
			continue // Skip invalid JIDs
		}

		pn, err := types.ParseJID(pnStr)
		if err != nil {
			continue // Skip invalid JIDs
		}

		result[pn] = lid
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return result, nil
}
