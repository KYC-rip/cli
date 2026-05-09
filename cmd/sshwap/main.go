package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kyc-rip/cli/internal/sshhost"
)

// Set by goreleaser via -ldflags="-X main.version=… -X main.commit=…".
var (
	version = "dev"
	commit  = "none"
)

func main() {
	addr := flag.String("addr", envOr("SSHWAP_ADDR", ":22"), "listen address")
	hostKey := flag.String("host-key", envOr("SSHWAP_HOST_KEY", ""), "path to ed25519 host key (auto-generated if missing)")
	apiBase := flag.String("api", envOr("SSHWAP_API_BASE", "https://api.kyc.rip"), "kyc.rip API base URL")
	apiKey := flag.String("api-key", envOr("SSHWAP_API_KEY", ""), "scoped API key (optional)")
	maxSessions := flag.Int("max-sessions", 200, "global concurrent session cap")
	maxPerIP := flag.Int("max-per-ip", 3, "concurrent sessions per source IP")
	idle := flag.Duration("idle", 90*time.Second, "session idle timeout")
	healthAddr := flag.String("health", envOr("SSHWAP_HEALTH_ADDR", "127.0.0.1:9090"), "loopback health/metrics listener (empty to disable)")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()
	if *showVersion {
		fmt.Printf("sshwap %s (%s)\n", version, commit)
		return
	}

	logger := log.New(os.Stderr, "[sshwap] ", log.LstdFlags|log.Lmicroseconds)

	srv, err := sshhost.New(sshhost.Config{
		Addr:        *addr,
		HostKeyPath: *hostKey,
		APIBase:     *apiBase,
		APIKey:      *apiKey,
		MaxSessions: *maxSessions,
		MaxPerIP:    *maxPerIP,
		IdleTimeout: *idle,
		// Banner intentionally omitted — it appears before the TUI and
		// looks like a stray pre-auth message in clients. The TUI itself
		// is the welcome surface.
	}, logger)
	if err != nil {
		logger.Fatalf("init: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Run the health/metrics endpoint in the background. Failure here
	// is non-fatal (the SSH listener is what we actually care about).
	if *healthAddr != "" {
		go func() {
			if err := srv.ServeHealth(ctx, *healthAddr); err != nil {
				logger.Printf("health: %v", err)
			}
		}()
	}

	if err := srv.ListenAndServe(ctx); err != nil {
		logger.Fatalf("serve: %v", err)
	}
}

func envOr(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
