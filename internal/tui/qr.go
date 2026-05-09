package tui

import (
	qrcodeTerminal "github.com/Baozisoftware/qrcode-terminal-go"
)

// renderQR uses Baozisoftware/qrcode-terminal-go — the library the user
// specifically reported renders correctly in their terminal stack.
//
// Default New() uses:
//   - front (dark module) = BG ANSI 256 index 0 (black)
//   - back (light module) = BG ANSI 256 index 7 (theme white / light gray)
//   - recovery level = Medium
//
// The library also strips 3 of skip2's 4-module quiet zone, leaving a
// 1-module quiet zone — tighter than other terminal QR libraries.
func renderQR(payload string) string {
	if payload == "" {
		return ""
	}
	s := qrcodeTerminal.New().Get(payload)
	if s == nil {
		return ""
	}
	return string(*s)
}
