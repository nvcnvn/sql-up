# sql-up

Always forward schema migration because you cannot just rollback in production.

## Usage

```bash
sql-up --dbms postgres --connection-string <connection-string> --sql-file <sql-file>
```

**sql-up** will apply the migration after lasted-non applied magic marker `-- sql-up` in the SQL file.

```sql
-- Create table
CREATE TABLE users (
  id SERIAL PRIMARY KEY,
  name TEXT NOT NULL
);

-- sql-up
ALTER TABLE users ADD COLUMN email TEXT NULL;

-- sql-up
ALTER TABLE users ADD COLUMN address TEXT NULL;
```

**sql-up** does that by creating a new table `sql_up` in the database to store the last applied file content and compare with the input file content to determine the migration. This mean **the SQL file must be append-only**.

## Supported Databases and Connection String

- postgres: https://github.com/jackc/pgx, `postgres://user:password@localhost:5432/dbname?sslmode=disable`
- mysql: https://github.com/go-sql-driver/mysql/, `user:password@tcp(localhost:3306)/dbname`
- sqlite3: https://github.com/ncruces/go-sqlite3, `file:/path/to/db.sqlite3`
