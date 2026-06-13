// Package bdm provides a stub client for the New Zealand Births, Deaths and
// Marriages (BDM) registry API.
//
// BDM does not currently expose a public API for death notifications.
// This package defines the interface and a stub implementation so that the
// Digital Will & Estate app (app-will-estate) can be built and tested
// without a live BDM integration. When a BDM API becomes available,
// replace the stub with a real implementation.
//
// In the meantime, death can be notified by:
//   - An authorised executor presenting a death certificate scan
//   - A registered funeral director via an authenticated webhook
package bdm

import (
	"context"
	"fmt"
	"time"
)

// DeathRecord is a death registration record from BDM.
type DeathRecord struct {
	RegistrationID string
	FullName       string
	DateOfDeath    time.Time
	AgeAtDeath     int
	PlaceOfDeath   string
	RegistrationDate time.Time
}

// Notifier is the interface for receiving death notifications.
// In production this would be a webhook from BDM or a polling integration.
type Notifier interface {
	// GetDeathRecord looks up a death registration by the deceased person's
	// NZ birth registration number or similar government identifier.
	GetDeathRecord(ctx context.Context, identifier string) (*DeathRecord, error)

	// NotifyDeath is called by an external webhook when a death is registered.
	// This should trigger the estate unlock flow in app-will-estate.
	NotifyDeath(ctx context.Context, record DeathRecord) error
}

// StubNotifier is a no-op implementation for development and testing.
// It always returns ErrNotImplemented.
type StubNotifier struct{}

// ErrNotImplemented is returned by the StubNotifier.
var ErrNotImplemented = fmt.Errorf("bdm: real BDM API not yet available; use manual death notification flow")

func (s *StubNotifier) GetDeathRecord(_ context.Context, _ string) (*DeathRecord, error) {
	return nil, ErrNotImplemented
}

func (s *StubNotifier) NotifyDeath(_ context.Context, _ DeathRecord) error {
	return ErrNotImplemented
}

// ManualNotifier accepts death notifications from authorised executors
// who have submitted a certified death certificate scan.
// This is the fallback pathway until a real BDM API is available.
type ManualNotification struct {
	DeceasedFLT        string    // RealMe FLT of the deceased
	DeceasedFullName   string    // As on the death certificate
	DateOfDeath        time.Time
	CertificateScanURL string    // Secure reference to uploaded scan
	NotifiedByFLT      string    // RealMe FLT of the notifying executor
	NotifiedAt         time.Time
}
