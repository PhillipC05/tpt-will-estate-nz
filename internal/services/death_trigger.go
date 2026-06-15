package services

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/tpt-nz/tpt-will-estate-nz/internal/models"
)

// DeathTrigger processes BDM death notifications and initiates the vault-unlock flow.
type DeathTrigger struct {
	willSvc *WillService
	log     *slog.Logger
	secret  []byte
}

func NewDeathTrigger(willSvc *WillService, log *slog.Logger) *DeathTrigger {
	return &DeathTrigger{
		willSvc: willSvc,
		log:     log,
		secret:  []byte(os.Getenv("BDM_WEBHOOK_SECRET")),
	}
}

// bdmDeathPayload is the expected JSON body of a BDM death notification webhook.
type bdmDeathPayload struct {
	WillID    string `json:"willId"`
	DelayDays int    `json:"delayDays"` // optional cooling-off period; 0 = immediate unlock
}

// HandleBDMNotification validates an HMAC-SHA256 signed BDM death notification
// and initiates the time-lock vault unlock flow.
func (d *DeathTrigger) HandleBDMNotification(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		http.Error(w, "read error", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if len(d.secret) > 0 && !d.validSignature(body, r.Header.Get("X-BDM-Signature")) {
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}

	var payload bdmDeathPayload
	if err := json.Unmarshal(body, &payload); err != nil || payload.WillID == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if err := d.NotifyDeath(r.Context(), models.WillID(payload.WillID), payload.DelayDays); err != nil {
		d.log.Error("notify death", "err", err, "will_id", payload.WillID)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// NotifyDeath applies the time-lock vault logic for a given will.
// Used directly in tests to bypass the HTTP layer.
func (d *DeathTrigger) NotifyDeath(ctx context.Context, id models.WillID, delayDays int) error {
	return d.willSvc.NotifyDeath(ctx, id, delayDays, time.Now().UTC())
}

func (d *DeathTrigger) validSignature(body []byte, sigHeader string) bool {
	mac := hmac.New(sha256.New, d.secret)
	mac.Write(body)
	expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(sigHeader))
}
