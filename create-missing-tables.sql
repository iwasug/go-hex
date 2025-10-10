-- Script to create missing WhatsApp tables
-- Run this in your PostgreSQL database (iwapi)

-- 1. Create app_state_sync_keys table (if not exists)
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

CREATE INDEX IF NOT EXISTS idx_whatsapp_app_state_sync_keys_session_id 
ON whatsapp_app_state_sync_keys(session_id);

CREATE INDEX IF NOT EXISTS idx_whatsapp_app_state_sync_keys_key_id 
ON whatsapp_app_state_sync_keys(key_id);

-- 2. Create lid_mappings table (if not exists - backup)
CREATE TABLE IF NOT EXISTS whatsapp_lid_mappings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES whatsapp_sessions(id) ON DELETE CASCADE,
    lid TEXT NOT NULL,
    phone_number TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    UNIQUE(session_id, lid)
);

CREATE INDEX IF NOT EXISTS idx_whatsapp_lid_mappings_pn 
ON whatsapp_lid_mappings(session_id, phone_number);

CREATE INDEX IF NOT EXISTS idx_whatsapp_lid_mappings_session 
ON whatsapp_lid_mappings(session_id);

-- Verify tables created
SELECT 
    'whatsapp_app_state_sync_keys' as table_name,
    CASE WHEN EXISTS (
        SELECT 1 FROM information_schema.tables 
        WHERE table_name = 'whatsapp_app_state_sync_keys'
    ) THEN '✅ EXISTS' ELSE '❌ NOT FOUND' END as status
UNION ALL
SELECT 
    'whatsapp_lid_mappings',
    CASE WHEN EXISTS (
        SELECT 1 FROM information_schema.tables 
        WHERE table_name = 'whatsapp_lid_mappings'
    ) THEN '✅ EXISTS' ELSE '❌ NOT FOUND' END;

-- Show all WhatsApp tables
SELECT table_name 
FROM information_schema.tables 
WHERE table_schema = 'public' 
  AND table_name LIKE 'whatsapp%'
ORDER BY table_name;
