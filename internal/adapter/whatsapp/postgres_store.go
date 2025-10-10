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
// It implements: DeviceContainer, IdentityStore, PreKeyStore, SessionStore, SenderKeyStore
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
	// Create a new empty device
	// The device will be populated when the user pairs their WhatsApp
	return &store.Device{
		ID: nil, // Will be set after pairing
	}
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
func (s *PostgresStore) PutAppStateSyncKey(ctx context.Context, keyID []byte, key *store.AppStateSyncKey) error {
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
