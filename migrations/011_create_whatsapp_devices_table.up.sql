-- Create whatsapp_devices table (whatsmeow store)
CREATE TABLE IF NOT EXISTS whatsapp_devices (
    jid VARCHAR(255) PRIMARY KEY,
    session_id UUID NOT NULL REFERENCES whatsapp_sessions(id) ON DELETE CASCADE,
    registration_id INTEGER NOT NULL,
    identity_key BYTEA NOT NULL,
    signed_pre_key BYTEA NOT NULL,
    signed_pre_key_id INTEGER NOT NULL,
    signed_pre_key_signature BYTEA NOT NULL,
    adv_secret_key BYTEA,
    platform VARCHAR(50) DEFAULT 'go',
    business_name VARCHAR(255),
    push_name VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_device_session_id UNIQUE(session_id)
);

-- Create indexes for performance optimization
CREATE INDEX idx_whatsapp_devices_session_id ON whatsapp_devices(session_id);
CREATE INDEX idx_whatsapp_devices_jid ON whatsapp_devices(jid);
