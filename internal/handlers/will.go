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

// WillHandler handles all will lifecycle endpoints.
type WillHandler struct {
	willSvc  *services.WillService
	auditSvc *services.AuditService
	log      *slog.Logger
}

func NewWillHandler(willSvc *services.WillService, auditSvc *services.AuditService, log *slog.Logger) *WillHandler {
	return &WillHandler{willSvc: willSvc, auditSvc: auditSvc, log: log.With("handler", "will")}
}

// ── Core lifecycle ────────────────────────────────────────────────────────────

// Create creates a new draft will. Existing active wills for the same testator
// are automatically superseded (Wills Act 2007 s.17).
func (h *WillHandler) Create(w http.ResponseWriter, r *http.Request) {
	identity := h.requireVerified(w, r)
	if identity == nil {
		return
	}
	var req services.CreateWillRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	id, err := h.willSvc.CreateDraft(r.Context(), identityClaims(identity), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	respondJSON(w, http.StatusCreated, map[string]any{"id": id})
}

// GetByID returns will metadata and vault ciphertext.
func (h *WillHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := willIDParam(r)
	will, err := h.willSvc.GetByID(r.Context(), id)
	if err != nil {
		h.handleServiceErr(w, err)
		return
	}
	respondJSON(w, http.StatusOK, will)
}

// SignTestator records the testator's digital signature over the will payload hash.
func (h *WillHandler) SignTestator(w http.ResponseWriter, r *http.Request) {
	identity := h.requireVerified(w, r)
	if identity == nil {
		return
	}
	var req struct {
		PayloadHash string `json:"payloadHash"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.PayloadHash == "" {
		http.Error(w, "payloadHash required", http.StatusBadRequest)
		return
	}
	if err := h.willSvc.SignTestator(r.Context(), willIDParam(r), identityClaims(identity), req.PayloadHash); err != nil {
		h.handleServiceErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// SignWitness records a witness's digital signature over the will payload hash.
func (h *WillHandler) SignWitness(w http.ResponseWriter, r *http.Request) {
	identity := h.requireVerified(w, r)
	if identity == nil {
		return
	}
	var req struct {
		PayloadHash string `json:"payloadHash"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.PayloadHash == "" {
		http.Error(w, "payloadHash required", http.StatusBadRequest)
		return
	}
	if err := h.willSvc.SignWitness(r.Context(), willIDParam(r), identityClaims(identity), req.PayloadHash); err != nil {
		h.handleServiceErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Lock finalises the will after two-witness signing is complete.
func (h *WillHandler) Lock(w http.ResponseWriter, r *http.Request) {
	if err := h.willSvc.Lock(r.Context(), willIDParam(r), time.Now().UTC()); err != nil {
		h.handleServiceErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ── Feature 4: Codicil support ───────────────────────────────────────────────

// CreateCodicil creates an amendment will attached to an existing locked will.
// The parent will is NOT superseded; both remain active together.
func (h *WillHandler) CreateCodicil(w http.ResponseWriter, r *http.Request) {
	identity := h.requireVerified(w, r)
	if identity == nil {
		return
	}
	var req services.CreateWillRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	id, err := h.willSvc.CreateCodicil(r.Context(), willIDParam(r), identityClaims(identity), req)
	if err != nil {
		h.handleServiceErr(w, err)
		return
	}
	respondJSON(w, http.StatusCreated, map[string]any{"id": id})
}

// ── Feature 2: Testator liveness check ───────────────────────────────────────

// ConfirmLiveness records that the testator has confirmed the will is still
// current. Resets the liveness reminder clock.
func (h *WillHandler) ConfirmLiveness(w http.ResponseWriter, r *http.Request) {
	identity := h.requireVerified(w, r)
	if identity == nil {
		return
	}
	if err := h.willSvc.ConfirmLiveness(r.Context(), willIDParam(r), identityClaims(identity)); err != nil {
		h.handleServiceErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ── Feature 7: Digital asset clauses ─────────────────────────────────────────

// AddDigitalAssetClause adds a digital-asset bequest to a draft will.
func (h *WillHandler) AddDigitalAssetClause(w http.ResponseWriter, r *http.Request) {
	identity := h.requireVerified(w, r)
	if identity == nil {
		return
	}
	var clause models.DigitalAssetClause
	if err := json.NewDecoder(r.Body).Decode(&clause); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if clause.AssetType == "" || clause.Platform == "" || clause.Instruction == "" {
		http.Error(w, "assetType, platform, and instruction are required", http.StatusBadRequest)
		return
	}
	created, err := h.willSvc.AddDigitalAssetClause(r.Context(), willIDParam(r), identityClaims(identity), clause)
	if err != nil {
		h.handleServiceErr(w, err)
		return
	}
	respondJSON(w, http.StatusCreated, created)
}

// ── Feature 8: LINZ property cross-reference ──────────────────────────────────

// AddPropertyClause adds a real-property bequest, optionally verified with LINZ.
func (h *WillHandler) AddPropertyClause(w http.ResponseWriter, r *http.Request) {
	identity := h.requireVerified(w, r)
	if identity == nil {
		return
	}
	var req struct {
		TitleReference string `json:"titleReference"`
		BeneficiaryID  string `json:"beneficiaryId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.TitleReference == "" {
		http.Error(w, "titleReference required", http.StatusBadRequest)
		return
	}
	clause, err := h.willSvc.AddPropertyClause(r.Context(), willIDParam(r), identityClaims(identity), req.TitleReference, req.BeneficiaryID)
	if err != nil {
		h.handleServiceErr(w, err)
		return
	}
	respondJSON(w, http.StatusCreated, clause)
}

// ── Feature 6: Audit trail ────────────────────────────────────────────────────

// GetAuditLog returns the hash-chained audit log for a will.
// The chain integrity can be verified client-side by re-deriving each entry ID.
func (h *WillHandler) GetAuditLog(w http.ResponseWriter, r *http.Request) {
	id := willIDParam(r)
	entries, err := h.auditSvc.GetByWillID(r.Context(), id)
	if err != nil {
		h.log.Error("get audit log", "err", err, "will_id", string(id))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{
		"willId":  string(id),
		"entries": entries,
	})
}

// ── Shared helpers ────────────────────────────────────────────────────────────

func (h *WillHandler) requireVerified(w http.ResponseWriter, r *http.Request) *realme.Identity {
	identity := realme.IdentityFromContext(r.Context())
	if identity == nil {
		http.Error(w, "unauthenticated", http.StatusUnauthorized)
	}
	return identity
}

func willIDParam(r *http.Request) models.WillID {
	return models.WillID(chi.URLParam(r, "id"))
}

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
		http.Error(w, "invalid state for this operation", http.StatusConflict)
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
