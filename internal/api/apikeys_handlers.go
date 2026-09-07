package api

import (
	"errors"
	"net/http"

	"github.com/ars1364/warden-mcp/internal/authn"
	"github.com/ars1364/warden-mcp/internal/store"
	"github.com/go-chi/chi/v5"
)

func (s *Server) handleListAPIKeys(w http.ResponseWriter, r *http.Request) {
	keys, err := s.db.ListAPIKeys()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL", "failed to list api keys")
		return
	}
	writeJSON(w, http.StatusOK, keys)
}

type createAPIKeyRequest struct {
	Name   string   `json:"name"`
	Scopes []string `json:"scopes"`
}

var validAPIKeyScopes = map[string]bool{"read": true, "write": true}

func (s *Server) handleCreateAPIKey(w http.ResponseWriter, r *http.Request) {
	var req createAPIKeyRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "malformed request body")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "name is required")
		return
	}

	if len(req.Scopes) == 0 {
		req.Scopes = []string{"read"}
	}
	for _, scope := range req.Scopes {
		if !validAPIKeyScopes[scope] {
			writeError(w, http.StatusBadRequest, "INVALID_BODY", "invalid scope: "+scope)
			return
		}
	}

	plain, hash, err := authn.GenerateAPIKey()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL", "failed to generate api key")
		return
	}

	userID := userIDFromContext(r)
	key, err := s.db.CreateAPIKey(req.Name, req.Scopes, hash, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL", "failed to store api key")
		return
	}

	_ = s.db.WriteAudit(store.AuditEntry{
		ActorType:  "user",
		ActorID:    userID,
		ActorLabel: usernameFromContext(r),
		Action:     "create",
		Detail:     "api_key:" + req.Name,
		IP:         r.RemoteAddr,
	})

	writeJSON(w, http.StatusCreated, struct {
		store.APIKey
		Key string `json:"key"`
	}{APIKey: *key, Key: plain})
}

func (s *Server) handleRevokeAPIKey(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID := userIDFromContext(r)

	if err := s.db.RevokeAPIKey(id, userID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "api key not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL", "failed to revoke api key")
		return
	}

	_ = s.db.WriteAudit(store.AuditEntry{
		ActorType:  "user",
		ActorID:    userID,
		ActorLabel: usernameFromContext(r),
		Action:     "revoke",
		Detail:     "api_key:" + id,
		IP:         r.RemoteAddr,
	})

	w.WriteHeader(http.StatusNoContent)
}
