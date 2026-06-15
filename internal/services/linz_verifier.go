package services

import (
	"context"
	"fmt"

	"github.com/tpt-nz/nz-common/linz"
)

// PropertyVerificationResult holds LINZ title data fetched during property clause creation.
type PropertyVerificationResult struct {
	TitleNo          string
	LegalDescription string
	LINZLandAreaM2   float64
}

// PropertyVerifier verifies a NZ land title reference against the LINZ Data Service.
// Implementations may be swapped for testing or when LINZ_API_KEY is not configured.
type PropertyVerifier interface {
	VerifyTitle(ctx context.Context, titleRef string) (*PropertyVerificationResult, error)
}

// LINZPropertyVerifier uses the nz-common linz package to verify land titles.
// Requires a LINZ Data Service API key (https://data.linz.govt.nz/).
type LINZPropertyVerifier struct {
	client *linz.Client
}

func NewLINZPropertyVerifier(client *linz.Client) *LINZPropertyVerifier {
	return &LINZPropertyVerifier{client: client}
}

func (v *LINZPropertyVerifier) VerifyTitle(ctx context.Context, titleRef string) (*PropertyVerificationResult, error) {
	titles, err := v.client.SearchByTitleReference(ctx, titleRef)
	if err != nil {
		return nil, fmt.Errorf("linz verifier: %w", err)
	}
	if len(titles) == 0 {
		return nil, fmt.Errorf("linz verifier: title reference %q not found", titleRef)
	}
	t := titles[0]
	return &PropertyVerificationResult{
		TitleNo:          t.TitleNo,
		LegalDescription: t.LegalDescription,
		LINZLandAreaM2:   t.AreaSqm,
	}, nil
}
