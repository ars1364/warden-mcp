CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    totp_secret TEXT,
    totp_enabled INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- `type` and `expires_at` are added via addColumnIfMissing in db.go for pre-existing
-- databases; CREATE TABLE only covers fresh installs.
CREATE TABLE IF NOT EXISTS secrets (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    tags TEXT NOT NULL DEFAULT '[]',
    type TEXT NOT NULL DEFAULT 'opaque',
    nonce BLOB NOT NULL,
    ciphertext BLOB NOT NULL,
    expires_at TIMESTAMP,
    created_by TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Per-field encrypted values for non-opaque secret types (structured / totp / reference).
-- Opaque secrets keep using secrets.nonce/ciphertext directly and never have rows here.
CREATE TABLE IF NOT EXISTS secret_fields (
    id TEXT PRIMARY KEY,
    secret_id TEXT NOT NULL REFERENCES secrets(id) ON DELETE CASCADE,
    field_key TEXT NOT NULL,
    position INTEGER NOT NULL DEFAULT 0,
    nonce BLOB NOT NULL,
    ciphertext BLOB NOT NULL,
    UNIQUE(secret_id, field_key)
);

CREATE INDEX IF NOT EXISTS idx_secret_fields_secret_id ON secret_fields(secret_id, position);

CREATE TABLE IF NOT EXISTS mcp_api_keys (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    key_hash TEXT NOT NULL UNIQUE,
    scopes TEXT NOT NULL DEFAULT '["read"]',
    created_by TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_used_at TIMESTAMP,
    last_used_ip TEXT,
    revoked_at TIMESTAMP,
    revoked_by TEXT
);

CREATE TABLE IF NOT EXISTS audit_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    ts TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    actor_type TEXT NOT NULL,
    actor_id TEXT NOT NULL,
    actor_label TEXT NOT NULL DEFAULT '',
    action TEXT NOT NULL,
    secret_name TEXT NOT NULL DEFAULT '',
    ip TEXT NOT NULL DEFAULT '',
    detail TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_audit_log_ts ON audit_log(ts DESC);
CREATE INDEX IF NOT EXISTS idx_secrets_name ON secrets(name);
