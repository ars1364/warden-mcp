# warden-mcp

Self-hosted credential vault with a web UI and an MCP server front-end, so
Claude Code (or any MCP client) can fetch/store secrets on demand instead of
holding them in config files or env vars.

## Components

- `internal/store` — SQLite-backed storage (users, secrets, MCP API keys, audit log)
- `internal/crypto` — AES-256-GCM envelope encryption for secret values at rest
- `internal/authn` — bcrypt passwords, JWT sessions, TOTP 2FA, MCP API key hashing
- `internal/api` — REST API for the web UI (chi router)
- `internal/mcpserver` — MCP tools (`list_secrets`, `get_secret`, `set_secret`) over
  streamable HTTP, authenticated per-caller by a `wmcp_...` API key with read/write scopes
- `web/` — React + TypeScript + Vite + Tailwind frontend
- `cmd/server` — single binary serving the API, the MCP endpoint, and (in production)
  the built frontend as static files

## Running locally

```bash
export WARDEN_MASTER_KEY=$(openssl rand -base64 32)
export WARDEN_JWT_SECRET=$(openssl rand -base64 32)
export WARDEN_ENV=development
go run ./cmd/server
```

In another terminal:

```bash
cd web && npm install && npm run dev
```

The Vite dev server proxies `/api` and `/healthz` to `http://localhost:7070`.

## Production

Binds to `127.0.0.1:$WARDEN_PORT` only — put a reverse proxy (nginx) in front.
See `/etc/systemd/system/warden-mcp.service` and `/etc/nginx/sites-available/warden-mcp.conf`
on the deploy host. Config comes entirely from env vars (see `.env.example`).

## MCP client setup

Point an MCP client at `https://<your-domain>/mcp` with header
`Authorization: Bearer wmcp_<your key>` (create one from the web UI's API Keys page).
Tools: `list_secrets`, `get_secret(name)`, `set_secret(name, value, description?, tags?)`.
