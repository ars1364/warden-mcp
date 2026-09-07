package store

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

type APIKey struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Scopes     []string   `json:"scopes"`
	CreatedBy  string     `json:"created_by"`
	CreatedAt  time.Time  `json:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	LastUsedIP string     `json:"last_used_ip,omitempty"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty"`
}

func (db *DB) CreateAPIKey(name string, scopes []string, keyHash, createdBy string) (*APIKey, error) {
	scopesJSON, err := json.Marshal(scopes)
	if err != nil {
		return nil, err
	}
	id := uuid.NewString()
	_, err = db.Exec(`INSERT INTO mcp_api_keys (id, name, key_hash, scopes, created_by) VALUES (?, ?, ?, ?, ?)`,
		id, name, keyHash, string(scopesJSON), createdBy)
	if err != nil {
		return nil, err
	}
	return db.getAPIKeyByID(id)
}

func (db *DB) ListAPIKeys() ([]APIKey, error) {
	rows, err := db.Query(`SELECT id, name, scopes, created_by, created_at, last_used_at, last_used_ip, revoked_at FROM mcp_api_keys ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []APIKey{}
	for rows.Next() {
		k, err := scanAPIKey(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *k)
	}
	return out, rows.Err()
}

// FindActiveByHash returns the key row matching keyHash, or ErrNotFound if absent or revoked.
func (db *DB) FindActiveByHash(keyHash string) (*APIKey, error) {
	row := db.QueryRow(`SELECT id, name, scopes, created_by, created_at, last_used_at, last_used_ip, revoked_at
		FROM mcp_api_keys WHERE key_hash = ? AND revoked_at IS NULL`, keyHash)
	return scanAPIKey(row)
}

func (db *DB) TouchAPIKey(id, ip string) error {
	_, err := db.Exec(`UPDATE mcp_api_keys SET last_used_at = CURRENT_TIMESTAMP, last_used_ip = ? WHERE id = ?`, ip, id)
	return err
}

func (db *DB) RevokeAPIKey(id, revokedBy string) error {
	res, err := db.Exec(`UPDATE mcp_api_keys SET revoked_at = CURRENT_TIMESTAMP, revoked_by = ? WHERE id = ? AND revoked_at IS NULL`, revokedBy, id)
	if err != nil {
		return err
	}
	return checkRowsAffected(res)
}

func (db *DB) getAPIKeyByID(id string) (*APIKey, error) {
	row := db.QueryRow(`SELECT id, name, scopes, created_by, created_at, last_used_at, last_used_ip, revoked_at FROM mcp_api_keys WHERE id = ?`, id)
	return scanAPIKey(row)
}

func scanAPIKey(row rowScanner) (*APIKey, error) {
	var k APIKey
	var scopesJSON string
	var lastUsedAt sql.NullTime
	var lastUsedIP sql.NullString
	var revokedAt sql.NullTime
	if err := row.Scan(&k.ID, &k.Name, &scopesJSON, &k.CreatedBy, &k.CreatedAt, &lastUsedAt, &lastUsedIP, &revokedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	_ = json.Unmarshal([]byte(scopesJSON), &k.Scopes)
	if k.Scopes == nil {
		k.Scopes = []string{}
	}
	if lastUsedAt.Valid {
		k.LastUsedAt = &lastUsedAt.Time
	}
	k.LastUsedIP = lastUsedIP.String
	if revokedAt.Valid {
		k.RevokedAt = &revokedAt.Time
	}
	return &k, nil
}
