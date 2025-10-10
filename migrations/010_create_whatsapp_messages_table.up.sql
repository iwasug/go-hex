-- Create whatsapp_messages table
CREATE TABLE IF NOT EXISTS whatsapp_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES whatsapp_sessions(id) ON DELETE CASCADE,
    message_id VARCHAR(255) NOT NULL,
    direction VARCHAR(20) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    from_number VARCHAR(20) NOT NULL,
    to_number VARCHAR(20) NOT NULL,
    content TEXT NOT NULL,
    message_type VARCHAR(50) NOT NULL DEFAULT 'text',
    timestamp TIMESTAMP NOT NULL,
    webhook_delivered BOOLEAN NOT NULL DEFAULT false,
    webhook_attempts INTEGER NOT NULL DEFAULT 0,
    error_message TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_session_message_id UNIQUE(session_id, message_id)
);

-- Create indexes for performance optimization
CREATE INDEX idx_whatsapp_messages_session_id ON whatsapp_messages(session_id);
CREATE INDEX idx_whatsapp_messages_direction ON whatsapp_messages(direction);
CREATE INDEX idx_whatsapp_messages_timestamp ON whatsapp_messages(timestamp DESC);
CREATE INDEX idx_whatsapp_messages_status ON whatsapp_messages(status);
CREATE INDEX idx_whatsapp_messages_message_id ON whatsapp_messages(message_id);
CREATE INDEX idx_whatsapp_messages_from_number ON whatsapp_messages(from_number);
CREATE INDEX idx_whatsapp_messages_to_number ON whatsapp_messages(to_number);
