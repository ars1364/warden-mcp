package api

import (
	"errors"
	"net/http"

	"github.com/ars1364/warden-mcp/internal/authn"
	"github.com/ars1364/warden-mcp/internal/store"
)

type setupRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (s *Server) handleSetup(w http.ResponseWriter, r *http.Request) {
	var req setupRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "malformed request body")
		return
	}
	if req.Username == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "username and password are required")
		return
	}

	count, err := s.db.CountUsers()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL", "failed to check existing users")
		return
	}
	if count > 0 {
		writeError(w, http.StatusForbidden, "ALREADY_INITIALIZED", "an admin user already exists")
		return
	}

	hash, err := authn.HashPassword(req.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL", "failed to hash password")
		return
	}

	user, err := s.db.CreateUser(req.Username, hash)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL", "failed to create user")
		return
	}

	writeJSON(w, http.StatusCreated, user)
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	TOTPCode string `json:"totp_code"`
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "malformed request body")
		return
	}

	user, err := s.db.GetUserByUsername(req.Username)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "invalid username or password")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL", "failed to look up user")
		return
	}

	if !authn.CheckPassword(user.PasswordHash, req.Password) {
		s.auditLoginFailed(r, user.ID, user.Username)
		writeError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "invalid username or password")
		return
	}

	if user.TOTPEnabled {
		if req.TOTPCode == "" {
			writeError(w, http.StatusBadRequest, "TOTP_REQUIRED", "totp code is required")
			return
		}
		if !authn.ValidateTOTP(user.TOTPSecret, req.TOTPCode) {
			s.auditLoginFailed(r, user.ID, user.Username)
			writeError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "invalid username or password")
			return
		}
	}

	token, err := s.jwt.Issue(user.ID, user.Username)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL", "failed to issue token")
		return
	}

	_ = s.db.WriteAudit(store.AuditEntry{
		ActorType:  "user",
		ActorID:    user.ID,
		ActorLabel: user.Username,
		Action:     "login",
		IP:         r.RemoteAddr,
	})

	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}

func (s *Server) auditLoginFailed(r *http.Request, userID, username string) {
	_ = s.db.WriteAudit(store.AuditEntry{
		ActorType:  "user",
		ActorID:    userID,
		ActorLabel: username,
		Action:     "login_failed",
		IP:         r.RemoteAddr,
	})
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	user, err := s.db.GetUserByID(userIDFromContext(r))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL", "failed to load user")
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (s *Server) handleTOTPSetup(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r)
	username := usernameFromContext(r)

	secret, otpauthURL, err := authn.GenerateTOTPSecret(username, "warden-mcp")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL", "failed to generate totp secret")
		return
	}

	if err := s.db.SetUserTOTP(userID, secret, false); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL", "failed to store totp secret")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"secret": secret, "otpauth_url": otpauthURL})
}

type totpEnableRequest struct {
	Code string `json:"code"`
}

func (s *Server) handleTOTPEnable(w http.ResponseWriter, r *http.Request) {
	var req totpEnableRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "malformed request body")
		return
	}

	userID := userIDFromContext(r)
	user, err := s.db.GetUserByID(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL", "failed to load user")
		return
	}

	if user.TOTPSecret == "" || !authn.ValidateTOTP(user.TOTPSecret, req.Code) {
		writeError(w, http.StatusBadRequest, "INVALID_TOTP_CODE", "invalid totp code")
		return
	}

	if err := s.db.SetUserTOTP(userID, user.TOTPSecret, true); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL", "failed to enable totp")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type totpDisableRequest struct {
	Password string `json:"password"`
}

type changePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

func (s *Server) handleChangePassword(w http.ResponseWriter, r *http.Request) {
	var req changePasswordRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "malformed request body")
		return
	}
	if len(req.NewPassword) < 8 {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "new password must be at least 8 characters")
		return
	}

	userID := userIDFromContext(r)
	user, err := s.db.GetUserByID(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL", "failed to load user")
		return
	}

	if !authn.CheckPassword(user.PasswordHash, req.CurrentPassword) {
		writeError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "invalid current password")
		return
	}

	hash, err := authn.HashPassword(req.NewPassword)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL", "failed to hash password")
		return
	}
	if err := s.db.SetUserPassword(userID, hash); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL", "failed to update password")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleTOTPDisable(w http.ResponseWriter, r *http.Request) {
	var req totpDisableRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "malformed request body")
		return
	}

	userID := userIDFromContext(r)
	user, err := s.db.GetUserByID(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL", "failed to load user")
		return
	}

	if !authn.CheckPassword(user.PasswordHash, req.Password) {
		writeError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "invalid password")
		return
	}

	if err := s.db.SetUserTOTP(userID, "", false); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL", "failed to disable totp")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
