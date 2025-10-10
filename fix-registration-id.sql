-- Fix registration_id column to use BIGINT instead of INTEGER
-- This fixes the "value out of range for type integer" error
-- Run this directly in your PostgreSQL database

ALTER TABLE whatsapp_devices 
ALTER COLUMN registration_id TYPE BIGINT;

-- Verify the change
SELECT 
    column_name, 
    data_type, 
    character_maximum_length,
    numeric_precision
FROM information_schema.columns 
WHERE table_name = 'whatsapp_devices' 
  AND column_name = 'registration_id';

SELECT 'registration_id column updated to BIGINT successfully!' AS status;
