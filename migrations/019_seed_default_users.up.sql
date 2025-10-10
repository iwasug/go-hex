-- Seed default users
-- Note: Default password for both users is "password123"
-- Password hash generated with bcrypt cost 10

-- Insert admin user
INSERT INTO users (id, email, password, name, is_deleted, created_at, updated_at) VALUES
    ('00000000-0000-0000-0000-000000000001', 'admin@example.com', '$2a$10$m.ILIyoi4MEZIc/QyKVRT.5dZVLbMGI.Z0.qvGR7cKYau10O.70GS', 'Admin User', FALSE, NOW(), NOW())
ON CONFLICT (email) DO NOTHING;

-- Insert regular user
INSERT INTO users (id, email, password, name, is_deleted, created_at, updated_at) VALUES
    ('00000000-0000-0000-0000-000000000002', 'user@example.com', '$2a$10$m.ILIyoi4MEZIc/QyKVRT.5dZVLbMGI.Z0.qvGR7cKYau10O.70GS', 'Regular User', FALSE, NOW(), NOW())
ON CONFLICT (email) DO NOTHING;

-- Assign admin role to admin user
INSERT INTO user_roles (user_id, role_id)
SELECT 
    '00000000-0000-0000-0000-000000000001',
    id
FROM roles
WHERE name = 'admin'
ON CONFLICT (user_id, role_id) DO NOTHING;

-- Assign user role to regular user
INSERT INTO user_roles (user_id, role_id)
SELECT 
    '00000000-0000-0000-0000-000000000002',
    id
FROM roles
WHERE name = 'user'
ON CONFLICT (user_id, role_id) DO NOTHING;
