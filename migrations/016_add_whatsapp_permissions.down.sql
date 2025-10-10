-- Remove WhatsApp permissions from role_permissions
DELETE FROM role_permissions 
WHERE permission_id IN (
    SELECT id FROM permissions 
    WHERE name IN ('whatsapp:read', 'whatsapp:write', 'whatsapp:delete')
);

-- Remove WhatsApp permissions
DELETE FROM permissions 
WHERE name IN ('whatsapp:read', 'whatsapp:write', 'whatsapp:delete');
