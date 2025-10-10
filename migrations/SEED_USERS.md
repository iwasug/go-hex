# Seed Users Documentation

This document describes the default users created by the `019_seed_default_users.up.sql` migration.

## Default Users

### Admin User
- **Email**: `admin@example.com`
- **Password**: `password123`
- **Name**: Admin User
- **Role**: admin
- **Permissions**: Full system access (all permissions)

### Regular User
- **Email**: `user@example.com`
- **Password**: `password123`
- **Name**: Regular User
- **Role**: user
- **Permissions**: Basic user access (profile:read, profile:write)

## Security Notes

⚠️ **IMPORTANT**: These are default seed users for development and testing purposes only.

### Production Deployment

Before deploying to production, you should:

1. **Change the default passwords** immediately after first login
2. **Delete or disable** these seed users if not needed
3. **Create proper admin accounts** with strong passwords
4. **Never use** `password123` in production

### Changing Default Passwords

You can change passwords using the API:

```bash
# Login as admin
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "password123"
  }'

# Use the token to update password (implement password change endpoint)
```

### Deleting Seed Users

To remove seed users, run the down migration:

```bash
# Using psql directly
psql -h localhost -U postgres -d iwapi -f migrations/019_seed_default_users.down.sql

# Or using make (if implemented)
make migrate-down
```

## Login Examples

### Admin Login

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "password123"
  }'
```

Response:
```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "access_token": "eyJhbGc...",
    "refresh_token": "eyJhbGc...",
    "user": {
      "id": "00000000-0000-0000-0000-000000000001",
      "email": "admin@example.com",
      "name": "Admin User",
      "roles": ["admin"]
    }
  }
}
```

### Regular User Login

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

## Testing with Seed Users

### Admin User Testing

The admin user has full access and can:
- Create/read/update/delete users
- Manage roles and permissions
- Access all WhatsApp sessions
- Perform all administrative tasks

```bash
# Example: List all users (admin only)
curl -X GET http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

### Regular User Testing

The regular user has limited access and can:
- View and update their own profile
- Create and manage their own WhatsApp sessions
- Send and receive messages through their sessions

```bash
# Example: Get own profile
curl -X GET http://localhost:8080/api/v1/users/profile \
  -H "Authorization: Bearer $USER_TOKEN"
```

## Migration Details

### Up Migration (019_seed_default_users.up.sql)

Creates two users with predefined UUIDs:
- Admin: `00000000-0000-0000-0000-000000000001`
- User: `00000000-0000-0000-0000-000000000002`

Uses `ON CONFLICT (email) DO NOTHING` to prevent duplicate entries if migration is run multiple times.

### Down Migration (019_seed_default_users.down.sql)

Removes both seed users by email. The `user_roles` entries are automatically deleted due to foreign key cascade constraints.

## Password Hash Generation

The password hash is generated using bcrypt with cost factor 10:

```go
package main

import (
    "fmt"
    "golang.org/x/crypto/bcrypt"
)

func main() {
    password := "password123"
    hash, _ := bcrypt.GenerateFromPassword([]byte(password), 10)
    fmt.Println(string(hash))
}
```

Current hash: `$2a$10$m.ILIyoi4MEZIc/QyKVRT.5dZVLbMGI.Z0.qvGR7cKYau10O.70GS`

## Troubleshooting

### Users Already Exist

If you see conflicts when running the migration, it means users with these emails already exist. The migration uses `ON CONFLICT DO NOTHING`, so it will skip existing users.

### Cannot Login

1. Verify the migration ran successfully
2. Check the password is exactly `password123`
3. Verify the user exists in the database:
   ```sql
   SELECT id, email, name FROM users WHERE email IN ('admin@example.com', 'user@example.com');
   ```

### Missing Roles

If users exist but have no roles:
1. Ensure migration `008_seed_default_roles_permissions.up.sql` ran first
2. Check the `user_roles` table:
   ```sql
   SELECT u.email, r.name 
   FROM users u
   JOIN user_roles ur ON u.id = ur.user_id
   JOIN roles r ON ur.role_id = r.id
   WHERE u.email IN ('admin@example.com', 'user@example.com');
   ```

## Best Practices

1. **Development**: Use seed users for quick testing and development
2. **Staging**: Create environment-specific users, don't rely on seeds
3. **Production**: Never use seed users, create proper accounts
4. **CI/CD**: Seed users are useful for automated testing
5. **Documentation**: Always document default credentials clearly

## Related Migrations

- `001_create_users_table.up.sql` - Creates users table
- `002_create_roles_table.up.sql` - Creates roles table
- `004_create_user_roles_table.up.sql` - Creates user-role relationship
- `008_seed_default_roles_permissions.up.sql` - Seeds default roles and permissions
