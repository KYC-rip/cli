package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/kyc-rip/sshwap/internal/api"
	"github.com/kyc-rip/sshwap/internal/tui"
)

// Set by goreleaser via -ldflags="-X main.version=… -X main.commit=…".
var (
	version = "dev"
	commit  = "none"
)

func main() {
	apiBase := flag.String("api", envOr("KYC_API_BASE", "https://api.kyc.rip"), "kyc.rip API base URL")
	apiKey := flag.String("api-key", envOr("KYC_API_KEY", ""), "scoped API key (optional)")
	timeout := flag.Duration("timeout", 12*time.Second, "API timeout per call")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()
	if *showVersion {
		fmt.Printf("kyc-cli %s (%s)\n", version, commit)
		return
	}

	cli := api.New(api.WithBase(*apiBase), api.WithAPIKey(*apiKey), api.WithTimeout(*timeout))
	prog := tea.NewProgram(
		tui.New(tui.Config{Client: cli}),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	if _, err := prog.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func envOr(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
