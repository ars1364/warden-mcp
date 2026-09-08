// Package store owns the SQLite connection and schema migration.
package store

import (
	"database/sql"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaSQL string

type DB struct {
	*sql.DB
}

func Open(path string) (*DB, error) {
	if dir := filepath.Dir(path); dir != "." {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return nil, fmt.Errorf("create db dir: %w", err)
		}
	}

	sqlDB, err := sql.Open("sqlite", path+"?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)")
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	sqlDB.SetMaxOpenConns(1) // modernc.org/sqlite is not safe for concurrent writers

	if _, err := sqlDB.Exec(schemaSQL); err != nil {
		return nil, fmt.Errorf("apply schema: %w", err)
	}

	if err := addColumnIfMissing(sqlDB, "secrets", "type", `ALTER TABLE secrets ADD COLUMN type TEXT NOT NULL DEFAULT 'opaque'`); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	if err := addColumnIfMissing(sqlDB, "secrets", "expires_at", `ALTER TABLE secrets ADD COLUMN expires_at TIMESTAMP`); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return &DB{sqlDB}, nil
}

// addColumnIfMissing backfills a column added after a database's initial schema was applied.
// ALTER TABLE ADD COLUMN has no "IF NOT EXISTS" form in SQLite, so we check PRAGMA table_info
// first; schema.sql's CREATE TABLE already covers fresh installs.
func addColumnIfMissing(db *sql.DB, table, column, alterSQL string) error {
	rows, err := db.Query(`PRAGMA table_info(` + table + `)`)
	if err != nil {
		return err
	}
	defer rows.Close()

	has := false
	for rows.Next() {
		var cid int
		var name, colType string
		var notNull, pk int
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &colType, &notNull, &dflt, &pk); err != nil {
			return err
		}
		if name == column {
			has = true
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	if !has {
		if _, err := db.Exec(alterSQL); err != nil {
			return err
		}
	}
	return nil
}
