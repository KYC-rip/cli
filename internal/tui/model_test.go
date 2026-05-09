package tui

import "testing"

func TestSplitTickerNet(t *testing.T) {
	cases := []struct {
		in           string
		ticker, net  string
	}{
		{"btc", "BTC", ""},
		{"USDT-TRC20", "USDT", "TRC20"},
		{"usdt/erc20", "USDT", "ERC20"},
		{"eth:base", "ETH", "BASE"},
		{"  BTC  ", "BTC", ""},
	}
	for _, tc := range cases {
		gt, gn := splitTickerNet(tc.in)
		if gt != tc.ticker || gn != tc.net {
			t.Errorf("splitTickerNet(%q) = (%q,%q), want (%q,%q)", tc.in, gt, gn, tc.ticker, tc.net)
		}
	}
}

func TestIsTerminal(t *testing.T) {
	cases := map[string]bool{
		"WAITING":   false,
		"finished":  true,
		"FINISHED":  true,
		"FAILED":    true,
		"EXPIRED":   true,
		"REFUNDED":  true,
		"EXCHANGING": false,
	}
	for k, want := range cases {
		if got := isTerminal(k); got != want {
			t.Errorf("isTerminal(%q) = %v, want %v", k, got, want)
		}
	}
}
