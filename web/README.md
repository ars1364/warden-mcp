# Warden MCP — Web UI

React + TypeScript + Vite frontend for the warden-mcp credentials vault, talking to the Go
backend REST API.

## Dev setup

- Package manager: npm (Node v20.20.2, npm 10.8.2)
- Dev server: `npm run dev` → **http://localhost:5173**
- `vite.config.ts` proxies `/api` and `/healthz` to `http://localhost:7070` in dev, so the
  frontend always calls relative paths (`/api/...`) and works unchanged once the Go binary
  serves `web/dist` in production (same origin).

## Build

```bash
npm install
npx tsc -b       # type-check, zero errors
npm run build    # emits web/dist
```

## Auth

JWT is kept in memory and mirrored to `localStorage` under `warden_token`
(`src/lib/api.ts`). Any `401` response clears the token and the app falls back to the
`/login` route via `AuthContext`.

## Note: one endpoint not in the original API contract

The spec for this UI did not include a "change password" endpoint. `ChangePasswordForm`
(`src/pages/settings/ChangePasswordForm.tsx`) calls:

```
POST /api/me/password  { current_password, new_password }  → 204
```

If the backend implements this differently (different path/shape), update that one call
site — nothing else depends on it.

## Structure

See `src/` — pages under `src/pages/**`, shared UI in `src/components/common`, layout shell
in `src/components/layout`, API client in `src/lib/api.ts`, types in `src/types/api.ts`.
