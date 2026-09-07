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

func (s *Server) handleCreateSecret(w http.ResponseWriter, r *http.Request) {
	var req secretRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "malformed request body")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "name is required")
		return
	}
	secretType := normalizeSecretType(req.Type)
	if err := validateSecretFields(secretType, req.Fields); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", err.Error())
		return
	}

	userID := userIDFromContext(r)
	var secret *store.Secret
	var err error

	if secretType == store.TypeOpaque {
		var nonce, ciphertext []byte
		nonce, ciphertext, err = s.box.Seal(req.Value)
		if err == nil {
			secret, err = s.db.CreateSecret(req.Name, req.Description, req.Tags, nonce, ciphertext, userID)
		}
	} else {
		secret, err = s.db.CreateSecretMeta(req.Name, req.Description, req.Tags, secretType, userID)
		if err == nil {
			err = s.sealAndReplaceFields(secret.ID, req.Fields)
		}
	}
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

	resp, err := s.readSecretDetail(meta)
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

	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleUpdateSecret(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	meta, err := s.db.GetSecretMeta(id)
	if err != nil {
		s.writeSecretLookupError(w, err)
		return
	}

	var req secretRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "malformed request body")
		return
	}
	if err := validateSecretFields(meta.Type, req.Fields); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", err.Error())
		return
	}

	if meta.Type == store.TypeOpaque {
		nonce, ciphertext, sealErr := s.box.Seal(req.Value)
		if sealErr != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL", "failed to encrypt value")
			return
		}
		err = s.db.UpdateSecret(id, req.Description, req.Tags, nonce, ciphertext)
	} else {
		err = s.db.UpdateSecretMeta(id, req.Description, req.Tags)
		if err == nil {
			err = s.sealAndReplaceFields(id, req.Fields)
		}
	}
	if err != nil {
		s.writeSecretLookupError(w, err)
		return
	}

	meta, err = s.db.GetSecretMeta(id)
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
