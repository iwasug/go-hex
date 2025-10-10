# Seed Users - Quick Start

## Default Credentials

After running migrations, you can login with these default users:

### 🔑 Admin User
```
Email: admin@example.com
Password: password123
```

### 👤 Regular User
```
Email: user@example.com
Password: password123
```

## Quick Login

```bash
# Admin Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"password123"}'

# Regular User Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}'
```

## ⚠️ Security Warning

**These are development/testing credentials only!**

- Change passwords immediately in production
- Delete these users if not needed
- Never use `password123` in production

## More Information

See `migrations/SEED_USERS.md` for complete documentation.
