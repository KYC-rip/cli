package tui

import (
	"io"
	"sync"
)

// LockedWriter is the same io.Writer Bubble Tea is configured with via
// tea.WithOutput, wrapped in a mutex so the model's clipboard-write path
// (OSC 52 emission) can share the single output stream without racing
// the renderer's frame writes.
//
// Codex review (2026-05-09 23:50) flagged routing OSC 52 through View()
// as the weak link — bubbletea's standard renderer runs ansi.Truncate /
// ansi.StringWidth / line-diff over View output, which is not a safe
// transport for device-control sequences. Direct out-of-band writes via
// this locked writer keep View() purely visual.
type LockedWriter struct {
	mu sync.Mutex
	w  io.Writer
}

func NewLockedWriter(w io.Writer) *LockedWriter {
	return &LockedWriter{w: w}
}

func (lw *LockedWriter) Write(p []byte) (int, error) {
	if lw == nil {
		return 0, nil
	}
	lw.mu.Lock()
	defer lw.mu.Unlock()
	return lw.w.Write(p)
}
