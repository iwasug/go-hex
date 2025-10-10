-- Create whatsapp_sessions table
CREATE TABLE IF NOT EXISTS whatsapp_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    session_name VARCHAR(255) NOT NULL,
    phone_number VARCHAR(20),
    status VARCHAR(50) NOT NULL DEFAULT 'disconnected',
    qr_code TEXT,
    qr_code_expiry TIMESTAMP,
    device_info JSONB,
    last_activity TIMESTAMP NOT NULL DEFAULT NOW(),
    webhook_url TEXT,
    webhook_secret TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    updated_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_user_session_name UNIQUE(user_id, session_name)
);

-- Create indexes for performance optimization
CREATE INDEX idx_whatsapp_sessions_user_id ON whatsapp_sessions(user_id);
CREATE INDEX idx_whatsapp_sessions_status ON whatsapp_sessions(status);
CREATE INDEX idx_whatsapp_sessions_phone_number ON whatsapp_sessions(phone_number);
CREATE INDEX idx_whatsapp_sessions_is_active ON whatsapp_sessions(is_active);
CREATE INDEX idx_whatsapp_sessions_last_activity ON whatsapp_sessions(last_activity DESC);
