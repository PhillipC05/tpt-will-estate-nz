package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/tpt-nz/tpt-will-estate-nz/internal/models"
	"github.com/tpt-nz/tpt-will-estate-nz/internal/repository"
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

// Create creates a draft will. The testator must hold a RealMe Verified Identity.
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

	claims := identityClaims(identity)
	id, err := h.willSvc.CreateDraft(r.Context(), claims, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	respondJSON(w, http.StatusCreated, map[string]any{"id": id})
}

// GetByID returns the will record for the authenticated testator or executor.
func (h *WillHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := models.WillID(chi.URLParam(r, "id"))
	will, err := h.willSvc.GetByID(r.Context(), id)
	if err != nil {
		h.handleServiceErr(w, err)
		return
	}
	respondJSON(w, http.StatusOK, will)
}

// SignTestator records the testator's digital signature over the will payload hash.
func (h *WillHandler) SignTestator(w http.ResponseWriter, r *http.Request) {
	identity := realme.IdentityFromContext(r.Context())
	if identity == nil {
		http.Error(w, "unauthenticated", http.StatusUnauthorized)
		return
	}
	id := models.WillID(chi.URLParam(r, "id"))

	var req struct {
		PayloadHash string `json:"payloadHash"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.PayloadHash == "" {
		http.Error(w, "payloadHash required", http.StatusBadRequest)
		return
	}

	if err := h.willSvc.SignTestator(r.Context(), id, identityClaims(identity), req.PayloadHash); err != nil {
		h.handleServiceErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// SignWitness records a witness's digital signature over the will payload hash.
func (h *WillHandler) SignWitness(w http.ResponseWriter, r *http.Request) {
	identity := realme.IdentityFromContext(r.Context())
	if identity == nil {
		http.Error(w, "unauthenticated", http.StatusUnauthorized)
		return
	}
	id := models.WillID(chi.URLParam(r, "id"))

	var req struct {
		PayloadHash string `json:"payloadHash"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.PayloadHash == "" {
		http.Error(w, "payloadHash required", http.StatusBadRequest)
		return
	}

	if err := h.willSvc.SignWitness(r.Context(), id, identityClaims(identity), req.PayloadHash); err != nil {
		h.handleServiceErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Lock finalises the will after two-witness signing is complete.
func (h *WillHandler) Lock(w http.ResponseWriter, r *http.Request) {
	id := models.WillID(chi.URLParam(r, "id"))
	if err := h.willSvc.Lock(r.Context(), id, time.Now().UTC()); err != nil {
		h.handleServiceErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// identityClaims maps a realme.Identity to the internal IdentityClaims model.
func identityClaims(id *realme.Identity) models.IdentityClaims {
	return models.IdentityClaims{
		RealMeFLT:      id.FLT,
		FullName:       id.FullName,
		AssuranceLevel: id.AssuranceLevel.String(),
		IsVerified:     id.IsVerified(),
	}
}

func (h *WillHandler) handleServiceErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, repository.ErrWillNotFound):
		http.Error(w, "not found", http.StatusNotFound)
	case errors.Is(err, services.ErrForbidden):
		http.Error(w, "forbidden", http.StatusForbidden)
	case errors.Is(err, services.ErrInvalidState):
		http.Error(w, "invalid state", http.StatusConflict)
	default:
		h.log.Error("service error", "err", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}

func respondJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
