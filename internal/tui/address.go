package tui

import (
	"regexp"
	"strings"
)

// validateAddress is a *cheap* heuristic for common chains so we can
// reject obvious typos before sending the order to the upstream API.
// We do NOT try to verify checksums — false positives are worse than
// false negatives here. Anything we can't recognise returns ok=true
// (don't block flows for chains we don't know).
//
// The check runs on the picker-resolved ticker (not the network
// suffix), so e.g. USDT-TRC20 → ticker=USDT, picks Tron-style rules.
func validateAddress(toTicker, toNet, addr string) (ok bool, hint string) {
	t := strings.ToUpper(toTicker)
	n := strings.ToUpper(toNet)
	a := strings.TrimSpace(addr)
	if a == "" {
		return false, "address is required"
	}

	switch {
	case t == "BTC":
		if reBTC.MatchString(a) {
			return true, ""
		}
		return false, "BTC address: starts with 1, 3, or bc1"

	case t == "XMR":
		if reXMR.MatchString(a) {
			return true, ""
		}
		return false, "XMR address: starts with 4 or 8 (95+ chars)"

	case t == "ETH" || strings.HasPrefix(n, "ERC20") || n == "ETHEREUM" ||
		n == "BEP20" || n == "POLYGON" || n == "ARBITRUM" || n == "OPTIMISM" || n == "BASE":
		if reEVM.MatchString(a) {
			return true, ""
		}
		return false, "EVM address: 0x followed by 40 hex chars"

	case t == "SOL" || strings.HasPrefix(n, "SOL"):
		if reSOL.MatchString(a) {
			return true, ""
		}
		return false, "SOL address: 32-44 base58 chars"

	case t == "TRX" || n == "TRC20":
		if reTRON.MatchString(a) {
			return true, ""
		}
		return false, "Tron address: starts with T (34 chars)"

	case t == "LTC":
		if reLTC.MatchString(a) {
			return true, ""
		}
		return false, "LTC address: starts with L, M, or ltc1"

	case t == "DOGE":
		if reDOGE.MatchString(a) {
			return true, ""
		}
		return false, "DOGE address: starts with D"

	case t == "BCH":
		if reBCH.MatchString(a) {
			return true, ""
		}
		return false, "BCH address: bitcoincash:q… or 1/3 prefix"

	case t == "ZEC":
		if reZEC.MatchString(a) {
			return true, ""
		}
		return false, "ZEC address: starts with t1, t3, or zs"
	}
	return true, ""
}

var (
	reBTC  = regexp.MustCompile(`^(?:bc1[a-z0-9]{6,87}|[13][a-km-zA-HJ-NP-Z1-9]{25,34})$`)
	reXMR  = regexp.MustCompile(`^[48][a-zA-Z0-9]{94,105}$`)
	reEVM  = regexp.MustCompile(`^0x[a-fA-F0-9]{40}$`)
	reSOL  = regexp.MustCompile(`^[1-9A-HJ-NP-Za-km-z]{32,44}$`)
	reTRON = regexp.MustCompile(`^T[A-Za-z0-9]{33}$`)
	reLTC  = regexp.MustCompile(`^(?:ltc1[a-z0-9]{6,87}|[LM3][a-km-zA-HJ-NP-Z1-9]{25,34})$`)
	reDOGE = regexp.MustCompile(`^D[a-km-zA-HJ-NP-Z1-9]{25,34}$`)
	reBCH  = regexp.MustCompile(`^(?:bitcoincash:[qp][a-z0-9]{41}|[13][a-km-zA-HJ-NP-Z1-9]{25,34})$`)
	reZEC  = regexp.MustCompile(`^(?:t[13][a-km-zA-HJ-NP-Z1-9]{33}|zs1[a-z0-9]{75})$`)
)
