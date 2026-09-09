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

-- Asset inventory: physical machines, VMs, VPSes, network gear. parent_host_id models
-- "this VM lives inside that physical host"; ssh_secret_name POINTS at a row in `secrets`
-- by name rather than duplicating credential material — the inventory never stores a key
-- or password itself, only where to find one.
CREATE TABLE IF NOT EXISTS hosts (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    host_type TEXT NOT NULL DEFAULT 'other',
    status TEXT NOT NULL DEFAULT 'active',
    description TEXT NOT NULL DEFAULT '',
    tags TEXT NOT NULL DEFAULT '[]',
    parent_host_id TEXT REFERENCES hosts(id) ON DELETE SET NULL,
    location_kind TEXT NOT NULL DEFAULT '',
    cloud_provider TEXT NOT NULL DEFAULT '',
    cloud_account TEXT NOT NULL DEFAULT '',
    physical_location TEXT NOT NULL DEFAULT '',
    ssh_port INTEGER NOT NULL DEFAULT 22,
    ssh_username TEXT NOT NULL DEFAULT '',
    ssh_secret_name TEXT NOT NULL DEFAULT '',
    ssh_jump_host_id TEXT REFERENCES hosts(id) ON DELETE SET NULL,
    created_by TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS host_addresses (
    id TEXT PRIMARY KEY,
    host_id TEXT NOT NULL REFERENCES hosts(id) ON DELETE CASCADE,
    label TEXT NOT NULL,
    address TEXT NOT NULL,
    position INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_hosts_name ON hosts(name);
CREATE INDEX IF NOT EXISTS idx_hosts_parent ON hosts(parent_host_id);
CREATE INDEX IF NOT EXISTS idx_host_addresses_host_id ON host_addresses(host_id, position);
CREATE INDEX IF NOT EXISTS idx_host_addresses_address ON host_addresses(address);

-- Row-level access for MCP API keys. An API key with ZERO rows here for a given
-- resource_type is UNRESTRICTED for that type — sees/touches everything, exactly today's
-- behavior, so every existing key keeps working unchanged. A key with ANY row for a type
-- becomes an allowlist: it can only see/act on the resource_names present, and only write
-- to ones where can_write=1. This is enforced in the MCP tool layer only — the web UI
-- (JWT-authenticated human) always has full access; it's the one configuring these grants.
CREATE TABLE IF NOT EXISTS api_key_resource_access (
    id TEXT PRIMARY KEY,
    api_key_id TEXT NOT NULL REFERENCES mcp_api_keys(id) ON DELETE CASCADE,
    resource_type TEXT NOT NULL,
    resource_name TEXT NOT NULL,
    can_write INTEGER NOT NULL DEFAULT 0,
    UNIQUE(api_key_id, resource_type, resource_name)
);

CREATE INDEX IF NOT EXISTS idx_api_key_resource_access_key ON api_key_resource_access(api_key_id, resource_type);
