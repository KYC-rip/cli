package tui

import (
	"encoding/base64"

	qrcodeTerminal "github.com/Baozisoftware/qrcode-terminal-go"
	"github.com/skip2/go-qrcode"
)

// renderQR returns the text-mode QR (Baozisoftware library, ANSI-256 BG
// paint with 1-module quiet zone). This is what shows by default.
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

// renderQRImage returns an iTerm2 inline-image OSC sequence carrying a PNG
// QR. Supported natively by Warp, iTerm2, WezTerm, mintty, Tabby, and a
// few others. Terminals that don't recognize OSC 1337 will print the
// raw escape as garbage — so this is opt-in via the `g` key in the TUI,
// not the default.
func renderQRImage(payload string) string {
	if payload == "" {
		return ""
	}
	qr, err := qrcode.New(payload, qrcode.Medium)
	if err != nil {
		return ""
	}
	pngBytes, err := qr.PNG(256)
	if err != nil {
		return ""
	}
	b64 := base64.StdEncoding.EncodeToString(pngBytes)
	// BEL terminator (\a) is widely supported. ESC \ also works on some
	// terminals; BEL is the older / more compatible form for OSC 1337.
	return "\x1b]1337;File=name=qr.png;inline=1;preserveAspectRatio=1:" + b64 + "\a"
}
