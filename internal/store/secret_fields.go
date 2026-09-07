package store

import "github.com/google/uuid"

// FieldCiphertext is one encrypted key/value pair belonging to a structured/totp/reference
// secret. Encryption happens in the caller, same pattern as secrets.go's nonce/ciphertext
// params, so this package never touches the crypto.Box.
type FieldCiphertext struct {
	Key        string
	Position   int
	Nonce      []byte
	Ciphertext []byte
}

// ReplaceSecretFields atomically swaps a secret's field rows for a new set, preserving the
// order given in fields via Position. Used for both initial creation and edits.
func (db *DB) ReplaceSecretFields(secretID string, fields []FieldCiphertext) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM secret_fields WHERE secret_id = ?`, secretID); err != nil {
		return err
	}

	for i, f := range fields {
		if _, err := tx.Exec(
			`INSERT INTO secret_fields (id, secret_id, field_key, position, nonce, ciphertext) VALUES (?, ?, ?, ?, ?, ?)`,
			uuid.NewString(), secretID, f.Key, i, f.Nonce, f.Ciphertext,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetSecretFields returns a secret's encrypted fields in display order; the caller decrypts
// each with crypto.Box.
func (db *DB) GetSecretFields(secretID string) ([]FieldCiphertext, error) {
	rows, err := db.Query(
		`SELECT field_key, position, nonce, ciphertext FROM secret_fields WHERE secret_id = ? ORDER BY position`,
		secretID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []FieldCiphertext{}
	for rows.Next() {
		var f FieldCiphertext
		if err := rows.Scan(&f.Key, &f.Position, &f.Nonce, &f.Ciphertext); err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

// GetSecretFieldsByName is a convenience for callers (MCP tools) that only have the name.
func (db *DB) GetSecretFieldsByName(name string) (secretID string, fields []FieldCiphertext, err error) {
	meta, err := db.GetSecretMetaByName(name)
	if err != nil {
		return "", nil, err
	}
	fields, err = db.GetSecretFields(meta.ID)
	return meta.ID, fields, err
}
