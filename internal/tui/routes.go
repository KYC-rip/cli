package tui

import (
	"sort"
	"strings"

	"github.com/kyc-rip/cli/internal/api"
)

// routeMode identifies one of the four buckets (Suggested / Safe / Rate /
// Speed). Mirror of the Telegram bot so the same pick lands in the same
// bucket regardless of channel.
type routeMode string

const (
	modeSuggested routeMode = "suggested"
	modeSafe      routeMode = "safe"
	modeRate      routeMode = "rate"
	modeSpeed     routeMode = "speed"
)

var routeModesOrdered = []routeMode{modeSuggested, modeSafe, modeRate, modeSpeed}

func (m routeMode) Label() string {
	switch m {
	case modeSuggested:
		return "Suggested"
	case modeSafe:
		return "Safe"
	case modeRate:
		return "Rate"
	case modeSpeed:
		return "Speed"
	}
	return string(m)
}

func (m routeMode) Glyph() string {
	switch m {
	case modeSuggested:
		return "★"
	case modeSafe:
		return "🛡"
	case modeRate:
		return "$"
	case modeSpeed:
		return "⚡"
	}
	return "·"
}

// routePicks holds one route per bucket. Pointers point into the input
// slice so the selected route can be passed directly to CreateReq.
type routePicks struct {
	Suggested *api.Route
	Safe      *api.Route
	Rate      *api.Route
	Speed     *api.Route
}

func (p routePicks) get(m routeMode) *api.Route {
	switch m {
	case modeSuggested:
		return p.Suggested
	case modeSafe:
		return p.Safe
	case modeRate:
		return p.Rate
	case modeSpeed:
		return p.Speed
	}
	return nil
}

func (p routePicks) modes() []routeMode {
	out := make([]routeMode, 0, 4)
	for _, m := range routeModesOrdered {
		if p.get(m) != nil {
			out = append(out, m)
		}
	}
	return out
}

// pickRoutesByMode mirrors bot/src/telegram/kyc.ts pickRoutesByMode():
//   - Suggested = routes[0] (V2's smart-ranked top)
//   - Safe      = strictest privacy pool (kyc=A + log=A) → kyc=A → any, by best amount_to
//   - Rate      = highest amount_to overall
//   - Speed     = lowest ETA
//
// Plus dedupe-by-provider and a greedy "no provider in two buckets" slot-fill
// so the four picks are visibly distinct.
func pickRoutesByMode(routes []api.Route) routePicks {
	if len(routes) == 0 {
		return routePicks{}
	}
	// Dedupe by provider, keeping the best amount_to per provider.
	byProvider := map[string]*api.Route{}
	order := []string{}
	for i := range routes {
		r := &routes[i]
		if cur, ok := byProvider[r.Provider]; !ok {
			byProvider[r.Provider] = r
			order = append(order, r.Provider)
		} else if r.AmountTo > cur.AmountTo {
			byProvider[r.Provider] = r
		}
	}
	deduped := make([]*api.Route, 0, len(byProvider))
	for _, p := range order {
		deduped = append(deduped, byProvider[p])
	}

	privacy := filterRoutes(deduped, func(r *api.Route) bool {
		return strings.EqualFold(r.KYC, "A") &&
			(strings.EqualFold(r.LogPolicy, "A") || strings.EqualFold(r.LogPolicy, "B"))
	})
	strict := filterRoutes(deduped, func(r *api.Route) bool {
		return strings.EqualFold(r.KYC, "A") && strings.EqualFold(r.LogPolicy, "A")
	})

	used := map[string]bool{}
	pickFirst := func(list []*api.Route) *api.Route {
		for _, r := range list {
			if !used[r.Provider] {
				used[r.Provider] = true
				return r
			}
		}
		return nil
	}

	suggested := pickFirst(deduped) // preserves V2 order

	safePool := strict
	if len(safePool) == 0 {
		safePool = privacy
	}
	if len(safePool) == 0 {
		safePool = deduped
	}
	safe := pickFirst(sortByAmountDesc(safePool))

	rate := pickFirst(sortByAmountDesc(deduped))
	speed := pickFirst(sortByETAAsc(deduped))

	return routePicks{Suggested: suggested, Safe: safe, Rate: rate, Speed: speed}
}

func filterRoutes(in []*api.Route, fn func(*api.Route) bool) []*api.Route {
	out := []*api.Route{}
	for _, r := range in {
		if fn(r) {
			out = append(out, r)
		}
	}
	return out
}

func sortByAmountDesc(in []*api.Route) []*api.Route {
	out := append([]*api.Route(nil), in...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].AmountTo > out[j].AmountTo })
	return out
}

func sortByETAAsc(in []*api.Route) []*api.Route {
	out := append([]*api.Route(nil), in...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].ETA < out[j].ETA })
	return out
}

// cycleMode advances selection within the available bucket list. step=+1
// moves forward, step=-1 moves back. Wraps around. Returns the input
// unchanged if the list is empty.
func cycleMode(modes []routeMode, cur routeMode, step int) routeMode {
	if len(modes) == 0 {
		return cur
	}
	idx := -1
	for i, m := range modes {
		if m == cur {
			idx = i
			break
		}
	}
	if idx < 0 {
		return modes[0]
	}
	idx = (idx + step + len(modes)) % len(modes)
	return modes[idx]
}
