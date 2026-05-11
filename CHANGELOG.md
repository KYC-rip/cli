# Changelog

All notable changes to **kyc-cli** and **sshwap** are documented here.
The format is loosely [Keep a Changelog](https://keepachangelog.com/);
versioning follows [SemVer](https://semver.org/).

## [0.1.37] — 2026-05-11

### Fixed
- **Ghost-mode banner border renders correctly.** The banner had a
  pinned `Width(cardInnerWidth)` and a skull glyph (☠) prefix. On
  terminals where ☠ measures as 2 cells but lipgloss scores it as 1,
  content overflowed the pinned inner width and the right/bottom
  borders of the box never drew. Dropped the fixed width (banner now
  sizes to content) and replaced ☠ with an ASCII `[GHOST]` label
  immune to emoji presentation rules.

## [0.1.36] — 2026-05-11

### Changed
- **`kyc-cli update` prints an `install.sh` fallback on check failure.**
  Network timeouts, transient API errors, or a future rate-limit
  surface now end with a concrete escape hatch — a one-line
  `curl ... | sh` that hits the CDN download path directly. The
  background nudge stays silent (we don't want to pester users on
  flaky networks).

## [0.1.35] — 2026-05-11

### Fixed
- **`kyc-cli update` no longer hits `api.github.com` (rate-limited).**
  Users behind shared NATs (Cloudflare WARP, mobile carriers, office
  networks) were hitting the 60-req/hour unauthenticated limit on
  `api.github.com/repos/.../releases/latest`. Switched to the HTML
  `github.com/.../releases/latest` endpoint, which 302-redirects to
  the latest tag and is not rate-limited the same way. The tag is
  parsed from the redirect Location header — no API key, no JSON.

## [0.1.34] — 2026-05-11

### Fixed
- **Local `kyc-cli` no longer launches into a blank screen.** `View()`
  short-circuits to `""` until `m.width > 0`. The SSH host seeds
  width/height from the PTY; the local CLI didn't, so on terminals
  where bubbletea's initial `WindowSizeMsg` is delayed (macOS Terminal
  with alt-screen, some Warp configs) the user saw nothing until they
  resized the window. Now seeds dimensions via
  `golang.org/x/term.GetSize` before constructing the program.

## [0.1.33] — 2026-05-10

### Changed
- **Select-mode toggle is now `ctrl+s`** (was `m`). The `m` binding only
  fired when no textinput was focused, so users couldn't flip into
  select-mode while at the amount/address/memo step — exactly when
  they need to copy a quote line above the input. `ctrl+s` lives in
  the always-on shortcut group with `ctrl+c`, so it works regardless
  of focus.

## [0.1.32] — 2026-05-10

### Added
- **`m` toggles select-mode** — runtime switch between mouse-click
  navigation (default) and native text-select. Some terminals (notably
  Warp) don't honor the modifier-bypass convention, so a hard toggle
  is the only reliable path to give users both interactions. Toast
  shows the active mode; deposit footer hint surfaces `m → select
  mode`. Implemented via `tea.DisableMouse` /
  `tea.EnableMouseCellMotion` commands so it just works at runtime.

## [0.1.31] — 2026-05-10

### Fixed
- **Restored tab/button mouse clicks (regression from v0.1.30).** v0.1.30
  dropped `tea.WithMouseCellMotion` to unlock native text-select, but
  that also disabled bubblezone hit-testing — tabs, buttons, and asset
  rows stopped responding to clicks. Re-enabled mouse-cell-motion;
  text-select still works via the terminal's modifier passthrough
  (⌥ on macOS, shift on Linux/Windows). Footer hint updated to
  surface the modifier.

## [0.1.30] — 2026-05-10

### Fixed
- **Native mouse-select-and-copy now works.** v0.1.29 still relied on
  OSC 52 escapes for the clipboard, which Warp and Termius gate behind
  a privacy setting that's OFF by default. With `tea.WithMouseCellMotion`
  enabled, the terminal forwarded every click to bubbletea, blocking
  native click-drag selection — leaving users with no working copy
  path. v0.1.30 drops mouse-cell-motion entirely so the user's
  terminal handles selection natively. OSC 52 (`c` / `C` / `enter`)
  still fires for terminals that have it enabled. Footer hint updated
  to mention mouse-select.

## [0.1.29] — 2026-05-10

### Fixed
- **TRC20 wallet URI no longer opens a native TRX transfer.** v0.1.17–
  0.1.28 emitted `tron:<addr>?amount=N` for USDT-TRC20 deposits; that
  URI scheme is interpreted by wallets as a native TRX transfer, which
  meant a user clicking it could send N TRX (not N USDT) to the deposit
  address. Now restricted to native TRX only — TRC20 token URIs need
  EIP-681-style with contract address; out of scope until done right.
- **Clipboard copy via mutex-protected out-of-band writer.** Codex
  review #3 flagged routing OSC 52 through `View()` as the weak link
  (bubbletea's renderer runs `ansi.Truncate`/`StringWidth`/line-diff
  over View output, not safe for device-control sequences). New
  `internal/tui/clipboard.go` defines `LockedWriter` — wraps the SSH
  session / stdout in a mutex; the model writes OSC 52 directly through
  it, sharing bubbletea's output stream without racing the renderer.
  Both `cmd/sshwap` and `cmd/kyc-cli` constructed via `tea.WithOutput
  (lockedWriter)`. View() is now purely visual.
- Track-tab `esc` and `resetSwap` now clear `qrFullScreen`,
  `qrImageMode`, `copyToast`, and `depositFocus` — those state flags
  were leaking into the next tracked order / next swap.

## [0.1.28] — 2026-05-10

### Added
- **Ghost tab** — privacy-routed multi-leg swap, the same product as
  `/ghost` on kyc.rip. Tab bar now reads `Swap | Ghost | Track`. Ghost
  reuses the swap wizard but routes API calls through `/v2/exchange/
  bridge/{estimate,create}`, which split the trade into legs so no
  single provider sees the full path. Bridge labels (`MONERO_TUNNEL`,
  `ZANO_PRIVACY_BRIDGE`, `FROST`, etc.) are surfaced on the route
  cards instead of provider names. Refund address defaults to the
  destination address — same fallback the Telegram bot uses.
- Ghost banner (☠ skull + accent border) above the wizard whenever
  Ghost mode is active.
- New `g` global hotkey switches to the Ghost tab. Image-mode toggle
  in the deposit panel renamed from `g` to `i` to free the keystroke.

### Changed
- About is no longer in the main tab bar — reachable only via the `a`
  hotkey or footer hint. The bar's three slots are reserved for the
  primary swap surfaces.
- `internal/api/client.go` — `Route` carries `BridgeLabel`,
  `BridgeBadge`, `BridgeHighlight`, `RequiresRefund` fields populated
  by the bridge endpoint. `CreateReq` gains `RefundAddress`. New
  `EstimateBridge` and `CreateBridge` methods. `CreateBridge` handles
  the bridge endpoint's array-or-single response (multi-leg trades
  return `[Trade, Trade…]`; we use the first leg as the user-facing
  trade and keep polling status on its id).

## [0.1.27] — 2026-05-09

### Fixed
- OSC 52 clipboard escape now uses the ST terminator (`ESC \`) instead
  of BEL (`\a`). Both are spec, but some terminals + tmux passthroughs
  only honor ST. Warp specifically appears to be one of them.

## [0.1.26] — 2026-05-09

### Fixed
- **OSC 52 clipboard copy now actually fires.** v0.1.22-25 was emitting
  the escape via `tea.Println(osc52Clipboard(...))` which doesn't
  reliably reach the terminal in alt-screen mode (Codex review). Switched
  to a one-frame `pendingOSC52` field that the top-level `View()`
  prepends to the next frame's output, then a 50ms `tea.Tick` clears it
  via a token-checked `clearOSC52Msg` so concurrent copies don't clobber
  each other. Address copy + QR-URL copy both now write through reliably.

## [0.1.25] — 2026-05-09

### Fixed
- **Fullscreen QR (`q`) now renders without horizontal dark bands.**
  Codex post-mortem identified the root cause: Bubble Tea's standard
  renderer suffixes every rendered line with `ESC[K` (erase line right)
  whenever `ansi.StringWidth(line) < r.width`, which paints the line
  trailer with the *current* BG — and Baozi's per-cell `ESC[0m` reset
  leaves the BG state at terminal default (dark). That suffix was
  bleeding into the row's box / line leading on Warp and Termius,
  producing the alternating module-row / dark-row pattern users kept
  seeing across 8 release iterations.
- Fix: pad every QR row to exactly `m.width` visible cells so the `[K`
  gate is never triggered. Also skip `lipgloss.Place` for the text-mode
  QR path — `Place`'s default-BG padding spaces have the same effect
  as the renderer's `[K` and undid the fix.

## [0.1.24] — 2026-05-09

### Fixed
- Arrow-key focus nav now works on the **Track tab** too. The handler
  was buried inside `updateSwap`, so on Track tab the textinput's
  cursor was eating up/down keys instead. Pulled the deposit-panel
  key handler out into a shared `handleDepositKeys` that runs before
  the per-tab dispatch.

## [0.1.23] — 2026-05-09

### Added
- **Arrow-key focus + enter to copy** in the deposit panel. `↑/↓`
  (or `j/k`) moves focus between the address row and the QR-URL row;
  a `▸` caret marks the focused item. `enter` copies it to the
  clipboard via OSC 52 — same mechanism as v0.1.22 but no longer
  requires memorising `c` / `C`. The shortcut keys still work for
  power users.

## [0.1.22] — 2026-05-09

### Added
- **`c` copies the deposit address to the system clipboard** via OSC 52
  escape — the terminal-side clipboard write travels in-band and works
  across SSH. Supported by Warp, iTerm2, Alacritty, kitty, mintty,
  WezTerm, Tabby, and most modern terminals.
- **`C` (shift+c) copies the QR URL** (`https://api.kyc.rip/v2/qr?d=…`)
  the same way — paste it into a browser to see the QR.
- "📋 copied" toast appears for 2 seconds after either keystroke.
- Address-line hint updated to `(press c to copy)` so the action is
  discoverable.

## [0.1.21] — 2026-05-09

### Added
- New `https://api.kyc.rip/v2/qr?d=<payload>` endpoint serves an HTML
  page that renders the QR client-side via `qrcode-min.js` from
  jsdelivr. No payload is logged or persisted by the worker — it's
  reflected straight into the page for the browser to render.
- CLI now prints the QR URL as **plain text below the OSC-8 link**:
  `or copy: https://api.kyc.rip/v2/qr?d=<urlencoded>`. Always works,
  no terminal click support required — select + copy + paste into a
  browser.

### Changed
- The "click to open in wallet" hint now reads "⌘+click" so users
  know the modifier is required when the app is consuming mouse
  clicks via bubbletea.

## [0.1.20] — 2026-05-09

### Changed
- **Default order/track view no longer renders the QR side-by-side.**
  Multiple iterations of in-terminal QR rendering (half-block,
  full-block, BG-color paint, switching libraries to mdp/qrterminal
  then Baozisoftware/qrcode-terminal-go) all hit terminal/font/
  bubbletea quirks. Defaulting to a layout that just works.
- Order/Track now show, in order: order info → deposit address (with
  OSC-8 wallet hyperlink) → `[ open QR in browser ]` link backed by
  a `data:image/png;base64,…` URL. Click in any modern terminal
  (Warp / iTerm2 / Terminal.app / Termius / VS Code / kitty) → the
  QR opens as an inline image in your default browser. No terminal
  rendering involved, no font dependency, scans 100%.
- The `q` (full-size in-terminal QR) and `g` (iTerm2 inline image)
  paths are still available for users who want to try them in
  terminals where they work.

## [0.1.19] — 2026-05-09

### Added
- **`g` key toggles iTerm2 inline-image protocol** for QR rendering.
  Press `g` in the order or track view (or while in fullscreen QR
  mode) to switch from text-mode QR (BG-painted spaces, broken on
  certain terminal/font combos) to a real raster PNG embedded via
  OSC 1337. Supported natively by Warp, iTerm2, WezTerm, mintty,
  Tabby. Terminals that don't recognize the OSC will print the raw
  base64 as garbage — that's why image mode is opt-in.
- The fullscreen view's hint now reads `q exit · g toggle image/text mode`.

## [0.1.18] — 2026-05-09

### Fixed
- Fullscreen QR (`q`) now bypasses the entire card/header/hint layout
  in `View()`. The styleCard wrapper has `Width(68)` + `Padding(1,2)` +
  rounded border, and applying that processing to a multi-line
  BG-painted QR string was injecting blank rows between QR rows
  (visible as alternating module-row / dark-gap pattern). Fullscreen
  now `lipgloss.Place`s the QR directly with no decoration.

## [0.1.17] — 2026-05-09

### Added
- **OSC-8 wallet hyperlink on the deposit address.** Click in the
  terminal to open the URI in your installed wallet — `bitcoin:…`,
  `tron:…?amount=N`, `monero:…?tx_amount=N`, `ethereum:…?value=wei`,
  `litecoin:…`, `dogecoin:…`, `bitcoincash:…`, `zcash:…`, `solana:…`.
  Skipped for EVM tokens (USDT-ERC20 etc.) — those need EIP-681 with
  per-chain contract address. Terminals that don't support OSC-8
  silently strip the escape and show plain text.
- **Full-size QR view.** Press `q` in the Order or Track view (when
  a deposit address is shown) to expand the QR to the whole terminal,
  bypassing the side-by-side layout that was clipping rows in tighter
  windows. `q` or `esc` returns.

## [0.1.16] — 2026-05-09

### Changed
- QR rendering swapped to `github.com/Baozisoftware/qrcode-terminal-go`
  — user reported this library renders correctly in their terminal
  stack (Warp + zsh + Hack font) where mdp/qrterminal still showed
  line-gap artifacts. Both libraries use BG-painted spaces, but
  Baozi's defaults (`BG ANSI 256 index 7` for light, `index 0` for
  dark) and tighter 1-module quiet zone evidently render cleaner.

## [0.1.15] — 2026-05-09

### Changed
- QR rendering now delegates to `github.com/mdp/qrterminal/v3` — the
  same library tailscale, 1password-cli, and a pile of others use for
  terminal QR output. Replaces the hand-rolled bitmap loop and the
  three iterations on glyph choice (half-block / full-block / BG paint).
  Default `Generate` emits ANSI 16-color BG-painted full-block cells
  (`\x1b[40m`/`\x1b[47m`) — the most universally compatible encoding.
- `github.com/skip2/go-qrcode` removed; replaced by `rsc.io/qr`
  (qrterminal's underlying encoder).

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
