# Database Migrations

This document describes how database migrations work in Landly.

## Overview

We use [goose](https://github.com/pressly/goose) for SQL migrations with the following principles:

- **Immutable migrations**: Once a migration is deployed, it must NEVER be modified
- **Checksum validation**: CI validates that migrations haven't changed
- **Separate from deploy**: Migrations run as a dedicated CI/CD step before deployment
- **Rollback capability**: Every migration has a `down` section

## Directory Structure

```
apps/backend/migrations/
├── 001_init_schema.sql           # First migration
├── 002_generation_chat.sql       # Second migration
├── 003_schema_versions.sql       # Third migration
└── checksums.sha256              # Checksums for validation
```

## Creating a New Migration

### 1. Generate migration file

```bash
make migration name=add_users_index
```

This creates a new file like `004_add_users_index.sql`.

### 2. Write the migration

Every migration MUST have both `Up` and `Down` sections:

```sql
-- +goose Up
-- +goose StatementBegin

CREATE INDEX idx_users_created_at ON users(created_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_users_created_at;

-- +goose StatementEnd
```

### 3. Update checksums

```bash
make migration-checksums
```

### 4. Test locally

```bash
make dev
make migration-status
```

### 5. Commit and push

```bash
git add apps/backend/migrations/
git commit -m "feat(db): add users created_at index"
git push
```

## Makefile Commands

| Command | Description |
|---------|-------------|
| `make migration name=xxx` | Create new migration |
| `make migration-checksums` | Update checksums.sha256 |
| `make migration-validate` | Validate checksums (used in CI) |
| `make migration-status` | Show pending migrations |
| `make migrate` | Apply migrations (docker-compose) |
| `make migrate-down` | Rollback last migration (docker-compose) |

## CI/CD Pipeline

```
┌─────────────────────┐     ┌─────────────────────┐     ┌─────────────────────┐     ┌─────────────────────┐
│ validate-migrations │ --> │       build         │ --> │   migrate-staging   │ --> │   deploy-staging    │
│   (check checksums) │     │   (docker image)    │     │   (goose up)        │     │   (yc deploy)       │
└─────────────────────┘     └─────────────────────┘     └─────────────────────┘     └─────────────────────┘
```

### validate-migrations

- Runs on every push to `main`
- Validates that `checksums.sha256` matches actual migration files
- **Fails if any migration was modified after deployment**

### migrate-staging

- Runs after build, before deploy
- Shows pending migrations (dry-run)
- Applies migrations with advisory lock
- Uses goose directly (not via container)

### deploy-staging

- Only runs after migrations succeed
- Deploys new container revision

## Rollback

### Via GitHub Actions (Recommended)

1. Go to Actions → "Migration Rollback"
2. Click "Run workflow"
3. Select environment: `staging`
4. Enter number of steps (1-5)
5. Type `ROLLBACK` to confirm
6. Click "Run workflow"

### Via Manual Command

```bash
# Connect to staging database
export DSN="postgres://user:pass@host:port/dbname?sslmode=require"

# Check current status
goose -dir apps/backend/migrations postgres "$DSN" status

# Rollback one migration
goose -dir apps/backend/migrations postgres "$DSN" down
```

## Best Practices

### DO

- Always write `down` migrations
- Use `IF EXISTS` / `IF NOT EXISTS` for idempotency
- Test migrations locally before pushing
- Run `make migration-checksums` after creating migration
- Keep migrations small and focused

### DON'T

- Never modify an already-deployed migration
- Don't use `DROP TABLE` without backup plan
- Don't run long-running migrations in a single transaction
- Don't add `NOT NULL` columns without defaults to existing tables

## Troubleshooting

### CI fails on "checksum validation"

**Cause**: You modified an existing migration file.

**Fix**: 
1. Revert your changes to the migration
2. Create a new migration instead
3. Run `make migration-checksums`

### Migration fails with "relation already exists"

**Cause**: Migration was partially applied.

**Fix**:
1. Check `goose_db_version` table
2. Manually fix the state
3. Re-run migration

### Need to fix a deployed migration

**Cause**: Bug in a migration that's already in production.

**Fix**:
1. Create a new "fix" migration
2. Never modify the original
3. Example: `005_fix_004_typo.sql`

## Schema Version Table

Goose creates `goose_db_version` table automatically:

```sql
SELECT * FROM goose_db_version ORDER BY id;
```

| id | version_id | is_applied | tstamp |
|----|------------|------------|--------|
| 1  | 1          | true       | 2024-01-01 12:00:00 |
| 2  | 2          | true       | 2024-01-02 12:00:00 |
| 3  | 3          | true       | 2024-01-03 12:00:00 |

