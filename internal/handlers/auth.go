package handlers

import (
	"log/slog"
	"net/http"

	"github.com/tpt-nz/realme-go"
)

type AuthHandler struct {
	provider *realme.Provider
	logger   *slog.Logger
}

func NewAuthHandler(provider *realme.Provider, logger *slog.Logger) *AuthHandler {
	return &AuthHandler{provider: provider, logger: logger.With("handler", "auth")}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	h.provider.LoginHandler()(w, r)
}
func (h *AuthHandler) Callback(w http.ResponseWriter, r *http.Request) {
	h.provider.CallbackHandler(nil)(w, r)
}
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	h.provider.LogoutHandler()(w, r)
}
func (h *AuthHandler) Metadata(w http.ResponseWriter, r *http.Request) {
	h.provider.MetadataHandler()(w, r)
}
func (h *AuthHandler) Status(w http.ResponseWriter, r *http.Request) {
	identity := realme.IdentityFromContext(r.Context())
	if identity == nil {
		respondJSON(w, http.StatusOK, map[string]any{"authenticated": false})
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{
		"authenticated": true,
		"identity": map[string]any{
			"flt":        identity.FLT,
			"fullName":   identity.FullName,
			"assurance":  identity.AssuranceLevel.String(),
			"isVerified": identity.IsVerified(),
		},
	})
}

