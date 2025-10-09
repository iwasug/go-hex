-- Remove seeded data
DELETE FROM role_permissions WHERE role_id IN (
    SELECT id FROM roles WHERE name IN ('admin', 'user')
);

DELETE FROM roles WHERE name IN ('admin', 'user');

DELETE FROM permissions WHERE name IN (
    'users:read', 'users:write', 'users:delete',
    'roles:read', 'roles:write', 'roles:delete',
    'profile:read', 'profile:write'
);
