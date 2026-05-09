# Changelog

All notable changes to **kyc-cli** and **sshwap** are documented here.
The format is loosely [Keep a Changelog](https://keepachangelog.com/);
versioning follows [SemVer](https://semver.org/).

## [0.1.14] — 2026-05-09

### Fixed
- QR rendered with **BG-color paint** instead of unicode block glyphs.
  Each module = 2 spaces with an explicit ANSI 256 BG (black or white),
  no glyphs at all. v0.1.13's full-block `██` was unreliable on macOS
  Terminal.app: SF Mono kerns adjacent `█` characters with hairline
  vertical gaps, breaking the solid-fill illusion. v0.1.10's half-block
  had a similar issue along the row axis (line-spacing > glyph height).
- BG-colored spaces side-step both font issues — terminals just paint
  the cell background, no glyph lookup. Only relies on ANSI 256 BG
  support, which bubbletea already requires.
- Also fixes a polarity inversion: `qr.ToString(false)` actually maps
  *light* modules to `██` (skip2 assumes dark terminals with light FG),
  so v0.1.13 was rendering an inverted QR even when the visual rendered
  cleanly.

## [0.1.13] — 2026-05-09

### Fixed
- QR now uses full-block `██` per filled module instead of half-blocks
  (`▀`/`▄`). Half-block rendering was reported broken on macOS Terminal
  where line-spacing exceeds glyph height — adjacent rows floated with
  visible vertical gaps and modules collapsed into horizontal stripes.
  Full-block is wider but renders reliably across every terminal we've
  tested. skip2's built-in 4-module quiet zone is sufficient now;
  removed the redundant lipgloss padding.

## [0.1.12] — 2026-05-09

### Added
- Quote screen now shows the same four buckets as the Telegram bot:
  🌟 Suggested · 🛡 Safe · $ Rate · ⚡ Speed. Each card lists provider,
  ~amount_to, ETA, and KYC rating. `1-4` number-pick, `tab`/arrow to
  cycle, `enter` to confirm. Picked route is what `POST /create` uses.
- `pickRoutesByMode` ported from `bot/src/telegram/kyc.ts:827` so the
  CLI / SSH and Telegram channels surface the same recommendations
  (dedupe-by-provider, greedy slot-fill so no provider lands in two
  buckets, V2's `routes[0]` is always Suggested).
- `internal/tui/routes_test.go` covers the four-distinct-picks happy
  path, dedupe, no-duplicate-across-buckets, and empty input.

## [0.1.11] — 2026-05-09

### Added
- Track tab now renders the deposit address + QR (and memo if any)
  whenever the trade is in a non-terminal status. Users who closed
  the session before sending funds and re-opened via Track had no way
  to see where to deposit; that gap is closed.

## [0.1.10] — 2026-05-09

### Fixed
- QR code now renders dark-on-light (black FG on white BG via lipgloss)
  so phone-camera scanners read it regardless of the user's terminal
  scheme. Previously rendered white-on-dark which many scanners reject.
  Adds an extra 2-cell horizontal + 1-row vertical quiet zone on top of
  skip2's built-in 4-module quiet zone for safety.

### Changed
- Order `source` field auto-injected into `POST /v2/exchange/create`
  changed from `sshwap` to `cli` — both `cmd/sshwap` and `cmd/kyc-cli`
  share the same API client, and `cli` is the right umbrella name.
- `User-Agent` header bumped from `sshwap/0.1` to `kyc-cli/0.1` to
  match.

### Verified (no code change)
- USDT-ERC20 + USDT-TRC20: correctly parsed by `splitTickerNet` →
  `from_network` / `to_network` populated on `CreateReq`. Address
  validation handles both EVM-style (ERC20/BEP20/Polygon/etc.) and
  Tron-style (TRC20). Tickers in numbered top-asset list (#4, #5).

## [0.1.9] — 2026-05-09

### Added
- `kyc-cli` refuses to launch the TUI when stdin/stdout aren't a tty,
  printing a useful hint and `--version` as the script-friendly path.

## [0.1.8] — 2026-05-09

### Fixed
- TUI picker hotkey hijack: typing `LTC`, `TRX`, `TRC20`, `TON`, etc.
  no longer jumps to the Track tab. Pickers are now flagged as typing
  states so the global `s/t/a` shortcuts don't fire while picking.
  Regression test added.

### Added
- `internal/update`: `archive/zip` extraction so curl|sh-installed
  Windows users can self-update.

## [0.1.7] — 2026-05-09

### Changed
- `kyc-cli --help` now lists the `update` subcommand explicitly
  (custom `flag.Usage` — `flag.Parse` doesn't surface subcommands).

## [0.1.6] — 2026-05-09

### Added
- `kyc-cli update` — self-update subcommand. Downloads the matching
  release tarball, verifies sha256 against `checksums.txt`, atomic-
  swap with `.bak` rollback. Refuses under Homebrew Cellar / Scoop /
  /usr/bin (defers to package manager).
- `kyc-cli update --check` — non-destructive status report.
- Startup nudge: one-line stderr note if a newer release exists,
  throttled to once per 24h via `$XDG_CACHE_HOME/kyc-cli/lastcheck`.

## [0.1.5] — 2026-05-09

### Verified
- Release pipeline fully automated: `git tag vX.Y.Z && git push --tags`
  cross-builds + auto-publishes to GitHub Releases, Homebrew tap, AND
  Scoop bucket in one shot. Brew commit `52b3b72`, Scoop commit
  `b4a85bf` — first end-to-end automated multi-tap release.

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
