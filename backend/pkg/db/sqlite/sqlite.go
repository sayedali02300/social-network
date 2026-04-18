package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

func Open(path string) (*sql.DB, error) {
	dsn := fmt.Sprintf("file:%s?_foreign_keys=on&_busy_timeout=5000&_journal_mode=WAL", path)
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}

	// SQLite supports one writer at a time. Capping the pool to a single
	// connection serialises writes and prevents SQLITE_BUSY errors that occur
	// when multiple goroutines open competing write transactions simultaneously.
	db.SetMaxOpenConns(1)

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}

func ApplyMigrations(ctx context.Context, db *sql.DB, dir string) error {
	if err := ensureMigrationsTable(ctx, db); err != nil {
		return err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}

	var migrationFiles []string
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".up.sql") {
			continue
		}
		migrationFiles = append(migrationFiles, name)
	}

	sort.Strings(migrationFiles)

	for _, file := range migrationFiles {
		version := strings.SplitN(file, "_", 2)[0]
		applied, err := isVersionApplied(ctx, db, version)
		if err != nil {
			return err
		}
		if applied {
			continue
		}

		query, err := os.ReadFile(filepath.Join(dir, file))
		if err != nil {
			return fmt.Errorf("read migration %s: %w", file, err)
		}

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}

		if _, err := tx.ExecContext(ctx, string(query)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("run migration %s: %w", file, err)
		}

		if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations (version) VALUES (?)`, version); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("mark migration %s: %w", file, err)
		}

		if err := tx.Commit(); err != nil {
			return err
		}
	}

	return nil
}

// EnsureSessionColumns idempotently adds the ip_address / user_agent columns
// to the sessions table.  It is called unconditionally at startup so that
// databases whose migration-runner only executed the first ALTER TABLE
// statement (a known go-sqlite3 transaction quirk) still get both columns.
// SQLite returns "duplicate column name" when the column already exists;
// we intentionally ignore that error.
func EnsureSessionColumns(db *sql.DB) {
	_, _ = db.Exec(`ALTER TABLE sessions ADD COLUMN ip_address TEXT`)
	_, _ = db.Exec(`ALTER TABLE sessions ADD COLUMN user_agent TEXT`)
}

func ensureMigrationsTable(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)
	return err
}

func isVersionApplied(ctx context.Context, db *sql.DB, version string) (bool, error) {
	var exists int
	err := db.QueryRowContext(ctx, `SELECT 1 FROM schema_migrations WHERE version = ?`, version).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
