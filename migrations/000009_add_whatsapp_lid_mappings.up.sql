-- Add LID-PN mapping table for WhatsApp linked devices
CREATE TABLE IF NOT EXISTS whatsapp_lid_mappings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES whatsapp_sessions(id) ON DELETE CASCADE,
    lid TEXT NOT NULL,
    phone_number TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    UNIQUE(session_id, lid)
);

-- Index for fast lookup by phone number
CREATE INDEX IF NOT EXISTS idx_whatsapp_lid_mappings_pn ON whatsapp_lid_mappings(session_id, phone_number);

-- Index for fast lookup by session
CREATE INDEX IF NOT EXISTS idx_whatsapp_lid_mappings_session ON whatsapp_lid_mappings(session_id);
