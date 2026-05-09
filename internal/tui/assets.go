package tui

// topAssets is the static reference list shown next to the From/To
// inputs. Picked so a user landing in the form has a workable surface
// without typing blind, and without a fuzzy-picker (Codex review
// "cheaper path"). Tickers and networks here mirror what /v2/exchange
// accepts directly.
var topAssets = []string{
	"BTC",
	"ETH",
	"XMR",
	"USDT-TRC20",
	"USDT-ERC20",
	"USDT-BEP20",
	"USDC-ERC20",
	"USDC-Base",
	"USDC-SOL",
	"SOL",
	"BNB",
	"TRX",
	"LTC",
	"BCH",
	"DOGE",
	"DASH",
	"ZEC",
	"ADA",
	"DOT",
	"XRP",
}

// assetHint returns a single-line suggestion strip rendered under
// each asset input. Width-aware: trims to fit so it never wraps.
func assetHint(width int) string {
	if width <= 0 {
		width = 80
	}
	out := "popular: "
	for i, a := range topAssets {
		piece := a
		if i+1 < len(topAssets) {
			piece += "  "
		}
		// +len("popular: ") + current
		if len(out)+len(piece) > width-2 {
			break
		}
		out += piece
	}
	return out
}
