-- Create whatsapp_sender_keys table (Group encryption)
CREATE TABLE IF NOT EXISTS whatsapp_sender_keys (
    id SERIAL PRIMARY KEY,
    session_id UUID NOT NULL REFERENCES whatsapp_sessions(id) ON DELETE CASCADE,
    group_id VARCHAR(255) NOT NULL,
    sender_id VARCHAR(255) NOT NULL,
    sender_key BYTEA NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_session_group_sender UNIQUE(session_id, group_id, sender_id)
);

-- Create indexes for performance optimization
CREATE INDEX idx_whatsapp_sender_keys_session_id ON whatsapp_sender_keys(session_id);
CREATE INDEX idx_whatsapp_sender_keys_group_id ON whatsapp_sender_keys(group_id);
CREATE INDEX idx_whatsapp_sender_keys_sender_id ON whatsapp_sender_keys(sender_id);
