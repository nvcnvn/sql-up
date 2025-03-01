package pkg

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v4/stdlib"
	_ "github.com/ncruces/go-sqlite3"
)

type Config struct {
	DBMS             string
	ConnectionString string
	SQLFile          string
}

func Up(ctx context.Context, config *Config) error {
	var driverName string
	switch config.DBMS {
	case "postgres":
		driverName = "pgx"
	case "mysql":
		driverName = "mysql"
	case "sqlite3":
		driverName = "sqlite3"
	default:
		return fmt.Errorf("unsupported DBMS: %s", config.DBMS)
	}

	file, err := os.ReadFile(config.SQLFile)
	if err != nil {
		return fmt.Errorf("failed to read SQLFile, path %s: %w", config.SQLFile, err)
	}
	newSQLContent := string(file)

	db, err := sql.Open(driverName, config.ConnectionString)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	err = createSQLUpTable(ctx, config.DBMS, db)
	if err != nil {
		return fmt.Errorf("failed to create table sql_up: %w", err)
	}

	return applyMigrations(ctx, db, newSQLContent)

}

func createSQLUpTable(ctx context.Context, dbms string, db *sql.DB) error {
	var stmt string
	switch dbms {
	case "postgres":
		stmt = `CREATE TABLE IF NOT EXISTS sql_up (
			content TEXT NOT NULL,
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`
	case "mysql":
		stmt = `CREATE TABLE IF NOT EXISTS sql_up (
			content TEXT NOT NULL,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`
	case "sqlite3":
		stmt = `CREATE TABLE IF NOT EXISTS sql_up (
			content TEXT NOT NULL,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`
	default:
		return fmt.Errorf("unsupported DBMS: %s", dbms)
	}

	_, err := db.ExecContext(ctx, stmt)
	if err != nil {
		return fmt.Errorf("failed to create table sql_up: %w", err)
	}
	return nil
}

func applyMigrations(ctx context.Context, db *sql.DB, newSQLContent string) error {
	stmt := `SELECT content FROM sql_up`
	tx, err := db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelRepeatableRead,
	})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	currentSQLContent := ""
	err = tx.QueryRowContext(ctx, stmt).Scan(&currentSQLContent)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to query sql_up: %w", err)
	}

	if currentSQLContent == newSQLContent {
		return nil
	}

	stmt = `UPDATE sql_up SET content = $1, updated_at = NOW()`
	// use currentSQLContent as prefix, remove prefix from newSQLContent
	// then check if the string begin with '-- sql-up'
	// if yes, apply the diff

	diff := strings.TrimPrefix(newSQLContent, currentSQLContent)
	if !strings.HasPrefix(diff, "-- sql-up") {
		return fmt.Errorf("new sql content does not start with '-- sql-up'")
	}

	fmt.Println("currentSQLContent", currentSQLContent)
	fmt.Println("newSQLContent", newSQLContent)
	fmt.Println("diff", diff)

	_, err = tx.ExecContext(ctx, diff)
	if err != nil {
		return fmt.Errorf("failed to apply new update: %w", err)
	}

	_, err = tx.ExecContext(ctx, stmt, newSQLContent)
	if err != nil {
		return fmt.Errorf("failed to update sql_up state table: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
