package store

import "github.com/google/uuid"

// HostAddress is one IP (or hostname) a host answers to, labeled by role — e.g. "public",
// "private", "mgmt". A host commonly has several: a VPS might have a public IP and a private
// VPC one; a VM inside a physical host might only have a LAN address.
type HostAddress struct {
	Label    string `json:"label"`
	Address  string `json:"address"`
	Position int    `json:"position"`
}

// ReplaceHostAddresses atomically swaps a host's address rows for a new set, same pattern as
// ReplaceSecretFields — used for both initial creation and edits.
func (db *DB) ReplaceHostAddresses(hostID string, addrs []HostAddress) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM host_addresses WHERE host_id = ?`, hostID); err != nil {
		return err
	}

	for i, a := range addrs {
		if _, err := tx.Exec(
			`INSERT INTO host_addresses (id, host_id, label, address, position) VALUES (?, ?, ?, ?, ?)`,
			uuid.NewString(), hostID, a.Label, a.Address, i,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (db *DB) GetHostAddresses(hostID string) ([]HostAddress, error) {
	rows, err := db.Query(
		`SELECT label, address, position FROM host_addresses WHERE host_id = ? ORDER BY position`,
		hostID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []HostAddress{}
	for rows.Next() {
		var a HostAddress
		if err := rows.Scan(&a.Label, &a.Address, &a.Position); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}
