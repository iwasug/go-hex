-- Create whatsapp_app_state_sync_keys table (for multi-device sync)
CREATE TABLE IF NOT EXISTS whatsapp_app_state_sync_keys (
    id SERIAL PRIMARY KEY,
    session_id UUID NOT NULL REFERENCES whatsapp_sessions(id) ON DELETE CASCADE,
    key_id BYTEA NOT NULL,
    key_data BYTEA NOT NULL,
    timestamp BIGINT NOT NULL,
    fingerprint BYTEA,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_session_key_id UNIQUE(session_id, key_id)
);

-- Create indexes for performance optimization
CREATE INDEX idx_whatsapp_app_state_sync_keys_session_id ON whatsapp_app_state_sync_keys(session_id);
CREATE INDEX idx_whatsapp_app_state_sync_keys_key_id ON whatsapp_app_state_sync_keys(key_id);
