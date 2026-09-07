package api

import (
	"errors"
	"net/http"

	"github.com/ars1364/warden-mcp/internal/store"
	"github.com/go-chi/chi/v5"
)

func (s *Server) handleListSecrets(w http.ResponseWriter, r *http.Request) {
	secrets, err := s.db.ListSecrets()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL", "failed to list secrets")
		return
	}
	writeJSON(w, http.StatusOK, secrets)
}

type createSecretRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	Value       string   `json:"value"`
}

func (s *Server) handleCreateSecret(w http.ResponseWriter, r *http.Request) {
	var req createSecretRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "malformed request body")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "name is required")
		return
	}

	nonce, ciphertext, err := s.box.Seal(req.Value)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL", "failed to encrypt value")
		return
	}

	userID := userIDFromContext(r)
	secret, err := s.db.CreateSecret(req.Name, req.Description, req.Tags, nonce, ciphertext, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL", "failed to create secret")
		return
	}

	_ = s.db.WriteAudit(store.AuditEntry{
		ActorType:  "user",
		ActorID:    userID,
		ActorLabel: usernameFromContext(r),
		Action:     "create",
		SecretName: secret.Name,
		IP:         r.RemoteAddr,
	})

	writeJSON(w, http.StatusCreated, secret)
}

func (s *Server) handleGetSecret(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	meta, err := s.db.GetSecretMeta(id)
	if err != nil {
		s.writeSecretLookupError(w, err)
		return
	}

	nonce, ciphertext, err := s.db.GetSecretValueByID(id)
	if err != nil {
		s.writeSecretLookupError(w, err)
		return
	}

	value, err := s.box.Open(nonce, ciphertext)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL", "failed to decrypt value")
		return
	}

	_ = s.db.WriteAudit(store.AuditEntry{
		ActorType:  "user",
		ActorID:    userIDFromContext(r),
		ActorLabel: usernameFromContext(r),
		Action:     "read",
		SecretName: meta.Name,
		IP:         r.RemoteAddr,
	})

	writeJSON(w, http.StatusOK, store.SecretWithValue{Secret: *meta, Value: value})
}

type updateSecretRequest struct {
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	Value       string   `json:"value"`
}

func (s *Server) handleUpdateSecret(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req updateSecretRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "malformed request body")
		return
	}

	nonce, ciphertext, err := s.box.Seal(req.Value)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL", "failed to encrypt value")
		return
	}

	if err := s.db.UpdateSecret(id, req.Description, req.Tags, nonce, ciphertext); err != nil {
		s.writeSecretLookupError(w, err)
		return
	}

	meta, err := s.db.GetSecretMeta(id)
	if err != nil {
		s.writeSecretLookupError(w, err)
		return
	}

	_ = s.db.WriteAudit(store.AuditEntry{
		ActorType:  "user",
		ActorID:    userIDFromContext(r),
		ActorLabel: usernameFromContext(r),
		Action:     "update",
		SecretName: meta.Name,
		IP:         r.RemoteAddr,
	})

	writeJSON(w, http.StatusOK, meta)
}

func (s *Server) handleDeleteSecret(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	meta, err := s.db.GetSecretMeta(id)
	if err != nil {
		s.writeSecretLookupError(w, err)
		return
	}

	if err := s.db.DeleteSecret(id); err != nil {
		s.writeSecretLookupError(w, err)
		return
	}

	_ = s.db.WriteAudit(store.AuditEntry{
		ActorType:  "user",
		ActorID:    userIDFromContext(r),
		ActorLabel: usernameFromContext(r),
		Action:     "delete",
		SecretName: meta.Name,
		IP:         r.RemoteAddr,
	})

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) writeSecretLookupError(w http.ResponseWriter, err error) {
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "secret not found")
		return
	}
	writeError(w, http.StatusInternalServerError, "INTERNAL", "failed to access secret")
}
