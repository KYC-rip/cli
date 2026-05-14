package tui

import (
	"os/exec"
	"runtime"

	"github.com/atotto/clipboard"
)

func writeLocalClipboard(text string) error {
	return clipboard.WriteAll(text)
}

func openLocalBrowser(rawURL string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", rawURL)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", rawURL)
	default:
		cmd = exec.Command("xdg-open", rawURL)
	}
	return cmd.Start()
}
