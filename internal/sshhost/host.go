package sshhost

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	gssh "github.com/gliderlabs/ssh"
	"github.com/muesli/termenv"
	"golang.org/x/crypto/ssh"

	"github.com/xbtoshi/sshwap/internal/api"
	"github.com/xbtoshi/sshwap/internal/tui"
)

// Config controls listener and runtime caps.
type Config struct {
	Addr            string        // ":2222"
	HostKeyPath     string        // path to ed25519 host key (auto-created if missing)
	APIBase         string        // override https://api.kyc.rip
	APIKey          string        // optional X-API-Key
	MaxSessions     int           // global cap (default 200)
	MaxPerIP        int           // per-source-IP cap (default 3)
	IdleTimeout     time.Duration // session idle timeout (default 90s)
	HandshakeWindow time.Duration // hard cap on time-to-PTY (default 8s)
	Banner          string        // pre-PTY banner text
}

// Defaults applies sensible POC defaults to zero-valued fields.
func (c *Config) Defaults() {
	if c.Addr == "" {
		c.Addr = ":2222"
	}
	if c.MaxSessions == 0 {
		c.MaxSessions = 200
	}
	if c.MaxPerIP == 0 {
		c.MaxPerIP = 3
	}
	if c.IdleTimeout == 0 {
		c.IdleTimeout = 90 * time.Second
	}
	if c.HandshakeWindow == 0 {
		c.HandshakeWindow = 8 * time.Second
	}
}

// Server wraps a gliderlabs/ssh server with hard caps.
type Server struct {
	cfg     Config
	client  *api.Client
	srv     *gssh.Server
	logger  *log.Logger
	active  int64
	perIP   sync.Map // string -> *atomic.Int64
	hostKey ssh.Signer
}

func New(cfg Config, logger *log.Logger) (*Server, error) {
	cfg.Defaults()
	if logger == nil {
		logger = log.New(os.Stderr, "[sshwap] ", log.LstdFlags|log.Lmicroseconds)
	}
	// Bubbletea/lipgloss detect color support against the host's stdout
	// when WithOutput() forwards to an SSH session writer that isn't a tty.
	// Force true-color globally so every connecting client gets the full
	// palette; clients with reduced color support down-rank locally.
	lipgloss.SetColorProfile(termenv.TrueColor)
	signer, fingerprint, err := loadOrCreateHostKey(cfg.HostKeyPath)
	if err != nil {
		return nil, fmt.Errorf("host key: %w", err)
	}
	logger.Printf("host key fingerprint: %s", fingerprint)
	cli := api.New(api.WithBase(cfg.APIBase), api.WithAPIKey(cfg.APIKey))
	s := &Server{cfg: cfg, client: cli, logger: logger, hostKey: signer}
	s.srv = &gssh.Server{
		Addr:        cfg.Addr,
		IdleTimeout: cfg.IdleTimeout,
		MaxTimeout:  60 * time.Minute,
		Handler:     s.handle,
		// No PasswordHandler / PublicKeyHandler / KeyboardInteractiveHandler =
		// gliderlabs sets NoClientAuth=true. Anonymous SSH; the TUI itself
		// is the entire surface.
		// Reject everything that isn't a session: no port-forward, no
		// agent-forward, no exec, no subsystem.
		LocalPortForwardingCallback:   func(ctx gssh.Context, h string, p uint32) bool { return false },
		ReversePortForwardingCallback: func(ctx gssh.Context, h string, p uint32) bool { return false },
		ChannelHandlers: map[string]gssh.ChannelHandler{
			"session": gssh.DefaultSessionHandler,
		},
		Banner: cfg.Banner,
	}
	s.srv.AddHostKey(signer)
	return s, nil
}

// Fingerprint returns the SHA256 fingerprint of the host key.
func (s *Server) Fingerprint() string {
	return ssh.FingerprintSHA256(s.hostKey.PublicKey())
}

// ListenAndServe blocks until ctx is cancelled.
func (s *Server) ListenAndServe(ctx context.Context) error {
	go func() {
		<-ctx.Done()
		_ = s.srv.Close()
	}()
	s.logger.Printf("sshwap listening on %s", s.cfg.Addr)
	if err := s.srv.ListenAndServe(); err != nil && err != gssh.ErrServerClosed {
		return err
	}
	return nil
}

func (s *Server) handle(sess gssh.Session) {
	remote := sess.RemoteAddr()
	host, _, _ := net.SplitHostPort(remote.String())

	// Caps
	if cur := atomic.AddInt64(&s.active, 1); cur > int64(s.cfg.MaxSessions) {
		atomic.AddInt64(&s.active, -1)
		_, _ = sess.Write([]byte("server busy — try again in a moment\r\n"))
		_ = sess.Exit(1)
		return
	}
	defer atomic.AddInt64(&s.active, -1)

	if !s.acquireIP(host) {
		_, _ = sess.Write([]byte("too many connections from your address\r\n"))
		_ = sess.Exit(1)
		return
	}
	defer s.releaseIP(host)

	// Reject anything that isn't a PTY session
	pty, winCh, isPty := sess.Pty()
	if !isPty {
		_, _ = sess.Write([]byte("interactive PTY required: try `ssh -t swap.kyc.rip`\r\n"))
		_ = sess.Exit(2)
		return
	}
	if cmd := strings.Join(sess.Command(), " "); cmd != "" {
		_, _ = sess.Write([]byte("exec disabled — interactive only\r\n"))
		_ = sess.Exit(2)
		return
	}

	s.logger.Printf("session start %s term=%s size=%dx%d", host, pty.Term, pty.Window.Width, pty.Window.Height)

	// Build TUI bound to this session.
	// Seed initial Width/Height on the model so View() renders on first
	// frame (otherwise alt-screen flips with empty content and the user
	// sees a black void until they press a key).
	user := sess.User()
	if user == "" {
		user = "guest"
	}
	cfg := tui.Config{
		Client:        s.client,
		Fingerprint:   s.Fingerprint(),
		Username:      user,
		InitialWidth:  pty.Window.Width,
		InitialHeight: pty.Window.Height,
	}
	prog := tea.NewProgram(
		tui.New(cfg),
		tea.WithInput(sess),
		tea.WithOutput(sess),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	go func() {
		for w := range winCh {
			prog.Send(tea.WindowSizeMsg{Width: w.Width, Height: w.Height})
		}
	}()

	if _, err := prog.Run(); err != nil {
		s.logger.Printf("session %s tui error: %v", host, err)
	}
	s.logger.Printf("session end %s", host)
}

func (s *Server) acquireIP(ip string) bool {
	v, _ := s.perIP.LoadOrStore(ip, new(atomic.Int64))
	cnt := v.(*atomic.Int64)
	if cnt.Add(1) > int64(s.cfg.MaxPerIP) {
		cnt.Add(-1)
		return false
	}
	return true
}

func (s *Server) releaseIP(ip string) {
	v, ok := s.perIP.Load(ip)
	if !ok {
		return
	}
	cnt := v.(*atomic.Int64)
	if cnt.Add(-1) <= 0 {
		s.perIP.Delete(ip)
	}
}

// resolveHostKeyPath defaults to $XDG_DATA_HOME/sshwap/host_ed25519 with
// ~/.local/share fallback.
func resolveHostKeyPath(p string) string {
	if p != "" {
		return p
	}
	if d := os.Getenv("XDG_DATA_HOME"); d != "" {
		return filepath.Join(d, "sshwap", "host_ed25519")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share", "sshwap", "host_ed25519")
}
