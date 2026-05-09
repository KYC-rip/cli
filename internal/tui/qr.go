package tui

import (
	"bytes"

	"github.com/mdp/qrterminal/v3"
	"rsc.io/qr"
)

// renderQR delegates to mdp/qrterminal — the same library tailscale,
// 1password-cli, and a pile of others use for terminal QR rendering.
// Default `Generate` emits full-block "  " cells with explicit ANSI 16
// background colors (`\x1b[40m` / `\x1b[47m`), which is the most
// universally compatible encoding.
func renderQR(payload string) string {
	if payload == "" {
		return ""
	}
	var buf bytes.Buffer
	qrterminal.Generate(payload, qr.M, &buf)
	return buf.String()
}
