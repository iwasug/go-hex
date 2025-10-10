-- Remove default users
-- This will cascade delete user_roles entries due to foreign key constraints

DELETE FROM users WHERE email IN ('admin@example.com', 'user@example.com');
