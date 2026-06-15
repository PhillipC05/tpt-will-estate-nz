package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/tpt-nz/tpt-will-estate-nz/internal/models"
	"github.com/tpt-nz/tpt-will-estate-nz/internal/repository"
	"github.com/tpt-nz/tpt-will-estate-nz/internal/services"
)

// LivenessHandler handles testator liveness check endpoints.
//
// In production, SendReminder would be called by a scheduled job (e.g. NATS
// JetStream scheduled consumer or an Atlas Cron trigger) that queries for wills
// where last_confirmed_at < now() - 1 year AND status IN ('draft','locked').
// The endpoint here allows the scheduler to trigger the check via HTTP.
type LivenessHandler struct {
	willSvc *services.WillService
	log     *slog.Logger
}

func NewLivenessHandler(willSvc *services.WillService, log *slog.Logger) *LivenessHandler {
	return &LivenessHandler{willSvc: willSvc, log: log.With("handler", "liveness")}
}

// SendReminder dispatches a liveness reminder to the testator of a will and
// records the timestamp. Email delivery is not yet implemented — wire in an
// email provider (SES, Postmark, Resend) here.
//
// This endpoint is intended to be called by a scheduled job, not by end users.
// In production, protect it with a bearer token or restrict to internal network.
func (h *LivenessHandler) SendReminder(w http.ResponseWriter, r *http.Request) {
	id := models.WillID(chi.URLParam(r, "id"))
	will, err := h.willSvc.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrWillNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		h.log.Error("liveness get will", "err", err, "will_id", string(id))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// Only active wills need liveness checks.
	switch will.Status {
	case models.WillStatusDraft, models.WillStatusLocked:
		// proceed
	default:
		http.Error(w, "will is not in an active state requiring a liveness check", http.StatusConflict)
		return
	}

	// Rate-limit: skip if a reminder was already sent within the last 24 hours.
	if will.LivenessNotifiedAt != nil && time.Since(*will.LivenessNotifiedAt) < 24*time.Hour {
		respondJSON(w, http.StatusOK, map[string]any{
			"skipped":           true,
			"lastReminderSentAt": will.LivenessNotifiedAt,
			"reason":            "reminder already sent within the last 24 hours",
		})
		return
	}

	if err := h.willSvc.RecordLivenessReminder(r.Context(), id); err != nil {
		h.log.Error("record liveness reminder", "err", err, "will_id", string(id))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// TODO: send email to will.Testator via your email provider.
	// Example with Resend SDK:
	//   resend.Send(ctx, resend.SendParams{
	//       To:      []string{testatorEmail},
	//       Subject: "Please confirm your will is still current",
	//       Html:    renderLivenessEmailTemplate(will),
	//   })
	h.log.Info("liveness reminder dispatched",
		"will_id", string(id),
		"testator_flt", will.Testator.RealMeFLT,
		"last_confirmed_at", will.LastConfirmedAt,
	)

	respondJSON(w, http.StatusAccepted, map[string]any{
		"dispatched": true,
		"note":       "email delivery not yet implemented — wire in an email provider",
	})
}

// BulkCheckLiveness is a convenience endpoint that returns all active wills
// that have not been confirmed within the past year. Call this from a scheduled
// job to build the list of wills needing reminders, then call SendReminder for each.
func (h *LivenessHandler) BulkCheckLiveness(w http.ResponseWriter, r *http.Request) {
	// Scaffold: returns a static message until a repo query is wired.
	// Production implementation: query WHERE status IN ('draft','locked')
	// AND (last_confirmed_at IS NULL OR last_confirmed_at < now() - interval '1 year')
	respondJSON(w, http.StatusOK, map[string]any{
		"note": "bulk liveness check not yet implemented — add a repository query filtering on last_confirmed_at",
	})
}
