# Changelog

All notable changes to **sshwap** and **kyc-cli** are documented here.
The format is loosely [Keep a Changelog](https://keepachangelog.com/);
versioning follows [SemVer](https://semver.org/).

## [Unreleased]

### Added
- Health endpoint at `127.0.0.1:9090` (`/healthz` JSON, `/metrics`
  Prometheus 0.0.4 plaintext) for monitoring.
- Pre-flight destination-address format validation (BTC, ETH/EVM, XMR,
  SOL, TRX, LTC, DOGE, BCH, ZEC) — catches typos before the order
  reaches an upstream engine that might already commit funds.
- `--version` flag on both binaries (set via `-ldflags`).
- `LICENSE` (MIT).
- `SECURITY.md` with host-key verification protocol and threat model.
- `.goreleaser.yml` + `.github/workflows/{ci,release}.yml` so a
  `git push --tags` cuts a cross-built GitHub Release.
- Integration test using a real `golang.org/x/crypto/ssh` client
  against the live SSH host.
- `Makefile` targets `release-snapshot`, `release-check`.

## [0.1.0] — 2026-05-09 (planned)

### Added
- SSH-only swap TUI (`cmd/sshwap`) and a downloadable equivalent
  (`cmd/kyc-cli`) sharing the same `internal/tui` + `internal/api`.
- Wizard flow: Pick From → Pick To → Amount → Address → Memo →
  Quote → Confirm → Order. Numbered (1-9) digit shortcuts and
  click-to-pick rows.
- Mouse support via `bubblezone` (tabs, asset rows, primary buttons).
- ANSI 256-colour palette (saturated lime/gold) tested across
  macOS Terminal, iTerm2, Alacritty, Kitty.
- Track tab for trade-status lookups.
- About tab with channel list (clearnet / Tor / I2P / HTTPS landing)
  and the host-key fingerprint.
- Clearnet on `swap.kyc.rip:22` with HTTPS landing (LE cert,
  auto-renew via deploy hook).
- Tor hidden service:
  `ozz6kgrbp6epsxhrid456udvwj3vzecb4f7jz5orxcrpxn4f2bejuyid.onion`
- I2P b32:
  `r4ziaqaec7w73x7ltpz5pi5kswclgjdw6ioyz25mbtrisprneqhq.b32.i2p`
- Hardening: `NoClientAuth`, PTY-only, exec/forward rejected,
  global+per-IP session caps, idle/handshake timeouts, non-root
  with `CAP_NET_BIND_SERVICE`.
