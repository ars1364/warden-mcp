package store

import (
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	TOTPSecret   string    `json:"-"`
	TOTPEnabled  bool      `json:"totp_enabled"`
	CreatedAt    time.Time `json:"created_at"`
}

func (db *DB) CreateUser(username, passwordHash string) (*User, error) {
	id := uuid.NewString()
	_, err := db.Exec(`INSERT INTO users (id, username, password_hash) VALUES (?, ?, ?)`, id, username, passwordHash)
	if err != nil {
		return nil, err
	}
	return db.GetUserByID(id)
}

func (db *DB) GetUserByUsername(username string) (*User, error) {
	row := db.QueryRow(`SELECT id, username, password_hash, totp_secret, totp_enabled, created_at FROM users WHERE username = ?`, username)
	return scanUser(row)
}

func (db *DB) GetUserByID(id string) (*User, error) {
	row := db.QueryRow(`SELECT id, username, password_hash, totp_secret, totp_enabled, created_at FROM users WHERE id = ?`, id)
	return scanUser(row)
}

func (db *DB) CountUsers() (int, error) {
	var n int
	err := db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&n)
	return n, err
}

func (db *DB) SetUserTOTP(id, secret string, enabled bool) error {
	_, err := db.Exec(`UPDATE users SET totp_secret = ?, totp_enabled = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, secret, enabled, id)
	return err
}

func (db *DB) SetUserPassword(id, passwordHash string) error {
	_, err := db.Exec(`UPDATE users SET password_hash = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, passwordHash, id)
	return err
}

func scanUser(row rowScanner) (*User, error) {
	var u User
	var totpSecret sql.NullString
	if err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &totpSecret, &u.TOTPEnabled, &u.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	u.TOTPSecret = totpSecret.String
	return &u, nil
}
