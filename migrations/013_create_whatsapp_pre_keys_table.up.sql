-- Create whatsapp_pre_keys table
CREATE TABLE IF NOT EXISTS whatsapp_pre_keys (
    id SERIAL PRIMARY KEY,
    session_id UUID NOT NULL REFERENCES whatsapp_sessions(id) ON DELETE CASCADE,
    key_id INTEGER NOT NULL,
    public_key BYTEA NOT NULL,
    private_key BYTEA NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_session_key_id UNIQUE(session_id, key_id)
);

-- Create indexes for performance optimization
CREATE INDEX idx_whatsapp_pre_keys_session_id ON whatsapp_pre_keys(session_id);
CREATE INDEX idx_whatsapp_pre_keys_key_id ON whatsapp_pre_keys(key_id);
