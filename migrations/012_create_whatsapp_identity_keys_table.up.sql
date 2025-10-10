-- Create whatsapp_identity_keys table
CREATE TABLE IF NOT EXISTS whatsapp_identity_keys (
    address VARCHAR(255) PRIMARY KEY,
    session_id UUID NOT NULL REFERENCES whatsapp_sessions(id) ON DELETE CASCADE,
    identity_key BYTEA NOT NULL,
    trust_level INTEGER DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create indexes for performance optimization
CREATE INDEX idx_whatsapp_identity_keys_session_id ON whatsapp_identity_keys(session_id);
CREATE INDEX idx_whatsapp_identity_keys_address ON whatsapp_identity_keys(address);
