-- Create whatsapp_session_keys table (Signal protocol sessions)
CREATE TABLE IF NOT EXISTS whatsapp_session_keys (
    id SERIAL PRIMARY KEY,
    session_id UUID NOT NULL REFERENCES whatsapp_sessions(id) ON DELETE CASCADE,
    address VARCHAR(255) NOT NULL,
    session_data BYTEA NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_session_address UNIQUE(session_id, address)
);

-- Create indexes for performance optimization
CREATE INDEX idx_whatsapp_session_keys_session_id ON whatsapp_session_keys(session_id);
CREATE INDEX idx_whatsapp_session_keys_address ON whatsapp_session_keys(address);
