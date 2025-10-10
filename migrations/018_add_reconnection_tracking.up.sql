-- Add reconnection tracking fields to whatsapp_sessions table
ALTER TABLE whatsapp_sessions
ADD COLUMN reconnect_attempts INTEGER NOT NULL DEFAULT 0,
ADD COLUMN last_reconnect_attempt TIMESTAMP;

-- Add index for sessions that need reconnection monitoring
CREATE INDEX idx_whatsapp_sessions_reconnect ON whatsapp_sessions(status, reconnect_attempts)
WHERE status IN ('disconnected', 'connecting');

-- Add comment for documentation
COMMENT ON COLUMN whatsapp_sessions.reconnect_attempts IS 'Number of automatic reconnection attempts made';
COMMENT ON COLUMN whatsapp_sessions.last_reconnect_attempt IS 'Timestamp of the last reconnection attempt';
