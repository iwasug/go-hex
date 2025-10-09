-- Seed default permissions
INSERT INTO permissions (name, resource, action, description) VALUES
    ('users:read', 'users', 'read', 'View users'),
    ('users:write', 'users', 'write', 'Create/update users'),
    ('users:delete', 'users', 'delete', 'Delete users'),
    ('roles:read', 'roles', 'read', 'View roles'),
    ('roles:write', 'roles', 'write', 'Create/update roles'),
    ('roles:delete', 'roles', 'delete', 'Delete roles'),
    ('profile:read', 'profile', 'read', 'View own profile'),
    ('profile:write', 'profile', 'write', 'Update own profile')
ON CONFLICT (name) DO NOTHING;

-- Seed default roles
INSERT INTO roles (name, description) VALUES
    ('admin', 'Full system access'),
    ('user', 'Basic user access')
ON CONFLICT (name) DO NOTHING;

-- Assign all permissions to admin role
INSERT INTO role_permissions (role_id, permission_id)
SELECT 
    (SELECT id FROM roles WHERE name = 'admin'),
    id
FROM permissions
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Assign profile permissions to user role
INSERT INTO role_permissions (role_id, permission_id)
SELECT 
    (SELECT id FROM roles WHERE name = 'user'),
    id
FROM permissions
WHERE name IN ('profile:read', 'profile:write')
ON CONFLICT (role_id, permission_id) DO NOTHING;
