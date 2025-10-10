-- Fix registration_id column to use BIGINT instead of INTEGER
-- This fixes the "value out of range for type integer" error
ALTER TABLE whatsapp_devices 
ALTER COLUMN registration_id TYPE BIGINT;
