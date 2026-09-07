package store

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("not found")

type Secret struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Tags        []string  `json:"tags"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// SecretWithValue is only used internally when the plaintext must be returned (MCP get_secret, edit form prefill).
type SecretWithValue struct {
	Secret
	Value string `json:"value"`
}

func (db *DB) CreateSecret(name, description string, tags []string, nonce, ciphertext []byte, createdBy string) (*Secret, error) {
	tagsJSON, err := json.Marshal(tags)
	if err != nil {
		return nil, err
	}
	id := uuid.NewString()
	_, err = db.Exec(
		`INSERT INTO secrets (id, name, description, tags, nonce, ciphertext, created_by) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		id, name, description, string(tagsJSON), nonce, ciphertext, createdBy,
	)
	if err != nil {
		return nil, err
	}
	return db.GetSecretMeta(id)
}

func (db *DB) UpdateSecret(id, description string, tags []string, nonce, ciphertext []byte) error {
	tagsJSON, err := json.Marshal(tags)
	if err != nil {
		return err
	}
	res, err := db.Exec(
		`UPDATE secrets SET description = ?, tags = ?, nonce = ?, ciphertext = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		description, string(tagsJSON), nonce, ciphertext, id,
	)
	if err != nil {
		return err
	}
	return checkRowsAffected(res)
}

func (db *DB) DeleteSecret(id string) error {
	res, err := db.Exec(`DELETE FROM secrets WHERE id = ?`, id)
	if err != nil {
		return err
	}
	return checkRowsAffected(res)
}

func (db *DB) ListSecrets() ([]Secret, error) {
	rows, err := db.Query(`SELECT id, name, description, tags, created_by, created_at, updated_at FROM secrets ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []Secret{}
	for rows.Next() {
		s, err := scanSecret(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (db *DB) GetSecretMeta(id string) (*Secret, error) {
	row := db.QueryRow(`SELECT id, name, description, tags, created_by, created_at, updated_at FROM secrets WHERE id = ?`, id)
	s, err := scanSecret(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// GetSecretValueByName returns the decrypted nonce/ciphertext pair for the caller to open; decryption happens
// in the caller (which holds the crypto.Box) so this package stays free of the encryption key.
func (db *DB) GetSecretValueByName(name string) (id string, nonce, ciphertext []byte, err error) {
	row := db.QueryRow(`SELECT id, nonce, ciphertext FROM secrets WHERE name = ?`, name)
	err = row.Scan(&id, &nonce, &ciphertext)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil, nil, ErrNotFound
	}
	return id, nonce, ciphertext, err
}

func (db *DB) GetSecretValueByID(id string) (nonce, ciphertext []byte, err error) {
	row := db.QueryRow(`SELECT nonce, ciphertext FROM secrets WHERE id = ?`, id)
	err = row.Scan(&nonce, &ciphertext)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil, ErrNotFound
	}
	return nonce, ciphertext, err
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanSecret(row rowScanner) (Secret, error) {
	var s Secret
	var tagsJSON string
	if err := row.Scan(&s.ID, &s.Name, &s.Description, &tagsJSON, &s.CreatedBy, &s.CreatedAt, &s.UpdatedAt); err != nil {
		return Secret{}, err
	}
	if err := json.Unmarshal([]byte(tagsJSON), &s.Tags); err != nil || s.Tags == nil {
		s.Tags = []string{}
	}
	return s, nil
}

func checkRowsAffected(res sql.Result) error {
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
