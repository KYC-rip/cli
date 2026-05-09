package tui

import (
	"strings"

	"github.com/skip2/go-qrcode"
)

// renderQR returns a unicode-block QR code rendered at half-cell density.
// Each row of two QR pixels becomes one terminal row using ▀ glyphs,
// halving vertical space vs. naive full-block rendering.
func renderQR(payload string) string {
	if payload == "" {
		return ""
	}
	qr, err := qrcode.New(payload, qrcode.Medium)
	if err != nil {
		return ""
	}
	bm := qr.Bitmap()
	if len(bm) == 0 {
		return ""
	}
	var b strings.Builder
	for y := 0; y < len(bm); y += 2 {
		for x := 0; x < len(bm[y]); x++ {
			top := bm[y][x]
			bot := false
			if y+1 < len(bm) {
				bot = bm[y+1][x]
			}
			switch {
			case top && bot:
				b.WriteString("█")
			case top && !bot:
				b.WriteString("▀")
			case !top && bot:
				b.WriteString("▄")
			default:
				b.WriteString(" ")
			}
		}
		b.WriteByte('\n')
	}
	return b.String()
}
