-- Add WhatsApp permissions
INSERT INTO permissions (name, resource, action, description) VALUES
    ('whatsapp:read', 'whatsapp', 'read', 'View WhatsApp sessions and messages'),
    ('whatsapp:write', 'whatsapp', 'write', 'Create sessions and send messages'),
    ('whatsapp:delete', 'whatsapp', 'delete', 'Delete WhatsApp sessions')
ON CONFLICT (name) DO NOTHING;

-- Assign WhatsApp permissions to admin role
INSERT INTO role_permissions (role_id, permission_id)
SELECT 
    (SELECT id FROM roles WHERE name = 'admin'),
    id
FROM permissions
WHERE name IN ('whatsapp:read', 'whatsapp:write', 'whatsapp:delete')
ON CONFLICT (role_id, permission_id) DO NOTHING;
