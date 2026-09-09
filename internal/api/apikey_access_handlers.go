package api

import (
	"net/http"

	"github.com/ars1364/warden-mcp/internal/store"
	"github.com/go-chi/chi/v5"
)

// resourceGrantDTO mirrors store.ResourceGrant on the wire.
type resourceGrantDTO struct {
	ResourceType string `json:"resource_type"`
	ResourceName string `json:"resource_name"`
	CanRead      bool   `json:"can_read"`
	CanWrite     bool   `json:"can_write"`
}

type setResourceGrantsRequest struct {
	Grants []resourceGrantDTO `json:"grants"`
}

func (s *Server) handleGetAPIKeyAccess(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	grants, err := s.db.ListResourceGrants(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL", "failed to list access grants")
		return
	}
	out := make([]resourceGrantDTO, len(grants))
	for i, g := range grants {
		out[i] = resourceGrantDTO{ResourceType: g.ResourceType, ResourceName: g.ResourceName, CanRead: g.CanRead, CanWrite: g.CanWrite}
	}
	writeJSON(w, http.StatusOK, out)
}

// handleSetAPIKeyAccess replaces the FULL allowlist for a key across both resource types.
// Submitting zero grants of a given type clears that type's restriction entirely (the key
// becomes unrestricted for it again) — both types are always replaced, not just the ones
// present in the request, so "remove all host grants" actually removes them rather than
// leaving stale rows because the request happened not to mention hosts.
func (s *Server) handleSetAPIKeyAccess(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req setResourceGrantsRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "malformed request body")
		return
	}

	var secretGrants, hostGrants []store.ResourceGrant
	for _, g := range req.Grants {
		if g.ResourceName == "" {
			writeError(w, http.StatusBadRequest, "INVALID_BODY", "resource_name is required for every grant")
			return
		}
		switch g.ResourceType {
		case store.ResourceSecret:
			secretGrants = append(secretGrants, store.ResourceGrant{ResourceName: g.ResourceName, CanRead: g.CanRead, CanWrite: g.CanWrite})
		case store.ResourceHost:
			hostGrants = append(hostGrants, store.ResourceGrant{ResourceName: g.ResourceName, CanRead: g.CanRead, CanWrite: g.CanWrite})
		default:
			writeError(w, http.StatusBadRequest, "INVALID_BODY", "resource_type must be \"secret\" or \"host\"")
			return
		}
	}

	if err := s.db.ReplaceResourceGrants(id, store.ResourceSecret, secretGrants); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL", "failed to save secret grants")
		return
	}
	if err := s.db.ReplaceResourceGrants(id, store.ResourceHost, hostGrants); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL", "failed to save host grants")
		return
	}

	_ = s.db.WriteAudit(store.AuditEntry{
		ActorType:  "user",
		ActorID:    userIDFromContext(r),
		ActorLabel: usernameFromContext(r),
		Action:     "update",
		Detail:     "api_key_access:" + id,
		IP:         r.RemoteAddr,
	})

	s.handleGetAPIKeyAccess(w, r)
}
