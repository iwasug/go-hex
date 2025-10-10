-- Revert registration_id column back to INTEGER
-- WARNING: This may fail if there are values larger than INTEGER max
ALTER TABLE whatsapp_devices 
ALTER COLUMN registration_id TYPE INTEGER;
