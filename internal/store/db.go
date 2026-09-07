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

	if err := migrateAddSecretType(sqlDB); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return &DB{sqlDB}, nil
}

// migrateAddSecretType backfills the secrets.type column for databases created before
// it existed. ALTER TABLE ADD COLUMN has no "IF NOT EXISTS" form in SQLite, so we check
// PRAGMA table_info first; schema.sql's CREATE TABLE already covers fresh installs.
func migrateAddSecretType(db *sql.DB) error {
	rows, err := db.Query(`PRAGMA table_info(secrets)`)
	if err != nil {
		return err
	}
	defer rows.Close()

	hasType := false
	for rows.Next() {
		var cid int
		var name, colType string
		var notNull, pk int
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &colType, &notNull, &dflt, &pk); err != nil {
			return err
		}
		if name == "type" {
			hasType = true
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	if !hasType {
		if _, err := db.Exec(`ALTER TABLE secrets ADD COLUMN type TEXT NOT NULL DEFAULT 'opaque'`); err != nil {
			return err
		}
	}
	return nil
}
