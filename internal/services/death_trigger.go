package services

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/tpt-nz/tpt-will-estate-nz/internal/models"
)

// DeathTrigger handles death notifications.
// In the scaffold it is not wired to a real BDM event source.
type DeathTrigger struct {
	willSvc *WillService
	log     *slog.Logger
}

func NewDeathTrigger(willSvc *WillService, log *slog.Logger) *DeathTrigger {
	return &DeathTrigger{willSvc: willSvc, log: log}
}

// HandleBDMNotification is a stub handler.
func (d *DeathTrigger) HandleBDMNotification(_ http.ResponseWriter, _ *http.Request) {
	// TODO: parse event payload, map to WillID, then call UnlockAtDeath.
}

// UnlockByWillID is a helper for tests.
func (d *DeathTrigger) UnlockByWillID(ctx context.Context, id models.WillID) error {
	return d.willSvc.UnlockAtDeath(ctx, id, d.now())
}

func (d *DeathTrigger) now() (t models.WillStatus) { return }

