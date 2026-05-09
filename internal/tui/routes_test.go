package tui

import (
	"testing"

	"github.com/kyc-rip/cli/internal/api"
)

func TestPickRoutesByMode_FourDistinctPicks(t *testing.T) {
	rs := []api.Route{
		// V2 top — direct engine, mid amount, mid ETA.
		{Provider: "lizex", AmountTo: 100, ETA: 12, KYC: "B", LogPolicy: "B"},
		// Best rate, but C-rated (no privacy filter).
		{Provider: "changenow", AmountTo: 110, ETA: 18, KYC: "C", LogPolicy: "C"},
		// Strictest privacy (kyc=A + log=A).
		{Provider: "exolix", AmountTo: 95, ETA: 15, KYC: "A", LogPolicy: "A"},
		// Fastest by far.
		{Provider: "trocador", AmountTo: 92, ETA: 5, KYC: "A", LogPolicy: "B"},
	}
	p := pickRoutesByMode(rs)
	if p.Suggested == nil || p.Suggested.Provider != "lizex" {
		t.Errorf("Suggested = %+v, want lizex", p.Suggested)
	}
	if p.Safe == nil || p.Safe.Provider != "exolix" {
		t.Errorf("Safe = %+v, want exolix (kyc=A+log=A)", p.Safe)
	}
	if p.Rate == nil || p.Rate.Provider != "changenow" {
		t.Errorf("Rate = %+v, want changenow (best amount_to)", p.Rate)
	}
	if p.Speed == nil || p.Speed.Provider != "trocador" {
		t.Errorf("Speed = %+v, want trocador (lowest ETA)", p.Speed)
	}
}

func TestPickRoutesByMode_DedupeByProvider(t *testing.T) {
	// Same provider with both fixed and floating quotes — only the better
	// amount survives dedupe.
	rs := []api.Route{
		{Provider: "lizex", AmountTo: 100, ETA: 12, Fixed: false},
		{Provider: "lizex", AmountTo: 102, ETA: 12, Fixed: true},
		{Provider: "exolix", AmountTo: 95, ETA: 15, KYC: "A", LogPolicy: "A"},
	}
	p := pickRoutesByMode(rs)
	if p.Suggested == nil || p.Suggested.AmountTo != 102 {
		t.Errorf("dedupe should keep amount_to=102, got %+v", p.Suggested)
	}
}

func TestPickRoutesByMode_NoDuplicateProviderAcrossBuckets(t *testing.T) {
	// Single dominant provider — Suggested takes it; Safe/Rate/Speed must
	// fall back to other providers (greedy slot-fill).
	rs := []api.Route{
		{Provider: "lizex", AmountTo: 100, ETA: 5, KYC: "A", LogPolicy: "A"},
		{Provider: "exolix", AmountTo: 95, ETA: 15, KYC: "A", LogPolicy: "A"},
		{Provider: "changenow", AmountTo: 90, ETA: 20, KYC: "C", LogPolicy: "C"},
	}
	p := pickRoutesByMode(rs)
	got := map[string]int{}
	for _, m := range routeModesOrdered {
		if r := p.get(m); r != nil {
			got[r.Provider]++
		}
	}
	for prov, n := range got {
		if n > 1 {
			t.Errorf("provider %s appears in %d buckets", prov, n)
		}
	}
}

func TestPickRoutesByMode_Empty(t *testing.T) {
	p := pickRoutesByMode(nil)
	if p.Suggested != nil || p.Safe != nil || p.Rate != nil || p.Speed != nil {
		t.Errorf("empty input must yield empty picks, got %+v", p)
	}
}
