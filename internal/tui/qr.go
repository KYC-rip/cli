package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/skip2/go-qrcode"
)

// QR scanners want dark modules on a light background. Force black FG on
// white BG so polarity is correct regardless of the user's terminal color
// scheme — without this, dark-themed terminals render the QR as white-on-
// black, which many phone-camera scanners reject. Half-block rendering keeps
// modules close to square in physical pixels and halves vertical space vs.
// full-block.
var qrStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("0")).
	Background(lipgloss.Color("15"))

// renderQR returns a unicode-block QR code rendered at half-cell density
// with a 2-row top/bottom and 2-col left/right quiet zone added by qrStyle's
// padding (skip2's Bitmap omits the quiet zone).
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
	// 2-module horizontal quiet zone on each side, blank for spacing.
	const hQuiet = 2
	pad := strings.Repeat(" ", hQuiet)
	var b strings.Builder
	// Top quiet zone (one terminal row = 2 module rows).
	b.WriteString(strings.Repeat(" ", width+2*hQuiet))
	b.WriteByte('\n')
	for y := 0; y < len(bm); y += 2 {
		b.WriteString(pad)
		for x := 0; x < width; x++ {
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
		b.WriteString(pad)
		b.WriteByte('\n')
	}
	// Bottom quiet zone.
	b.WriteString(strings.Repeat(" ", width+2*hQuiet))
	return qrStyle.Render(b.String())
}
