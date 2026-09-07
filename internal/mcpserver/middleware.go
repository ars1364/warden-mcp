package mcpserver

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/ars1364/warden-mcp/internal/authn"
	"github.com/ars1364/warden-mcp/internal/store"
)

type ctxKey int

const apiKeyCtxKey ctxKey = iota

func withAPIKey(ctx context.Context, key *store.APIKey) context.Context {
	return context.WithValue(ctx, apiKeyCtxKey, key)
}

func apiKeyFromContext(ctx context.Context) *store.APIKey {
	key, _ := ctx.Value(apiKeyCtxKey).(*store.APIKey)
	return key
}

// authMiddleware authenticates the MCP transport itself; per-tool scope checks happen in tools.go
// against the *store.APIKey captured in getServer's closure, not via ambient context lookups inside
// tool handlers, since a fresh mcp.Server is built per request in server.go.
func authMiddleware(db *store.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			if token == "" || !authn.IsAPIKeyFormat(token) {
				writeUnauthorized(w)
				return
			}
			key, err := db.FindActiveByHash(authn.HashAPIKey(token))
			if err != nil {
				writeUnauthorized(w)
				return
			}
			_ = db.TouchAPIKey(key.ID, clientIP(r))
			next.ServeHTTP(w, r.WithContext(withAPIKey(r.Context(), key)))
		})
	}
}

func writeUnauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": "UNAUTHORIZED", "message": "missing or invalid API key"},
	})
}

func clientIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		return strings.Split(fwd, ",")[0]
	}
	return r.RemoteAddr
}
