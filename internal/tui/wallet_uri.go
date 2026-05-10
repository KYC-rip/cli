package tui

import (
	"fmt"
	"strconv"
	"strings"
)

// walletURI returns a BIP-21-style payment URI for the from-side currency
// of a swap, suitable for OSC-8 hyperlinks that launch the user's local
// wallet on click. Returns "" for tokens whose URI scheme requires the
// EIP-681 contract-call form — too brittle to assemble without the token
// contract address per chain.
func walletURI(ticker, network, addr string, amount float64) string {
	t := strings.ToUpper(strings.TrimSpace(ticker))
	n := strings.ToUpper(strings.TrimSpace(network))
	if addr == "" {
		return ""
	}
	amt := ""
	if amount > 0 {
		amt = strconv.FormatFloat(amount, 'f', -1, 64)
	}

	q := func(k, v string) string {
		if v == "" {
			return ""
		}
		return "?" + k + "=" + v
	}

	switch {
	case t == "BTC":
		return "bitcoin:" + addr + q("amount", amt)
	case t == "LTC":
		return "litecoin:" + addr + q("amount", amt)
	case t == "DOGE":
		return "dogecoin:" + addr + q("amount", amt)
	case t == "BCH":
		return "bitcoincash:" + addr + q("amount", amt)
	case t == "XMR":
		return "monero:" + addr + q("tx_amount", amt)
	case t == "ZEC":
		return "zcash:" + addr + q("amount", amt)
	case t == "ETH":
		// ETH native: value is in wei. Convert from human ether.
		if amount > 0 {
			wei := amount * 1e18
			return fmt.Sprintf("ethereum:%s?value=%.0f", addr, wei)
		}
		return "ethereum:" + addr
	case t == "TRX" && n != "TRC20":
		// Native TRX only. We deliberately do NOT emit `tron:…?amount=N`
		// for TRC20 tokens (USDT-TRC20, USDC-TRC20) — the URI scheme is
		// interpreted as a NATIVE TRX transfer, so a wallet opening it
		// would try to send N TRX to the deposit address rather than N
		// tokens. Token transfers need TRC20-specific URI formats with
		// the contract address; out of scope here.
		return "tron:" + addr + q("amount", amt)
	case t == "SOL":
		return "solana:" + addr + q("amount", amt)
	}
	// EVM tokens (USDT-ERC20, USDC-ERC20, BEP20, Polygon, etc.) need
	// EIP-681 token-transfer form including the contract address. Skip.
	return ""
}

// osc8 wraps text in an OSC-8 hyperlink escape so terminals that support
// it (Warp, iTerm2, kitty, modern Terminal.app, Termius, VS Code) render
// the text as clickable. Non-supporting terminals strip the OSC and show
// just the text — no harm.
func osc8(uri, text string) string {
	if uri == "" {
		return text
	}
	return "\x1b]8;;" + uri + "\x1b\\" + text + "\x1b]8;;\x1b\\"
}
