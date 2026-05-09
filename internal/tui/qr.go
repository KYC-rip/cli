package tui

import (
	"strings"

	"github.com/skip2/go-qrcode"
)

// QR rendered by painting terminal-cell backgrounds — no glyphs at all.
// Each module is 2 cells wide × 1 row tall (≈ square in 2:1 cells).
//
// Why not unicode blocks (▀ / ▄ / █):
//   - Half-blocks (▀ / ▄): unreliable when the terminal's line-spacing
//     exceeds glyph height. Adjacent rows float with visible vertical gaps
//     and modules collapse into stripes (macOS Terminal.app).
//   - Full blocks (██): some monospaced fonts kern adjacent glyphs with
//     hairline gaps, breaking the solid-fill illusion (also seen on
//     macOS Terminal with SF Mono).
//
// Spaces with explicit BG colors render solidly in every terminal that
// supports ANSI 256 background colors (which is everything bubbletea
// already targets).
const (
	qrBlackBG = "\x1b[48;5;0m"  // ANSI 256 index 0 — pure black
	qrWhiteBG = "\x1b[48;5;15m" // ANSI 256 index 15 — pure white
	qrReset   = "\x1b[0m"
)

// renderQR returns a black-on-white QR. skip2's Bitmap already includes a
// 4-module quiet zone, so we don't add extra padding.
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
	width := len(bm[0])
	var b strings.Builder
	for y := 0; y < len(bm); y++ {
		// Emit a BG-set sequence only when polarity changes — keeps the
		// escape-code overhead linear in the number of run boundaries
		// rather than O(width).
		last := -1
		for x := 0; x < width; x++ {
			cur := 1 // light/empty
			if bm[y][x] {
				cur = 0 // dark/filled
			}
			if cur != last {
				if cur == 0 {
					b.WriteString(qrBlackBG)
				} else {
					b.WriteString(qrWhiteBG)
				}
				last = cur
			}
			b.WriteString("  ")
		}
		b.WriteString(qrReset)
		if y < len(bm)-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}
