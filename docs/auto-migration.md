# Auto-Migration Feature

## Overview

The application now supports automatic database migrations on startup using the `golang-migrate` library. This feature ensures your database schema is always up-to-date without manual intervention.

## Configuration

Auto-migration is configured via environment variables:

```env
# Enable/disable auto-migration (default: true)
DB_AUTO_MIGRATE=true

# Path to migration files (default: migrations)
DB_MIGRATIONS_PATH=migrations
```

## How It Works

1. **Application Startup**: When the application starts, it connects to the database
2. **Migration Check**: If `DB_AUTO_MIGRATE=true`, it checks for pending migrations
3. **Migration Execution**: Runs all pending migrations in order (001, 002, 003, etc.)
4. **Version Tracking**: Logs the current migration version and status
5. **Server Start**: Starts the HTTP server after successful migration

## Migration Files

Migration files are located in the `migrations/` directory and follow this naming convention:

```
migrations/
├── 001_create_users_table.up.sql
├── 001_create_users_table.down.sql
├── 002_create_roles_table.up.sql
├── 002_create_roles_table.down.sql
└── ...
```

- **`.up.sql`**: Applied when migrating forward
- **`.down.sql`**: Applied when rolling back

## Benefits

### Development
- No need to manually run migrations
- Consistent database state across team members
- Faster onboarding for new developers

### Production
- Zero-downtime deployments with automatic schema updates
- Consistent deployment process
- Reduced human error

### CI/CD
- Automated testing with fresh database schemas
- Integration tests always run against the latest schema
- Simplified deployment pipelines

## Usage Examples

### Enable Auto-Migration (Default)

```env
DB_AUTO_MIGRATE=true
DB_MIGRATIONS_PATH=migrations
```

When you start the application:
```bash
go run cmd/api/main.go
```

You'll see logs like:
```
INFO: Database connection established successfully
INFO: Running database migrations... migrations_path=migrations
INFO: Database migrations completed version=8 dirty=false
INFO: Starting HTTP server port=8080
```

### Disable Auto-Migration

If you prefer manual control:

```env
DB_AUTO_MIGRATE=false
```

Then run migrations manually:
```bash
make migrate-up
```

### Custom Migration Path

If your migrations are in a different directory:

```env
DB_AUTO_MIGRATE=true
DB_MIGRATIONS_PATH=db/migrations
```

## Error Handling

If a migration fails:
1. The application logs the error with details
2. The application exits with status code 1
3. The database is left in a consistent state (transaction rollback)
4. Fix the migration file and restart the application

Example error log:
```
ERROR: Failed to run migrations error="migration failed: syntax error at line 5"
```

## Migration Version Tracking

The migration system creates a `schema_migrations` table to track:
- Current migration version
- Dirty state (indicates incomplete migration)

Check migration status:
```sql
SELECT * FROM schema_migrations;
```

## Best Practices

1. **Always test migrations locally** before deploying to production
2. **Keep migrations small and focused** on a single change
3. **Write both up and down migrations** for rollback capability
4. **Never modify existing migrations** that have been deployed
5. **Use transactions** in migration files when possible
6. **Test rollback migrations** to ensure they work correctly

## Troubleshooting

### Migration Fails on Startup

**Problem**: Application exits with migration error

**Solution**:
1. Check the error message in logs
2. Review the failing migration file
3. Fix the SQL syntax or logic
4. Restart the application

### Dirty Migration State

**Problem**: Migration is marked as "dirty"

**Solution**:
```sql
-- Check current state
SELECT * FROM schema_migrations;

-- Manually fix the issue and force version
UPDATE schema_migrations SET dirty = false;
```

### Migrations Not Running

**Problem**: Migrations don't execute on startup

**Solution**:
1. Verify `DB_AUTO_MIGRATE=true` in your `.env` file
2. Check that migration files exist in the specified path
3. Ensure database connection is successful
4. Check file permissions on migration directory

## Manual Migration Commands

If you need manual control, use these commands:

```bash
# Run all pending migrations
make migrate-up

# Rollback last migration
make migrate-down

# Create new migration
make migrate-create NAME=add_new_table
```

## Integration with Docker

The Docker setup automatically runs migrations on container startup:

```yaml
# docker-compose.yml
services:
  app:
    environment:
      - DB_AUTO_MIGRATE=true
      - DB_MIGRATIONS_PATH=migrations
```

## Security Considerations

- Migration files should be version controlled
- Sensitive data should not be in migration files
- Use environment variables for credentials
- Review migrations in code reviews
- Test migrations in staging before production

## Performance

- Migrations run only once per version
- Subsequent startups skip already-applied migrations
- Minimal overhead on application startup
- No impact on runtime performance

## Comparison with Manual Migrations

| Feature | Auto-Migration | Manual Migration |
|---------|---------------|------------------|
| Setup Complexity | Low | Medium |
| Human Error | Minimal | Higher |
| CI/CD Integration | Seamless | Requires scripts |
| Rollback | Automatic | Manual |
| Development Speed | Fast | Slower |
| Production Safety | High (with testing) | High (with process) |

## Conclusion

Auto-migration simplifies database schema management and reduces deployment complexity. It's recommended for most use cases, especially in containerized environments and CI/CD pipelines.

For production deployments, always:
1. Test migrations in staging
2. Have a rollback plan
3. Monitor migration logs
4. Keep backups of your database
