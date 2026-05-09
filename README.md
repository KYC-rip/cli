# kyc-rip / cli

[![release](https://img.shields.io/github/v/release/kyc-rip/cli?style=flat-square&color=00d7af)](https://github.com/kyc-rip/cli/releases)
[![ci](https://img.shields.io/github/actions/workflow/status/kyc-rip/cli/ci.yml?branch=master&style=flat-square&label=ci)](https://github.com/kyc-rip/cli/actions/workflows/ci.yml)
[![go report](https://goreportcard.com/badge/github.com/kyc-rip/cli?style=flat-square)](https://goreportcard.com/report/github.com/kyc-rip/cli)
[![license](https://img.shields.io/github/license/kyc-rip/cli?style=flat-square)](LICENSE)
[![brew](https://img.shields.io/badge/brew-kyc--rip%2Ftap%2Fkyc--cli-FFD700?style=flat-square)](https://github.com/kyc-rip/homebrew-tap)

**A privacy-first crypto swap aggregator, served as a TUI over SSH.**

```
$ ssh swap.kyc.rip
```

That's it. No browser, no JS, no cookies, no fingerprint. The
[kyc.rip](https://kyc.rip) aggregator (≈10 exchange engines, all major
chains) rendered in your terminal.

Also runs as a downloadable local CLI — same code, same flow, same
colours.

---

## Quick start

### Use the hosted SSH server

```sh
ssh swap.kyc.rip
```

**Verify the host key fingerprint the first time you connect:**

```
SHA256:wavvotTfJrgK/kY3qG3rdA3OY7Qs9sRXYXCi2tO8KYY
```

It is also published on
[`https://swap.kyc.rip`](https://swap.kyc.rip) and inside the TUI's
*About* tab. If your `ssh` client shows a different fingerprint, **do
not proceed** — that's a man-in-the-middle.

#### Encrypted channels

| Channel | Connect |
|---|---|
| clearnet | `ssh swap.kyc.rip` |
| Tor | `torsocks ssh ozz6kgrbp6epsxhrid456udvwj3vzecb4f7jz5orxcrpxn4f2bejuyid.onion` |
| I2P | `ssh -o ProxyCommand='nc -X 5 -x 127.0.0.1:4447 %h %p' r4ziaqaec7w73x7ltpz5pi5kswclgjdw6ioyz25mbtrisprneqhq.b32.i2p` |

All three channels go to the same SSH server with the same host key.

### Install the local CLI

```sh
# macOS / linux via Homebrew
brew install kyc-rip/tap/kyc-cli

# linux / macOS one-liner (verifies sha256, drops binary in PATH)
curl -fsSL https://cli.kyc.rip/install.sh | sh

# Windows / arbitrary unix
# Grab a tarball/zip from the GitHub Releases page:
#   https://github.com/kyc-rip/cli/releases
```

Then run:

```sh
kyc-cli
```

The local CLI talks to `api.kyc.rip` directly — no SSH layer involved.

---

## How it works

The wizard walks you through:

```
PickFrom → PickTo → Amount → Address → Memo
   ↓         ↓        ↓        ↓        ↓
        Quote (best route across all engines)
                       ↓
                   Confirm
                       ↓
              Deposit address + QR
                       ↓
            Auto-poll status (5s) → Done
```

- **Numbered shortcuts** (`1`–`9`) on the asset pickers, plus free-text
  ticker entry (e.g. `USDT-TRC20`).
- **Mouse support** — click tabs, asset rows, buttons.
- **Pre-flight address validation** for BTC, ETH/EVM, XMR, SOL, TRX,
  LTC, DOGE, BCH, ZEC. Catches typos before they cost you money.
- **Track tab** — paste a trade ID, watch the status auto-refresh.

---

## Repo layout

```
cmd/sshwap          SSH server entry point  (server-side, runs the host)
cmd/kyc-cli         Local CLI entry point   (client-side, what users install)

internal/tui        Bubble Tea model + Lip Gloss styles + bubblezone (mouse)
internal/api        REST client for /v2/exchange/{currencies,estimate,create,status}
internal/sshhost    gliderlabs/ssh wrapper, hardening, /healthz + /metrics

deploy/             nginx vhost, systemd unit, install.sh
scripts/install.sh  curl|sh installer (verifies sha256)
.goreleaser.yml     cross-build pipeline (linux/macos/windows × amd64/arm64)
```

The two binaries share `internal/tui` and `internal/api` 1:1.

---

## Build from source

Requires Go 1.25+.

```sh
git clone https://github.com/kyc-rip/cli
cd cli
make build           # → ./bin/sshwap and ./bin/kyc-cli
make test            # go test -race ./...
```

Or directly:

```sh
go install github.com/kyc-rip/cli/cmd/kyc-cli@latest
```

---

## Self-host your own SSH server

The SSH host is a single static binary; the systemd unit and an installer
helper are in [`deploy/`](deploy/).

```sh
make sshwap                                            # build
scp ./bin/sshwap deploy/sshwap.service user@vps:/tmp/  # ship
ssh user@vps sudo bash -s < deploy/install.sh          # install
```

Defaults: binds `:22` (or `:2222` if you keep OpenSSH on 22), runs as a
non-root `sshwap` user with `CAP_NET_BIND_SERVICE`, persists the host
key at `/var/lib/sshwap/host_ed25519`, exposes `/healthz` and `/metrics`
on `127.0.0.1:9090`.

Hardening baked in: PTY-only, no exec / forward / subsystem / agent
forwarding; per-IP and global session caps; idle + handshake timeouts.

---

## Security

See [SECURITY.md](SECURITY.md) for the threat model and responsible-
disclosure contact.

**TL;DR:** sshwap holds no funds, no accounts, and no persistent state.
It's a thin client over the kyc.rip aggregator REST API; orders settle
directly between the upstream exchange engine and your destination
wallet.

---

## License

MIT — see [LICENSE](LICENSE).

---

- Website · <https://kyc.rip>
- Landing · <https://swap.kyc.rip>
- Source · <https://github.com/kyc-rip/cli>
- Releases · <https://github.com/kyc-rip/cli/releases>
- Brew tap · <https://github.com/kyc-rip/homebrew-tap>
