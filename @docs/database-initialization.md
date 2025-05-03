# Database Initialization and Migration

This document outlines the database initialization and migration process for the product and inventory services.

## Overview

We've implemented automatic database creation and migration for both the product service and inventory service. This ensures that:

1. The database is created if it doesn't exist
2. Tables are created if the database exists but tables don't
3. Migrations are applied in the correct order

## Implementation Details

### Database Initialization

Both services now follow this process during startup:

1. Connect to the PostgreSQL server using the `postgres` database
2. Check if the service-specific database exists
3. Create the database if it doesn't exist
4. Connect to the service-specific database
5. Run migrations to create or update tables

### Migration Management

We've implemented a simple migration system that:

1. Creates a `schema_migrations` table to track applied migrations
2. Reads migration files from the `migrations` directory
3. Applies migrations in order based on their version number
4. Records each successful migration in the `schema_migrations` table

### Migration Files

Migration files follow this naming convention:

```
000001_init_schema.up.sql   # For applying migrations
000001_init_schema.down.sql # For reverting migrations
```

The version number (e.g., `000001`) is used to determine the order of migrations.

## Benefits

This implementation provides several benefits:

1. **Simplified Setup**: Developers don't need to manually create databases or tables
2. **Consistent Environments**: All environments (development, staging, production) use the same schema
3. **Version Control**: Database schema changes are tracked in version control
4. **Automated Deployment**: New deployments automatically apply necessary migrations

## Usage

No special action is required to use this feature. When the services start:

1. If the database doesn't exist, it will be created
2. If tables don't exist, they will be created
3. If new migrations are available, they will be applied

## Troubleshooting

### Dirty Migrations

If a migration fails halfway through, it can leave the database in a "dirty" state. To fix this:

1. Manually fix any issues in the database
2. Use the `ForceFixMigration` function to mark the migration as applied:

```go
err := db.ForceFixMigration(dbConfig.Master, "000007", log)
if err != nil {
    log.Fatal("Failed to fix dirty migration", zap.Error(err))
}
```

### Migration Errors

Common migration errors include:

1. **Syntax errors**: Check the SQL syntax in your migration files
2. **Dependency errors**: Ensure migrations are ordered correctly
3. **Constraint violations**: Check that data constraints are satisfied
