package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/tpt-nz/tpt-will-estate-nz/internal/models"
	"github.com/tpt-nz/tpt-will-estate-nz/internal/services"
	"github.com/tpt-nz/realme-go"
)

type WillHandler struct {
	willSvc *services.WillService
	log     *slog.Logger
}

func NewWillHandler(willSvc *services.WillService, log *slog.Logger) *WillHandler {
	return &WillHandler{willSvc: willSvc, log: log.With("handler", "will")}
}

// Create creates a draft will.
func (h *WillHandler) Create(w http.ResponseWriter, r *http.Request) {
	identity := realme.IdentityFromContext(r.Context())
	if identity == nil {
		http.Error(w, "unauthenticated", http.StatusUnauthorized)
		return
	}

	var req services.CreateWillRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	claims := models.IdentityClaims{
		RealMeFLT:       identity.FLT,
		FullName:       identity.FullName,
		AssuranceLevel: identity.AssuranceLevel.String(),
		IsVerified:     identity.IsVerified(),
	}

	id, err := h.willSvc.CreateDraft(r.Context(), claims, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	respondJSON(w, http.StatusCreated, map[string]any{"id": id})
}

func respondJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

// Router placeholder for future endpoints.
func (h *WillHandler) Routes() chi.Router { return chi.NewRouter() }

