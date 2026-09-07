// Package mcpserver exposes the credential vault to MCP clients (e.g. Claude Code)
// over streamable HTTP, authenticated by a per-caller wmcp_ API key (see internal/authn).
package mcpserver

import (
	"net/http"

	"github.com/ars1364/warden-mcp/internal/crypto"
	"github.com/ars1364/warden-mcp/internal/store"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// NewHandler returns an http.Handler to mount at /mcp. A fresh *mcp.Server is built per
// request so each tool closure captures the caller's own *store.APIKey for scope checks,
// rather than relying on context propagation into the SDK's tool-call machinery.
func NewHandler(db *store.DB, box *crypto.Box) http.Handler {
	getServer := func(r *http.Request) *mcp.Server {
		key := apiKeyFromContext(r.Context())
		if key == nil {
			return nil
		}

		server := mcp.NewServer(&mcp.Implementation{Name: "warden-mcp", Version: "0.1.0"}, nil)
		mcp.AddTool(server, &mcp.Tool{
			Name:        "list_secrets",
			Description: "List available secret names, descriptions and tags. Never returns values.",
		}, listSecretsHandler(db, key))
		mcp.AddTool(server, &mcp.Tool{
			Name:        "get_secret",
			Description: "Fetch the decrypted value of a secret by name. Requires the read scope.",
		}, getSecretHandler(db, box, key))
		mcp.AddTool(server, &mcp.Tool{
			Name:        "set_secret",
			Description: "Create or update a secret's value, description and tags. Requires the write scope.",
		}, setSecretHandler(db, box, key))
		return server
	}

	httpHandler := mcp.NewStreamableHTTPHandler(getServer, &mcp.StreamableHTTPOptions{Stateless: true})
	return authMiddleware(db)(httpHandler)
}
