package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/tpt-nz/tpt-will-estate-nz/internal/models"
	"github.com/tpt-nz/tpt-will-estate-nz/internal/repository"
	"github.com/tpt-nz/tpt-will-estate-nz/internal/services"
)

// BeneficiaryHandler handles beneficiary notification workflow endpoints.
type BeneficiaryHandler struct {
	willSvc *services.WillService
	log     *slog.Logger
}

func NewBeneficiaryHandler(willSvc *services.WillService, log *slog.Logger) *BeneficiaryHandler {
	return &BeneficiaryHandler{willSvc: willSvc, log: log.With("handler", "beneficiary")}
}

// NotifyAll triggers email notifications to all beneficiaries named in a will.
// The will must be in the unlocked_at_death state. Typically called by the
// executor after they have confirmed the will contents.
//
// Email delivery is not yet implemented — this endpoint logs the intent and
// returns 202 Accepted with a count. Wire in an email/queue provider here.
func (h *BeneficiaryHandler) NotifyAll(w http.ResponseWriter, r *http.Request) {
	id := models.WillID(chi.URLParam(r, "id"))
	will, err := h.willSvc.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrWillNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		h.log.Error("beneficiary notify", "err", err, "will_id", string(id))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if will.Status != models.WillStatusUnlockedDead {
		http.Error(w, "will vault is not yet unlocked", http.StatusForbidden)
		return
	}

	h.log.Info("beneficiary notification queued",
		"will_id", string(id),
		"count", len(will.Beneficiaries),
	)

	respondJSON(w, http.StatusAccepted, map[string]any{
		"queued": len(will.Beneficiaries),
		"note":   "email delivery not yet implemented — integrate an email provider here",
	})
}
