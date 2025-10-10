-- Remove reconnection tracking fields from whatsapp_sessions table
DROP INDEX IF EXISTS idx_whatsapp_sessions_reconnect;

ALTER TABLE whatsapp_sessions
DROP COLUMN IF EXISTS reconnect_attempts,
DROP COLUMN IF EXISTS last_reconnect_attempt;
