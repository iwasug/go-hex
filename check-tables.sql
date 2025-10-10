-- Check which WhatsApp tables exist in database

SELECT 
    table_name,
    CASE 
        WHEN table_name LIKE 'whatsapp%' THEN '✅'
        ELSE ''
    END as status
FROM information_schema.tables 
WHERE table_schema = 'public' 
  AND table_name LIKE 'whatsapp%'
ORDER BY table_name;

-- Count rows in each table
SELECT 'whatsapp_sessions' as table_name, COUNT(*) as row_count FROM whatsapp_sessions
UNION ALL
SELECT 'whatsapp_devices', COUNT(*) FROM whatsapp_devices
UNION ALL
SELECT 'whatsapp_messages', COUNT(*) FROM whatsapp_messages
UNION ALL
SELECT 'whatsapp_lid_mappings', COUNT(*) FROM whatsapp_lid_mappings
UNION ALL
SELECT 'whatsapp_app_state_sync_keys', COUNT(*) FROM whatsapp_app_state_sync_keys;
