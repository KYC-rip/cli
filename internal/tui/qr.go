package tui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/skip2/go-qrcode"
)

// QR scanners want dark-on-light. Force black FG / white BG so polarity
// is correct regardless of terminal scheme.
var qrStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("0")).
	Background(lipgloss.Color("15"))

// renderQR returns a QR code as ██ (two terminal cells per filled module)
// on a white background. Two cells wide × one row tall gives roughly
// square modules in any terminal (cells are ~2:1 tall:wide).
//
// We deliberately avoid half-block ▀/▄ rendering. It's more compact but
// renders unreliably in fonts/terminals where line-spacing exceeds the
// glyph height — the half-blocks then float with visible vertical gaps
// that turn modules into horizontal stripes and break scanning. Reported
// in the wild on macOS Terminal.app.
//
// skip2's ToString already includes a 4-module quiet zone.
func renderQR(payload string) string {
	if payload == "" {
		return ""
	}
	qr, err := qrcode.New(payload, qrcode.Medium)
	if err != nil {
		return ""
	}
	return qrStyle.Render(qr.ToString(false))
}
