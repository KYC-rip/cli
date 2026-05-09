package sshhost

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"
)

// HealthSnapshot is the JSON shape returned by the /healthz and
// /metrics endpoints. Stable enough to point monitoring at.
type HealthSnapshot struct {
	OK           bool   `json:"ok"`
	Version      string `json:"version,omitempty"`
	UptimeSec    int64  `json:"uptime_sec"`
	Active       int64  `json:"active_sessions"`
	TotalConns   int64  `json:"total_conns"`
	RejectedPty  int64  `json:"rejected_pty_or_exec"`
	RejectedCaps int64  `json:"rejected_caps"`
	Fingerprint  string `json:"host_key_fingerprint,omitempty"`
}

// Snapshot returns the current health state. Safe to call concurrently.
func (s *Server) Snapshot() HealthSnapshot {
	return HealthSnapshot{
		OK:           true,
		Active:       atomic.LoadInt64(&s.active),
		TotalConns:   atomic.LoadInt64(&s.totalConns),
		RejectedPty:  atomic.LoadInt64(&s.rejectedPty),
		RejectedCaps: atomic.LoadInt64(&s.rejectedCaps),
		Fingerprint:  s.Fingerprint(),
	}
}

// ServeHealth runs an HTTP listener with /healthz and /metrics.
// Intended for loopback only (e.g. 127.0.0.1:9090). Returns when ctx
// is cancelled or ListenAndServe fails. Logs but does not panic.
func (s *Server) ServeHealth(ctx context.Context, addr string) error {
	if addr == "" {
		return nil
	}
	started := time.Now()
	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		snap := s.Snapshot()
		snap.UptimeSec = int64(time.Since(started).Seconds())
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(snap)
	})

	// Prometheus-compatible plaintext metrics.
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		snap := s.Snapshot()
		uptime := int64(time.Since(started).Seconds())
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		fmt.Fprintf(w, "# HELP sshwap_uptime_seconds Process uptime in seconds.\n")
		fmt.Fprintf(w, "# TYPE sshwap_uptime_seconds counter\n")
		fmt.Fprintf(w, "sshwap_uptime_seconds %d\n", uptime)
		fmt.Fprintf(w, "# TYPE sshwap_active_sessions gauge\n")
		fmt.Fprintf(w, "sshwap_active_sessions %d\n", snap.Active)
		fmt.Fprintf(w, "# TYPE sshwap_total_connections counter\n")
		fmt.Fprintf(w, "sshwap_total_connections %d\n", snap.TotalConns)
		fmt.Fprintf(w, "# TYPE sshwap_rejected_pty counter\n")
		fmt.Fprintf(w, "sshwap_rejected_pty %d\n", snap.RejectedPty)
		fmt.Fprintf(w, "# TYPE sshwap_rejected_caps counter\n")
		fmt.Fprintf(w, "sshwap_rejected_caps %d\n", snap.RejectedCaps)
	})

	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 3 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      5 * time.Second,
	}
	go func() {
		<-ctx.Done()
		_ = srv.Close()
	}()
	s.logger.Printf("health endpoint listening on %s", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}
