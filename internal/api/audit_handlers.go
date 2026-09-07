package api

import (
	"net/http"
	"strconv"
)

func (s *Server) handleListAudit(w http.ResponseWriter, r *http.Request) {
	limit := 100
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}

	entries, err := s.db.ListAudit(limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL", "failed to list audit log")
		return
	}
	writeJSON(w, http.StatusOK, entries)
}
