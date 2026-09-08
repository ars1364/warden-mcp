package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ars1364/warden-mcp/internal/authn"
	"github.com/ars1364/warden-mcp/internal/store"
	"github.com/go-chi/chi/v5"
)

// secretFieldInput/Output carry one key/value pair of a structured/totp/reference secret
// over the wire. Order in the JSON array is preserved as display position.
type secretFieldInput struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type secretFieldOutput struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// secretRequest is shared by create and update. Value is used for type=opaque only; Fields
// is used for structured/totp/reference only. The unused one is simply omitted by the caller.
// ExpiresAt is a plain "YYYY-MM-DD" date, or "" for no expiry.
type secretRequest struct {
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Tags        []string           `json:"tags"`
	Type        string             `json:"type"`
	Value       string             `json:"value"`
	Fields      []secretFieldInput `json:"fields"`
	ExpiresAt   string             `json:"expires_at"`
}

// parseExpiresAt turns the request's "YYYY-MM-DD" (or "") into the *time.Time the store
// layer wants, expressed as UTC midnight on that date.
func parseExpiresAt(s string) (*time.Time, error) {
	if s == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil, fmt.Errorf("expires_at must be a YYYY-MM-DD date")
	}
	return &t, nil
}

// secretDetailResponse is what GET /secrets/{id} returns: metadata plus the decrypted value
// (opaque) or fields (everything else).
type secretDetailResponse struct {
	store.Secret
	Value  string              `json:"value,omitempty"`
	Fields []secretFieldOutput `json:"fields,omitempty"`
}

func normalizeSecretType(t string) string {
	switch t {
	case store.TypeStructured, store.TypeTOTP, store.TypeReference:
		return t
	default:
		return store.TypeOpaque
	}
}

// validateSecretFields enforces the one type-specific invariant: a totp secret must carry a
// "seed" field, since get_totp_code depends on it existing.
func validateSecretFields(secretType string, fields []secretFieldInput) error {
	if secretType != store.TypeTOTP {
		return nil
	}
	for _, f := range fields {
		if f.Key == "seed" && f.Value != "" {
			return nil
		}
	}
	return fmt.Errorf("totp secrets require a non-empty %q field", "seed")
}

func (s *Server) sealAndReplaceFields(secretID string, fields []secretFieldInput) error {
	sealed := make([]store.FieldCiphertext, len(fields))
	for i, f := range fields {
		nonce, ciphertext, err := s.box.Seal(f.Value)
		if err != nil {
			return err
		}
		sealed[i] = store.FieldCiphertext{Key: f.Key, Position: i, Nonce: nonce, Ciphertext: ciphertext}
	}
	return s.db.ReplaceSecretFields(secretID, sealed)
}

func (s *Server) readSecretDetail(meta *store.Secret) (secretDetailResponse, error) {
	resp := secretDetailResponse{Secret: *meta}

	if meta.Type == store.TypeOpaque {
		nonce, ciphertext, err := s.db.GetSecretValueByID(meta.ID)
		if err != nil {
			return resp, err
		}
		value, err := s.box.Open(nonce, ciphertext)
		if err != nil {
			return resp, err
		}
		resp.Value = value
		return resp, nil
	}

	fields, err := s.db.GetSecretFields(meta.ID)
	if err != nil {
		return resp, err
	}
	resp.Fields = make([]secretFieldOutput, len(fields))
	for i, f := range fields {
		value, err := s.box.Open(f.Nonce, f.Ciphertext)
		if err != nil {
			return resp, err
		}
		resp.Fields[i] = secretFieldOutput{Key: f.Key, Value: value}
	}
	return resp, nil
}

// handleGetTOTPCode computes the current 6-digit code for a totp secret's seed without
// exposing the seed itself in the response.
func (s *Server) handleGetTOTPCode(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	meta, err := s.db.GetSecretMeta(id)
	if err != nil {
		s.writeSecretLookupError(w, err)
		return
	}
	if meta.Type != store.TypeTOTP {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "secret is not a totp credential")
		return
	}

	fields, err := s.db.GetSecretFields(meta.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL", "failed to load secret")
		return
	}
	var seed string
	for _, f := range fields {
		if f.Key == "seed" {
			seed, err = s.box.Open(f.Nonce, f.Ciphertext)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "INTERNAL", "failed to decrypt seed")
				return
			}
			break
		}
	}
	if seed == "" {
		writeError(w, http.StatusInternalServerError, "INTERNAL", "totp secret has no seed field")
		return
	}

	code, secondsRemaining, err := authn.CurrentTOTPCode(seed)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL", "failed to compute code")
		return
	}

	_ = s.db.WriteAudit(store.AuditEntry{
		ActorType:  "user",
		ActorID:    userIDFromContext(r),
		ActorLabel: usernameFromContext(r),
		Action:     "read",
		SecretName: meta.Name,
		IP:         r.RemoteAddr,
		Detail:     "totp_code",
	})

	writeJSON(w, http.StatusOK, map[string]any{"code": code, "seconds_remaining": secondsRemaining})
}
