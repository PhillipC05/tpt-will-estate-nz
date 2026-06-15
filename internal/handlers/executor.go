package handlers

import (
	"errors"
	"log/slog"
	"net/http"

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

// GetWill returns a will's metadata and vault ciphertext to a nominated executor.
// The will must be in the unlocked_at_death state (vault has been unlocked after
// a verified BDM death notification).
//
// Future: cross-check the executor's RealMe FLT against the nominated executor list
// once executor FLTs are stored at will-creation time.
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

	if will.Status != models.WillStatusUnlockedDead {
		http.Error(w, "will vault is not yet unlocked", http.StatusForbidden)
		return
	}

	respondJSON(w, http.StatusOK, will)
}
