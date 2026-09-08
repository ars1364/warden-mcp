package store

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("not found")

// Secret types. "opaque" is a single value (secrets.nonce/ciphertext); the other three
// store their values as encrypted rows in secret_fields instead.
const (
	TypeOpaque     = "opaque"
	TypeStructured = "structured"
	TypeTOTP       = "totp"
	TypeReference  = "reference"
)

type Secret struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Tags        []string   `json:"tags"`
	Type        string     `json:"type"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	CreatedBy   string     `json:"created_by"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// SecretWithValue is only used internally when the plaintext must be returned (MCP get_secret, edit form prefill).
type SecretWithValue struct {
	Secret
	Value string `json:"value"`
}

func (db *DB) CreateSecret(name, description string, tags []string, nonce, ciphertext []byte, expiresAt *time.Time, createdBy string) (*Secret, error) {
	tagsJSON, err := json.Marshal(tags)
	if err != nil {
		return nil, err
	}
	id := uuid.NewString()
	_, err = db.Exec(
		`INSERT INTO secrets (id, name, description, tags, nonce, ciphertext, expires_at, created_by) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		id, name, description, string(tagsJSON), nonce, ciphertext, expiresAt, createdBy,
	)
	if err != nil {
		return nil, err
	}
	return db.GetSecretMeta(id)
}

// CreateSecretMeta creates a structured/totp/reference secret's row without a top-level
// value; its actual values live in secret_fields (see ReplaceSecretFields). nonce/ciphertext
// are stored empty since the columns are NOT NULL but unused for these types.
func (db *DB) CreateSecretMeta(name, description string, tags []string, secretType string, expiresAt *time.Time, createdBy string) (*Secret, error) {
	tagsJSON, err := json.Marshal(tags)
	if err != nil {
		return nil, err
	}
	id := uuid.NewString()
	_, err = db.Exec(
		`INSERT INTO secrets (id, name, description, tags, type, nonce, ciphertext, expires_at, created_by) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, name, description, string(tagsJSON), secretType, []byte{}, []byte{}, expiresAt, createdBy,
	)
	if err != nil {
		return nil, err
	}
	return db.GetSecretMeta(id)
}

// UpdateSecretMeta updates description/tags/expiry for a structured/totp/reference secret;
// its field values are replaced separately via ReplaceSecretFields.
func (db *DB) UpdateSecretMeta(id, description string, tags []string, expiresAt *time.Time) error {
	tagsJSON, err := json.Marshal(tags)
	if err != nil {
		return err
	}
	res, err := db.Exec(
		`UPDATE secrets SET description = ?, tags = ?, expires_at = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		description, string(tagsJSON), expiresAt, id,
	)
	if err != nil {
		return err
	}
	return checkRowsAffected(res)
}

func (db *DB) UpdateSecret(id, description string, tags []string, nonce, ciphertext []byte, expiresAt *time.Time) error {
	tagsJSON, err := json.Marshal(tags)
	if err != nil {
		return err
	}
	res, err := db.Exec(
		`UPDATE secrets SET description = ?, tags = ?, nonce = ?, ciphertext = ?, expires_at = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		description, string(tagsJSON), nonce, ciphertext, expiresAt, id,
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
	rows, err := db.Query(`SELECT id, name, description, tags, type, expires_at, created_by, created_at, updated_at FROM secrets ORDER BY name`)
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
	row := db.QueryRow(`SELECT id, name, description, tags, type, expires_at, created_by, created_at, updated_at FROM secrets WHERE id = ?`, id)
	s, err := scanSecret(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// GetSecretMetaByName looks up a secret's metadata (including type) by name, used by the
// MCP tools to decide whether to read secrets.ciphertext (opaque) or secret_fields (other types).
func (db *DB) GetSecretMetaByName(name string) (*Secret, error) {
	row := db.QueryRow(`SELECT id, name, description, tags, type, expires_at, created_by, created_at, updated_at FROM secrets WHERE name = ?`, name)
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
	if err := row.Scan(&s.ID, &s.Name, &s.Description, &tagsJSON, &s.Type, &s.ExpiresAt, &s.CreatedBy, &s.CreatedAt, &s.UpdatedAt); err != nil {
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
