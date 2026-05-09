package sshhost

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestHealthEndpoint(t *testing.T) {
	tmp := t.TempDir() + "/hostkey"
	logger := log.New(io.Discard, "", 0)
	srv, err := New(Config{
		Addr:        "127.0.0.1:0",
		HostKeyPath: tmp,
	}, logger)
	if err != nil {
		t.Fatal(err)
	}
	// Bump some counters so we can prove they surface in the JSON.
	atomic.AddInt64(&srv.totalConns, 5)
	atomic.AddInt64(&srv.rejectedPty, 2)
	atomic.AddInt64(&srv.rejectedCaps, 1)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	healthAddr := ln.Addr().String()
	_ = ln.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = srv.ServeHealth(ctx, healthAddr)
	}()

	// Poll until the server is up (max 2s)
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if c, err := net.DialTimeout("tcp", healthAddr, 100*time.Millisecond); err == nil {
			c.Close()
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	// /healthz JSON
	resp, err := http.Get("http://" + healthAddr + "/healthz")
	if err != nil {
		t.Fatalf("GET /healthz: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("status %d", resp.StatusCode)
	}
	var snap HealthSnapshot
	if err := json.NewDecoder(resp.Body).Decode(&snap); err != nil {
		t.Fatal(err)
	}
	if !snap.OK {
		t.Errorf("OK=false")
	}
	if snap.TotalConns != 5 || snap.RejectedPty != 2 || snap.RejectedCaps != 1 {
		t.Errorf("counters off: %+v", snap)
	}
	if snap.Fingerprint == "" {
		t.Errorf("fingerprint empty")
	}

	// /metrics text
	resp2, err := http.Get("http://" + healthAddr + "/metrics")
	if err != nil {
		t.Fatalf("GET /metrics: %v", err)
	}
	defer resp2.Body.Close()
	body, _ := io.ReadAll(resp2.Body)
	for _, want := range []string{
		"sshwap_uptime_seconds",
		"sshwap_active_sessions",
		"sshwap_total_connections 5",
		"sshwap_rejected_pty 2",
		"sshwap_rejected_caps 1",
	} {
		if !strings.Contains(string(body), want) {
			t.Errorf("metrics missing %q in body:\n%s", want, body)
		}
	}
}
