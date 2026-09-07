package main

import (
	"log"
	"net/http"
	"os"

	"github.com/ars1364/warden-mcp/internal/api"
	"github.com/ars1364/warden-mcp/internal/authn"
	"github.com/ars1364/warden-mcp/internal/config"
	"github.com/ars1364/warden-mcp/internal/crypto"
	"github.com/ars1364/warden-mcp/internal/mcpserver"
	"github.com/ars1364/warden-mcp/internal/store"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	db, err := store.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("store: %v", err)
	}
	defer db.Close()

	box, err := crypto.NewBox(cfg.MasterKey)
	if err != nil {
		log.Fatalf("crypto: %v", err)
	}

	jwtIssuer := authn.NewJWTIssuer(cfg.JWTSecret)

	mux := http.NewServeMux()
	mux.Handle("/mcp", mcpserver.NewHandler(db, box))
	mux.Handle("/mcp/", mcpserver.NewHandler(db, box))
	mux.Handle("/", withStaticFallback(api.NewRouter(db, box, jwtIssuer, cfg.Env)))

	addr := "127.0.0.1:" + cfg.Port
	log.Printf("warden-mcp listening on %s (env=%s)", addr, cfg.Env)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server: %v", err)
	}
}

// withStaticFallback serves the built frontend (web/dist) for any path the API router doesn't
// claim, so the Go binary can serve the SPA in production without a separate web server.
func withStaticFallback(apiHandler http.Handler) http.Handler {
	const distDir = "./web/dist"
	fileServer := http.FileServer(http.Dir(distDir))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(r.URL.Path) >= 5 && r.URL.Path[:5] == "/api/" || r.URL.Path == "/healthz" {
			apiHandler.ServeHTTP(w, r)
			return
		}
		if _, err := os.Stat(distDir + r.URL.Path); err != nil {
			r.URL.Path = "/"
		}
		fileServer.ServeHTTP(w, r)
	})
}
