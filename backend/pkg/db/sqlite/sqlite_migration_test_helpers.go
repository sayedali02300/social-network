package sqlite

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
)

func setupMigratedDB(t *testing.T) *sql.DB {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "migrations-test.db")
	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	migrationsDir := filepath.Join("..", "migrations", "sqlite")
	if err := ApplyMigrations(context.Background(), db, migrationsDir); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}

	return db
}

func tableExists(db *sql.DB, tableName string) (bool, error) {
	var exists int
	err := db.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM sqlite_master WHERE type = 'table' AND name = ?)`,
		tableName,
	).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists == 1, nil
}

func tableColumns(db *sql.DB, tableName string) (map[string]bool, error) {
	rows, err := db.Query("PRAGMA table_info(" + tableName + ")")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns := make(map[string]bool)
	for rows.Next() {
		var (
			cid      int
			name     string
			colType  string
			notNull  int
			defaultV interface{}
			primaryK int
		)
		if err := rows.Scan(&cid, &name, &colType, &notNull, &defaultV, &primaryK); err != nil {
			return nil, err
		}
		columns[name] = true
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return columns, nil
}
