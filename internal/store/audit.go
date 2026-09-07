package store

import "time"

type AuditEntry struct {
	ID         int64     `json:"id"`
	Timestamp  time.Time `json:"ts"`
	ActorType  string    `json:"actor_type"` // "user" | "mcp_key"
	ActorID    string    `json:"actor_id"`
	ActorLabel string    `json:"actor_label"`
	Action     string    `json:"action"` // "read" | "create" | "update" | "delete" | "login" | "login_failed"
	SecretName string    `json:"secret_name,omitempty"`
	IP         string    `json:"ip,omitempty"`
	Detail     string    `json:"detail,omitempty"`
}

func (db *DB) WriteAudit(e AuditEntry) error {
	_, err := db.Exec(
		`INSERT INTO audit_log (actor_type, actor_id, actor_label, action, secret_name, ip, detail) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		e.ActorType, e.ActorID, e.ActorLabel, e.Action, e.SecretName, e.IP, e.Detail,
	)
	return err
}

// ListAudit returns the most recent entries, newest first, capped at limit (max 500).
func (db *DB) ListAudit(limit int) ([]AuditEntry, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := db.Query(`SELECT id, ts, actor_type, actor_id, actor_label, action, secret_name, ip, detail
		FROM audit_log ORDER BY ts DESC, id DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []AuditEntry{}
	for rows.Next() {
		var e AuditEntry
		if err := rows.Scan(&e.ID, &e.Timestamp, &e.ActorType, &e.ActorID, &e.ActorLabel, &e.Action, &e.SecretName, &e.IP, &e.Detail); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}
