package api

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/ars1364/warden-mcp/internal/store"
	"github.com/go-chi/chi/v5"
)

// hostAddressDTO is one labeled IP/hostname on the wire, e.g. {"label":"public","address":"5.6.7.8"}.
type hostAddressDTO struct {
	Label   string `json:"label"`
	Address string `json:"address"`
}

// hostRequest is shared by create and update. Parent/jump hosts are referenced by name (not
// the opaque id) since that's what a human or an agent actually has in hand.
type hostRequest struct {
	Name             string           `json:"name"`
	HostType         string           `json:"host_type"`
	Status           string           `json:"status"`
	Description      string           `json:"description"`
	Tags             []string         `json:"tags"`
	ParentHostName   string           `json:"parent_host_name"`
	LocationKind     string           `json:"location_kind"`
	CloudProvider    string           `json:"cloud_provider"`
	CloudAccount     string           `json:"cloud_account"`
	PhysicalLocation string           `json:"physical_location"`
	SSHPort          int              `json:"ssh_port"`
	SSHUsername      string           `json:"ssh_username"`
	SSHSecretName    string           `json:"ssh_secret_name"`
	SSHJumpHostName  string           `json:"ssh_jump_host_name"`
	Addresses        []hostAddressDTO `json:"addresses"`
}

// hostResponse enriches a store.Host with resolved parent/jump host names and, for the detail
// view, its addresses.
type hostResponse struct {
	store.Host
	ParentHostName  string           `json:"parent_host_name,omitempty"`
	SSHJumpHostName string           `json:"ssh_jump_host_name,omitempty"`
	Addresses       []hostAddressDTO `json:"addresses,omitempty"`
}

func (s *Server) resolveHostRefs(req hostRequest) (parentID, jumpID *string, err error) {
	if req.ParentHostName != "" {
		p, err := s.db.GetHostByName(req.ParentHostName)
		if err != nil {
			return nil, nil, fmt.Errorf("parent_host_name %q not found", req.ParentHostName)
		}
		parentID = &p.ID
	}
	if req.SSHJumpHostName != "" {
		j, err := s.db.GetHostByName(req.SSHJumpHostName)
		if err != nil {
			return nil, nil, fmt.Errorf("ssh_jump_host_name %q not found", req.SSHJumpHostName)
		}
		jumpID = &j.ID
	}
	return parentID, jumpID, nil
}

func toHostInput(req hostRequest, parentID, jumpID *string) store.HostInput {
	return store.HostInput{
		Name: req.Name, HostType: req.HostType, Status: req.Status,
		Description: req.Description, Tags: req.Tags, ParentHostID: parentID,
		LocationKind: req.LocationKind, CloudProvider: req.CloudProvider,
		CloudAccount: req.CloudAccount, PhysicalLocation: req.PhysicalLocation,
		SSHPort: req.SSHPort, SSHUsername: req.SSHUsername, SSHSecretName: req.SSHSecretName,
		SSHJumpHostID: jumpID,
	}
}

// hostName looks up a host's name by id, tolerating a missing/nil id — used to enrich responses.
func (s *Server) hostName(id *string) string {
	if id == nil {
		return ""
	}
	h, err := s.db.GetHostByID(*id)
	if err != nil {
		return ""
	}
	return h.Name
}

func (s *Server) toHostResponse(h *store.Host, withAddresses bool) hostResponse {
	resp := hostResponse{Host: *h, ParentHostName: s.hostName(h.ParentHostID), SSHJumpHostName: s.hostName(h.SSHJumpHostID)}
	if withAddresses {
		addrs, err := s.db.GetHostAddresses(h.ID)
		if err == nil {
			resp.Addresses = make([]hostAddressDTO, len(addrs))
			for i, a := range addrs {
				resp.Addresses[i] = hostAddressDTO{Label: a.Label, Address: a.Address}
			}
		}
	}
	return resp
}

func (s *Server) handleListHosts(w http.ResponseWriter, r *http.Request) {
	hosts, err := s.db.ListHosts()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL", "failed to list hosts")
		return
	}
	out := make([]hostResponse, len(hosts))
	for i, h := range hosts {
		// Unlike secret values, addresses aren't sensitive — the whole point of an IP
		// inventory is seeing them at a glance, so the list response includes them too.
		out[i] = s.toHostResponse(&h, true)
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleCreateHost(w http.ResponseWriter, r *http.Request) {
	var req hostRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "malformed request body")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "name is required")
		return
	}
	parentID, jumpID, err := s.resolveHostRefs(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", err.Error())
		return
	}

	host, err := s.db.CreateHost(toHostInput(req, parentID, jumpID), userIDFromContext(r))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL", "failed to create host")
		return
	}
	if err := s.db.ReplaceHostAddresses(host.ID, toStoreAddresses(req.Addresses)); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL", "failed to save addresses")
		return
	}

	writeJSON(w, http.StatusCreated, s.toHostResponse(host, true))
}

func (s *Server) handleGetHost(w http.ResponseWriter, r *http.Request) {
	host, err := s.db.GetHostByID(chi.URLParam(r, "id"))
	if err != nil {
		s.writeHostLookupError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, s.toHostResponse(host, true))
}

func (s *Server) handleUpdateHost(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := s.db.GetHostByID(id); err != nil {
		s.writeHostLookupError(w, err)
		return
	}

	var req hostRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "malformed request body")
		return
	}
	parentID, jumpID, err := s.resolveHostRefs(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", err.Error())
		return
	}

	if err := s.db.UpdateHost(id, toHostInput(req, parentID, jumpID)); err != nil {
		s.writeHostLookupError(w, err)
		return
	}
	if err := s.db.ReplaceHostAddresses(id, toStoreAddresses(req.Addresses)); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL", "failed to save addresses")
		return
	}

	host, err := s.db.GetHostByID(id)
	if err != nil {
		s.writeHostLookupError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, s.toHostResponse(host, true))
}

func (s *Server) handleDeleteHost(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := s.db.DeleteHost(id); err != nil {
		s.writeHostLookupError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) writeHostLookupError(w http.ResponseWriter, err error) {
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "host not found")
		return
	}
	writeError(w, http.StatusInternalServerError, "INTERNAL", "failed to access host")
}

func toStoreAddresses(in []hostAddressDTO) []store.HostAddress {
	out := make([]store.HostAddress, len(in))
	for i, a := range in {
		out[i] = store.HostAddress{Label: a.Label, Address: a.Address, Position: i}
	}
	return out
}
