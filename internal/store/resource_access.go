package store

import (
	"database/sql"
	"errors"

	"github.com/google/uuid"
)

// Resource types a grant can target.
const (
	ResourceSecret = "secret"
	ResourceHost   = "host"
)

// ResourceGrant is one row of an API key's allowlist. Read and write are independent — a
// grant can be read-only, write-only (e.g. a rotation-only agent that never reads the value
// back), or both.
type ResourceGrant struct {
	ResourceType string `json:"resource_type"`
	ResourceName string `json:"resource_name"`
	CanRead      bool   `json:"can_read"`
	CanWrite     bool   `json:"can_write"`
}

// ReplaceResourceGrants atomically swaps an API key's grants for one resource_type — same
// replace-all pattern as ReplaceSecretFields/ReplaceHostAddresses. Grants for the OTHER
// resource_type (e.g. host grants when replacing secret grants) are left untouched. A grant
// with neither CanRead nor CanWrite set is silently dropped — it would be a no-op row.
func (db *DB) ReplaceResourceGrants(apiKeyID, resourceType string, grants []ResourceGrant) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(
		`DELETE FROM api_key_resource_access WHERE api_key_id = ? AND resource_type = ?`,
		apiKeyID, resourceType,
	); err != nil {
		return err
	}

	for _, g := range grants {
		if !g.CanRead && !g.CanWrite {
			continue
		}
		if _, err := tx.Exec(
			`INSERT INTO api_key_resource_access (id, api_key_id, resource_type, resource_name, can_read, can_write) VALUES (?, ?, ?, ?, ?, ?)`,
			uuid.NewString(), apiKeyID, resourceType, g.ResourceName, boolToInt(g.CanRead), boolToInt(g.CanWrite),
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// ListResourceGrants returns every grant for an API key, both resource types — used by the
// REST API to show the current allowlist for editing.
func (db *DB) ListResourceGrants(apiKeyID string) ([]ResourceGrant, error) {
	rows, err := db.Query(
		`SELECT resource_type, resource_name, can_read, can_write FROM api_key_resource_access
		 WHERE api_key_id = ? ORDER BY resource_type, resource_name`,
		apiKeyID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []ResourceGrant{}
	for rows.Next() {
		var g ResourceGrant
		var canRead, canWrite int
		if err := rows.Scan(&g.ResourceType, &g.ResourceName, &canRead, &canWrite); err != nil {
			return nil, err
		}
		g.CanRead = canRead != 0
		g.CanWrite = canWrite != 0
		out = append(out, g)
	}
	return out, rows.Err()
}

// ResourceAccess reports what one API key may do with one named resource. When the key has
// no grants at all for resourceType, it is unrestricted (canRead=canWrite=true) — this is
// what keeps every pre-existing key working exactly as before this feature existed. Once a
// key has at least one grant for that type, it becomes an allowlist and an ungranted name is
// fully denied (canRead=canWrite=false), so a restricted key can't tell an ungranted name
// apart from one that doesn't exist. A granted name's access is exactly its row's flags —
// read and write are independent, neither implies the other.
func (db *DB) ResourceAccess(apiKeyID, resourceType, resourceName string) (canRead, canWrite bool, err error) {
	var total int
	if err := db.QueryRow(
		`SELECT COUNT(*) FROM api_key_resource_access WHERE api_key_id = ? AND resource_type = ?`,
		apiKeyID, resourceType,
	).Scan(&total); err != nil {
		return false, false, err
	}
	if total == 0 {
		return true, true, nil
	}

	var canReadInt, canWriteInt int
	err = db.QueryRow(
		`SELECT can_read, can_write FROM api_key_resource_access WHERE api_key_id = ? AND resource_type = ? AND resource_name = ?`,
		apiKeyID, resourceType, resourceName,
	).Scan(&canReadInt, &canWriteInt)
	if errors.Is(err, sql.ErrNoRows) {
		return false, false, nil // restricted, and this name isn't on the allowlist
	}
	if err != nil {
		return false, false, err
	}
	return canReadInt != 0, canWriteInt != 0, nil
}
