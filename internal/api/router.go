package api

import (
	"net/http"

	"github.com/ars1364/warden-mcp/internal/authn"
	"github.com/ars1364/warden-mcp/internal/crypto"
	"github.com/ars1364/warden-mcp/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	db  *store.DB
	box *crypto.Box
	jwt *authn.JWTIssuer
	env string
}

func NewRouter(db *store.DB, box *crypto.Box, jwt *authn.JWTIssuer, env string) http.Handler {
	s := &Server{db: db, box: box, jwt: jwt, env: env}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(s.cors)

	r.Get("/healthz", s.handleHealthz)

	r.Route("/api", func(api chi.Router) {
		api.Post("/setup", s.handleSetup)
		api.Post("/login", s.handleLogin)

		api.Group(func(pr chi.Router) {
			pr.Use(s.requireAuth)

			pr.Get("/me", s.handleMe)
			pr.Post("/me/password", s.handleChangePassword)

			pr.Post("/totp/setup", s.handleTOTPSetup)
			pr.Post("/totp/enable", s.handleTOTPEnable)
			pr.Post("/totp/disable", s.handleTOTPDisable)

			pr.Get("/secrets", s.handleListSecrets)
			pr.Post("/secrets", s.handleCreateSecret)
			pr.Get("/secrets/{id}", s.handleGetSecret)
			pr.Put("/secrets/{id}", s.handleUpdateSecret)
			pr.Delete("/secrets/{id}", s.handleDeleteSecret)

			pr.Get("/apikeys", s.handleListAPIKeys)
			pr.Post("/apikeys", s.handleCreateAPIKey)
			pr.Delete("/apikeys/{id}", s.handleRevokeAPIKey)

			pr.Get("/audit", s.handleListAudit)
		})
	})

	return r
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
