# hexago add migration

Add a database migration to an existing project.

## Synopsis

```shell
hexago add migration <name>
```

Must be run from the project root directory.

---

## Description

Generates a pair of SQL migration files with sequential numbering. Uses [golang-migrate](https://github.com/golang-migrate/migrate) conventions.

---

## Examples

```shell
hexago add migration create_users_table
hexago add migration add_email_index
hexago add migration alter_products_table
hexago add migration create_orders_table
hexago add migration add_foreign_keys
```

---

## Generated Files

For `hexago add migration create_users_table` (assuming this is migration #1):

```
migrations/
├── 000001_create_users_table.up.sql    # Apply migration
└── 000001_create_users_table.down.sql  # Rollback migration
```

Subsequent migrations are automatically numbered sequentially:

```
migrations/
├── 000001_create_users_table.up.sql
├── 000001_create_users_table.down.sql
├── 000002_add_email_index.up.sql
├── 000002_add_email_index.down.sql
├── 000003_create_orders_table.up.sql
└── 000003_create_orders_table.down.sql
```

---

## Generated File Contents

**`000001_create_users_table.up.sql`:**

```sql
-- Migration: create_users_table
-- TODO: Add your UP migration SQL here

-- Example:
-- CREATE TABLE users (
--     id UUID PRIMARY KEY,
--     name VARCHAR(255) NOT NULL,
--     email VARCHAR(255) UNIQUE NOT NULL,
--     created_at TIMESTAMP NOT NULL DEFAULT NOW()
-- );
```

**`000001_create_users_table.down.sql`:**

```sql
-- Migration: create_users_table (rollback)
-- TODO: Add your DOWN migration SQL here

-- Example:
-- DROP TABLE IF EXISTS users;
```

---

## Running Migrations

Generated projects with `--with-migrations` include Makefile targets:

```shell
make migrate-up       # Apply all pending migrations
make migrate-down     # Rollback the last migration
make migrate-version  # Show current migration version
```

Or run directly with golang-migrate:

```shell
migrate -path ./migrations -database "postgres://..." up
migrate -path ./migrations -database "postgres://..." down 1
```

---

## Prerequisites

Migrations require the project was initialized with `--with-migrations`:

```shell
hexago init my-app --module github.com/me/my-app --with-migrations
```

Or the migrations directory must exist.
