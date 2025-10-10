-- Simple SQL script to create LID mappings table
-- Run this directly in your PostgreSQL database

-- Create the table
CREATE TABLE IF NOT EXISTS whatsapp_lid_mappings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES whatsapp_sessions(id) ON DELETE CASCADE,
    lid TEXT NOT NULL,
    phone_number TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    UNIQUE(session_id, lid)
);

-- Create indexes for fast lookup
CREATE INDEX IF NOT EXISTS idx_whatsapp_lid_mappings_pn ON whatsapp_lid_mappings(session_id, phone_number);
CREATE INDEX IF NOT EXISTS idx_whatsapp_lid_mappings_session ON whatsapp_lid_mappings(session_id);

-- Verify table created
SELECT 'Table created successfully!' AS status;
SELECT COUNT(*) AS table_exists FROM information_schema.tables WHERE table_name = 'whatsapp_lid_mappings';
