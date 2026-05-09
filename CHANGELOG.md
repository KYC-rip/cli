# Changelog

All notable changes to **kyc-cli** and **sshwap** are documented here.
The format is loosely [Keep a Changelog](https://keepachangelog.com/);
versioning follows [SemVer](https://semver.org/).

## [0.1.4] — 2026-05-09

### Added
- README badges (release / CI / Go report / license / brew tap).
- GitHub repo metadata + topics for discoverability.
- `kyc-rip/scoop-bucket` repo created; `.goreleaser.yml` now publishes
  a Scoop manifest alongside the Homebrew formula on every tag.
- `cli.kyc.rip` installer alias — short `curl cli.kyc.rip/install.sh`.

### Changed
- Landing page restructured with an **Or run it locally** section:
  brew, curl-pipe, Scoop snippets, all on `https://swap.kyc.rip`.
- HTTP traffic on `swap.kyc.rip` now 301-redirects to HTTPS.

## [0.1.3] — 2026-05-09

### Added
- Track tab auto-refreshes trade status every 5s until terminal state.
- Brew formula auto-update via `HOMEBREW_TAP_TOKEN` PAT.

## [0.1.2] — 2026-05-09

### Changed
- Repo renamed `kyc-rip/sshwap` → **`kyc-rip/cli`**. Module path
  updated to `github.com/kyc-rip/cli`. Binary names unchanged
  (`kyc-cli`, `sshwap`). Old GitHub URLs auto-redirect.

## [0.1.0–0.1.1] — 2026-05-09

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
