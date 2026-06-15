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
	"github.com/tpt-nz/realme-go"
)

// ExecutorHandler handles post-death executor access endpoints.
type ExecutorHandler struct {
	willSvc *services.WillService
	log     *slog.Logger
}

func NewExecutorHandler(willSvc *services.WillService, log *slog.Logger) *ExecutorHandler {
	return &ExecutorHandler{willSvc: willSvc, log: log.With("handler", "executor")}
}

// GetWill returns a will's full record — including the vault ciphertext — to a
// nominated executor once the vault is unlocked.
//
// Feature 1 — Executor FLT verification:
// The executor's RealMe FLT is checked against the FLTs stored on each nominated
// executor at will-creation time. A will where executors were created before this
// feature was deployed (FLT == "") grants access to any verified user; once FLTs
// are stored, only the exact executor may access the will.
//
// Feature 5 — Time-lock vault:
// If the will is in death_notified state and the cooling-off period has elapsed,
// the vault is transitioned to unlocked_at_death lazily in this request before
// serving the response.
func (h *ExecutorHandler) GetWill(w http.ResponseWriter, r *http.Request) {
	identity := realme.IdentityFromContext(r.Context())
	if identity == nil {
		http.Error(w, "unauthenticated", http.StatusUnauthorized)
		return
	}

	id := models.WillID(chi.URLParam(r, "id"))
	will, err := h.willSvc.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrWillNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		h.log.Error("executor get will", "err", err, "will_id", string(id))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// Feature 5: lazily complete the unlock if the cooling-off period has elapsed.
	if will.Status == models.WillStatusDeathNotified {
		now := time.Now().UTC()
		if will.UnlockAfter != nil && !now.Before(*will.UnlockAfter) {
			if err := h.willSvc.CompleteUnlock(r.Context(), id, now); err != nil {
				h.log.Error("complete unlock", "err", err, "will_id", string(id))
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}
			// Reload after status transition.
			will, err = h.willSvc.GetByID(r.Context(), id)
			if err != nil {
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}
		} else {
			// Still in cooling-off period — tell the client when to retry.
			respondJSON(w, http.StatusAccepted, map[string]any{
				"status":      "death_notified",
				"unlockAfter": will.UnlockAfter,
				"message":     "vault is in the mandatory cooling-off period",
			})
			return
		}
	}

	if will.Status != models.WillStatusUnlockedDead {
		http.Error(w, "will vault is not unlocked", http.StatusForbidden)
		return
	}

	// Feature 1: verify the caller is a nominated executor by RealMe FLT.
	if !isNominatedExecutor(will.Executors, identity.FLT) {
		http.Error(w, "you are not a nominated executor for this will", http.StatusForbidden)
		return
	}

	respondJSON(w, http.StatusOK, will)
}

// isNominatedExecutor returns true if flt matches any nominated executor.
// Wills created before executor FLT capture was introduced (FLT == "") allow
// any verified identity through to preserve backward compatibility.
func isNominatedExecutor(executors []models.Executor, flt string) bool {
	anyFLTStored := false
	for _, e := range executors {
		if e.FLT != "" {
			anyFLTStored = true
			if e.FLT == flt {
				return true
			}
		}
	}
	// Backward-compatible: if no executor FLTs were captured, allow any verified user.
	return !anyFLTStored
}
