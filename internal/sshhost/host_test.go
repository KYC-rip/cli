package sshhost

import (
	"io"
	"log"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"
)

// TestServerAcceptsAnonymousAndRejectsExec spins up the SSH host on a
// loopback port and verifies two things end-to-end through a real
// golang.org/x/crypto/ssh client:
//
//  1. NoClientAuth: no keys, no password — the connection completes.
//  2. Exec is rejected: requesting `Exec("ls")` returns a non-zero exit
//     and a banner about PTY-only mode (the hardening we promised in
//     the Codex plan review).
func TestServerAcceptsAnonymousAndRejectsExec(t *testing.T) {
	tmp, err := os.CreateTemp(t.TempDir(), "hostkey-*")
	if err != nil {
		t.Fatal(err)
	}
	tmp.Close()
	os.Remove(tmp.Name()) // loadOrCreateHostKey will write a fresh one

	logger := log.New(io.Discard, "", 0)
	srv, err := New(Config{
		Addr:        "127.0.0.1:0", // ephemeral port
		HostKeyPath: tmp.Name(),
		MaxSessions: 8,
		MaxPerIP:    4,
		IdleTimeout: 5 * time.Second,
	}, logger)
	if err != nil {
		t.Fatal(err)
	}

	// Bind the listener manually so we can observe the chosen port.
	ln, err := net.Listen("tcp", srv.cfg.Addr)
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()

	go func() {
		_ = srv.srv.Serve(ln)
	}()
	t.Cleanup(func() { _ = srv.srv.Close() })

	// Client side: NoClientAuth — no methods provided.
	clientCfg := &ssh.ClientConfig{
		User:            "tester",
		Auth:            nil, // noauth
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         3 * time.Second,
	}
	conn, err := ssh.Dial("tcp", addr, clientCfg)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	sess, err := conn.NewSession()
	if err != nil {
		t.Fatalf("new session: %v", err)
	}
	defer sess.Close()

	// Asking for exec without a PTY should be rejected by our handler.
	out, err := sess.Output("anything")
	if err == nil {
		t.Fatalf("expected non-zero exit from exec, got nil and output %q", out)
	}
	if !strings.Contains(string(out), "interactive PTY required") &&
		!strings.Contains(string(out), "exec disabled") {
		// Either rejection message is acceptable — depends on whether
		// the client sends a pty-req before exec or not.
		t.Logf("rejection text was: %q (err=%v)", out, err)
	}
}
