package linz

import (
	"context"
	"fmt"
	"net/url"
)

// LandTitle holds the result of a LINZ land title lookup.
// Owner information is not available via the public LINZ Data Service WFS API;
// that data requires access to LINZ e-Dealing (secure system).
type LandTitle struct {
	TitleNo          string
	LegalDescription string
	AreaSqm          float64
}

// LandTitleResult holds WFS query results for the NZ Parcels layer (51564).
type LandTitleResult struct {
	Features []struct {
		Properties PropertyBoundary `json:"properties"`
	} `json:"features"`
}

// SearchByTitleReference looks up land parcel data by NZ land title reference
// (e.g. "NA1234/1") using the LINZ Data Service NZ Parcels dataset (layer-51564).
//
// The title_no field in the parcels dataset stores the primary title reference.
// If multiple parcels share the same title (compound titles), all are returned.
//
// Note: availability of specific parcels depends on your LINZ LDS subscription.
func (c *Client) SearchByTitleReference(ctx context.Context, titleRef string) ([]LandTitle, error) {
	if titleRef == "" {
		return nil, fmt.Errorf("linz: title reference cannot be empty")
	}

	params := url.Values{}
	params.Set("service", "WFS")
	params.Set("version", "2.0.0")
	params.Set("request", "GetFeature")
	params.Set("typeNames", "layer-51564") // NZ Parcels
	params.Set("outputFormat", "application/json")
	params.Set("count", "10")
	params.Set("cql_filter", fmt.Sprintf("title_no='%s'", titleRef))

	path := "/services;key=" + c.apiKey + "/wfs?" + params.Encode()

	var result LandTitleResult
	if err := c.wfsGet(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("linz: title lookup %q: %w", titleRef, err)
	}

	out := make([]LandTitle, 0, len(result.Features))
	for _, f := range result.Features {
		p := f.Properties
		out = append(out, LandTitle{
			TitleNo:          p.TitleNo,
			LegalDescription: p.LegalDesc,
			AreaSqm:          p.AreaSqm,
		})
	}
	return out, nil
}
