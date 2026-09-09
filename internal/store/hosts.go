package store

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

// Host is one entry in the asset inventory: a physical machine, VM, VPS, switch, router, etc.
// SSHSecretName points at a row in `secrets` by name rather than duplicating credential
// material here — the inventory records where to find a key/password, never the value itself.
type Host struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	HostType         string    `json:"host_type"`
	Status           string    `json:"status"`
	Description      string    `json:"description"`
	Tags             []string  `json:"tags"`
	ParentHostID     *string   `json:"parent_host_id,omitempty"`
	LocationKind     string    `json:"location_kind"`
	CloudProvider    string    `json:"cloud_provider"`
	CloudAccount     string    `json:"cloud_account"`
	PhysicalLocation string    `json:"physical_location"`
	SSHPort          int       `json:"ssh_port"`
	SSHUsername      string    `json:"ssh_username"`
	SSHSecretName    string    `json:"ssh_secret_name"`
	SSHJumpHostID    *string   `json:"ssh_jump_host_id,omitempty"`
	CreatedBy        string    `json:"created_by"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// HostInput carries every writable Host field for create/update, as a struct rather than a
// long positional argument list — a mismatched positional arg count is exactly the class of
// bug an earlier change in this codebase shipped with (see UpdateSecretMeta's fixed bug).
type HostInput struct {
	Name             string
	HostType         string
	Status           string
	Description      string
	Tags             []string
	ParentHostID     *string
	LocationKind     string
	CloudProvider    string
	CloudAccount     string
	PhysicalLocation string
	SSHPort          int
	SSHUsername      string
	SSHSecretName    string
	SSHJumpHostID    *string
}

const hostColumns = `id, name, host_type, status, description, tags, parent_host_id,
	location_kind, cloud_provider, cloud_account, physical_location,
	ssh_port, ssh_username, ssh_secret_name, ssh_jump_host_id,
	created_by, created_at, updated_at`

func (db *DB) CreateHost(in HostInput, createdBy string) (*Host, error) {
	tagsJSON, err := json.Marshal(in.Tags)
	if err != nil {
		return nil, err
	}
	id := uuid.NewString()
	_, err = db.Exec(
		`INSERT INTO hosts (
			id, name, host_type, status, description, tags, parent_host_id,
			location_kind, cloud_provider, cloud_account, physical_location,
			ssh_port, ssh_username, ssh_secret_name, ssh_jump_host_id, created_by
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, in.Name, in.HostType, in.Status, in.Description, string(tagsJSON), in.ParentHostID,
		in.LocationKind, in.CloudProvider, in.CloudAccount, in.PhysicalLocation,
		in.SSHPort, in.SSHUsername, in.SSHSecretName, in.SSHJumpHostID, createdBy,
	)
	if err != nil {
		return nil, err
	}
	return db.GetHostByID(id)
}

func (db *DB) UpdateHost(id string, in HostInput) error {
	tagsJSON, err := json.Marshal(in.Tags)
	if err != nil {
		return err
	}
	res, err := db.Exec(
		`UPDATE hosts SET
			host_type = ?, status = ?, description = ?, tags = ?, parent_host_id = ?,
			location_kind = ?, cloud_provider = ?, cloud_account = ?, physical_location = ?,
			ssh_port = ?, ssh_username = ?, ssh_secret_name = ?, ssh_jump_host_id = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`,
		in.HostType, in.Status, in.Description, string(tagsJSON), in.ParentHostID,
		in.LocationKind, in.CloudProvider, in.CloudAccount, in.PhysicalLocation,
		in.SSHPort, in.SSHUsername, in.SSHSecretName, in.SSHJumpHostID, id,
	)
	if err != nil {
		return err
	}
	return checkRowsAffected(res)
}

func (db *DB) DeleteHost(id string) error {
	res, err := db.Exec(`DELETE FROM hosts WHERE id = ?`, id)
	if err != nil {
		return err
	}
	return checkRowsAffected(res)
}

func (db *DB) ListHosts() ([]Host, error) {
	rows, err := db.Query(`SELECT ` + hostColumns + ` FROM hosts ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []Host{}
	for rows.Next() {
		h, err := scanHost(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	return out, rows.Err()
}

func (db *DB) GetHostByID(id string) (*Host, error) {
	row := db.QueryRow(`SELECT `+hostColumns+` FROM hosts WHERE id = ?`, id)
	h, err := scanHost(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &h, nil
}

func (db *DB) GetHostByName(name string) (*Host, error) {
	row := db.QueryRow(`SELECT `+hostColumns+` FROM hosts WHERE name = ?`, name)
	h, err := scanHost(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &h, nil
}

func scanHost(row rowScanner) (Host, error) {
	var h Host
	var tagsJSON string
	if err := row.Scan(
		&h.ID, &h.Name, &h.HostType, &h.Status, &h.Description, &tagsJSON, &h.ParentHostID,
		&h.LocationKind, &h.CloudProvider, &h.CloudAccount, &h.PhysicalLocation,
		&h.SSHPort, &h.SSHUsername, &h.SSHSecretName, &h.SSHJumpHostID,
		&h.CreatedBy, &h.CreatedAt, &h.UpdatedAt,
	); err != nil {
		return Host{}, err
	}
	if err := json.Unmarshal([]byte(tagsJSON), &h.Tags); err != nil || h.Tags == nil {
		h.Tags = []string{}
	}
	return h, nil
}
