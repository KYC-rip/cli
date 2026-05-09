package tui

import (
	"encoding/base64"
	"net/url"

	qrcodeTerminal "github.com/Baozisoftware/qrcode-terminal-go"
	"github.com/skip2/go-qrcode"
)

// qrDataURL builds a `data:image/png;base64,…` URL containing a PNG of
// the QR for `payload`. Wrapped in an OSC-8 hyperlink, clicking it in
// terminals that honor OSC-8 opens the QR as an inline image.
func qrDataURL(payload string) string {
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
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(pngBytes)
}

// osc52Clipboard returns the OSC 52 escape sequence that asks the user's
// terminal to write `text` into the system clipboard. Supported by
// Warp, iTerm2, Alacritty, kitty, mintty, WezTerm, Tabby, and others.
// Works across SSH because the escape travels in-band — the terminal
// emulator on the user's machine processes it.
//
// Terminator: ESC \ (ST) rather than BEL — the spec lists both but some
// terminals (notably tmux passthrough and a few wrappers) only accept ST.
func osc52Clipboard(text string) string {
	if text == "" {
		return ""
	}
	return "\x1b]52;c;" + base64.StdEncoding.EncodeToString([]byte(text)) + "\x1b\\"
}

// qrBrowserURL returns a short, copy-paste-friendly kyc.rip URL that
// renders the QR client-side. Intended as the always-works fallback for
// terminals where OSC-8 hyperlinks don't get clicked through (bubbletea
// mouse-mode interception, missing terminal support, etc.). Users select
// + copy + paste into a browser.
func qrBrowserURL(payload string) string {
	if payload == "" {
		return ""
	}
	return "https://api.kyc.rip/v2/qr?d=" + url.QueryEscape(payload)
}

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
