package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/kyc-rip/cli/internal/api"
	"github.com/kyc-rip/cli/internal/tui"
	"github.com/kyc-rip/cli/internal/update"
)

// Set by goreleaser via -ldflags="-X main.version=… -X main.commit=…".
var (
	version = "dev"
	commit  = "none"
)

func main() {
	// Subcommand dispatch — `kyc-cli update [--check]`. Handled before
	// flag.Parse() so the global flags don't choke on the subcommand.
	if len(os.Args) > 1 && os.Args[1] == "update" {
		runUpdate(os.Args[2:])
		return
	}

	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "kyc-cli — terminal-only crypto swap (the kyc.rip aggregator as a TUI)")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Usage:")
		fmt.Fprintln(os.Stderr, "  kyc-cli [flags]              run the swap TUI")
		fmt.Fprintln(os.Stderr, "  kyc-cli update [--check]     self-update to the latest release")
		fmt.Fprintln(os.Stderr, "  kyc-cli --version            print version and exit")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Flags:")
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Source: https://github.com/kyc-rip/cli")
	}

	apiBase := flag.String("api", envOr("KYC_API_BASE", "https://api.kyc.rip"), "kyc.rip API base URL")
	apiKey := flag.String("api-key", envOr("KYC_API_KEY", ""), "scoped API key (optional)")
	timeout := flag.Duration("timeout", 12*time.Second, "API timeout per call")
	dryRun := flag.Bool("dry-run", false, "stop at the quote step — never call POST /create")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()
	if *showVersion {
		fmt.Printf("kyc-cli %s (%s)\n", version, commit)
		return
	}

	// Background nudge: print a single-line note to stderr if a newer
	// version is available. Throttled to once per 24h on disk.
	if version != "dev" {
		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
		_ = update.PromptIfNewer(ctx, version)
		cancel()
	}

	cli := api.New(api.WithBase(*apiBase), api.WithAPIKey(*apiKey), api.WithTimeout(*timeout))
	prog := tea.NewProgram(
		tui.New(tui.Config{Client: cli, DryRun: *dryRun}),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	if _, err := prog.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func runUpdate(args []string) {
	fs := flag.NewFlagSet("update", flag.ExitOnError)
	checkOnly := fs.Bool("check", false, "report whether an update is available, then exit")
	_ = fs.Parse(args)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if origin := update.PackageManagerOrigin(); origin != "" && !*checkOnly {
		fmt.Fprintf(os.Stderr, "kyc-cli was installed via %s — use that to upgrade.\n", origin)
		os.Exit(2)
	}

	tag, err := update.CheckLatest(ctx, version)
	if err != nil {
		fmt.Fprintln(os.Stderr, "update check failed:", err)
		os.Exit(1)
	}
	if tag == "" {
		fmt.Printf("kyc-cli %s — up to date.\n", version)
		return
	}
	if *checkOnly {
		fmt.Printf("kyc-cli %s available (you have %s) — run `kyc-cli update` to upgrade\n", tag, version)
		return
	}
	fmt.Printf("upgrading kyc-cli %s → %s …\n", version, tag)
	if err := update.DoUpdate(ctx, tag); err != nil {
		fmt.Fprintln(os.Stderr, "update failed:", err)
		os.Exit(1)
	}
}

func envOr(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
