package tui

import "testing"

func TestValidateAddress(t *testing.T) {
	cases := []struct {
		ticker, net, addr string
		want              bool
	}{
		// BTC — legacy, segwit-p2sh, native segwit
		{"BTC", "Mainnet", "1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2", true},
		{"BTC", "Mainnet", "bc1qar0srrr7xfkvy5l643lydnw9re59gtzzwf5mdq", true},
		{"BTC", "Mainnet", "not-a-btc-address", false},

		// XMR
		{"XMR", "Mainnet", "8As3ohpyc6Tfvbuow6cR5Y5cs44CEzRLhxJrqbPfgVMaHEqL3zeWmmNKxCpLPGuYwMnbbXVjUVvMzJrYbnGRJBpw2bfaecF", true},
		{"XMR", "Mainnet", "4short", false},

		// EVM
		{"USDT", "ERC20", "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb1", true},
		{"USDT", "BEP20", "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb1", true},
		{"ETH", "Mainnet", "742d35Cc6634C0532925a3b844Bc9e7595f0bEb1", false}, // missing 0x

		// Tron
		{"USDT", "TRC20", "TRX9aVxCdHJ8Y4Pbh3UJjNjXg3oAnkTu7w", true},
		{"USDT", "TRC20", "0xnotTron", false},

		// SOL
		{"SOL", "Mainnet", "DRiP2Pn2K6fuMLKQmt5rZWwGFoEbJfxVumKBfFRXKAFW", true},

		// LTC, DOGE, BCH, ZEC — basic prefix checks
		{"LTC", "Mainnet", "ltc1qar0srrr7xfkvy5l643lydnw9re59gtzz5xzmvr", true},
		{"DOGE", "Mainnet", "DRkbCTwbsjPCJpef4SST3FerNFGCVjm3sk", true},

		// Unknown chain → permissive
		{"BTRFLY", "MAINNET", "literally-anything-goes-here", true},

		// Empty
		{"BTC", "Mainnet", "", false},
	}
	for _, tc := range cases {
		ok, hint := validateAddress(tc.ticker, tc.net, tc.addr)
		if ok != tc.want {
			t.Errorf("validateAddress(%q,%q,%q) = (%v,%q), want ok=%v",
				tc.ticker, tc.net, tc.addr, ok, hint, tc.want)
		}
	}
}
