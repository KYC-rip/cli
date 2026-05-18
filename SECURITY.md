# Security

## Verifying the host key

Always verify the SSH host key fingerprint **before** typing a wallet
address into the swap form. The legitimate fingerprint is published in
two independent places — match against both:

- The HTTPS landing page: <https://swap.kyc.rip>
- The `About` tab inside the running TUI itself

Current fingerprint:
```
SHA256:wavvotTfJrgK/kY3qG3rdA3OY7Qs9sRXYXCi2tO8KYY
```

When you connect for the first time, your `ssh` client will display
the fingerprint and ask whether to trust it. **Compare the displayed
fingerprint to the one above before answering `yes`.** Any mismatch
is a man-in-the-middle and you should not proceed.

## What sshwap does and does not custody

- **No funds.** sshwap is a thin client over the kyc.rip aggregator
  REST API. Swap orders settle directly between the upstream
  exchange engine and your destination wallet. The VPS sees only the
  metadata you type into the form (assets, amount, address).
- **No accounts, no auth.** Connections are anonymous (`NoClientAuth`).
  Your username is purely a cosmetic header in the TUI.
- **No persistent state.** Session state lives only for the duration
  of the connection. Trade IDs are in your hands — write them down if
  you want to track later via the `Track` tab.

## Reporting a vulnerability

If you find a security issue, please email **security@kyc.rip** with:
- A clear description of the issue
- Steps to reproduce
- Potential impact

Please do **not** open a public GitHub issue for security problems.
We aim to acknowledge within 48 hours.

## Threat model

This service deliberately exposes a public unauthenticated SSH server.
The hardening we apply:

- Listens only on TCP/22 (no shell, no exec, no port-forwarding,
  no agent-forwarding, no subsystems — sessions are PTY-only).
- Per-IP and global concurrent-session caps.
- Per-session idle timeout, hard handshake timeout.
- Runs as non-root with `CAP_NET_BIND_SERVICE` (no privilege
  escalation surface).
- Outbound HTTPS only to `api.kyc.rip` from the VPS — not a general-
  purpose tunnel.

## Channels

- **clearnet**: `ssh swap.kyc.rip`
- **tor**: `torsocks ssh kyccli2b6y3iwxhkpoetzfyozmmrwipaakznyvhgl7a264l7tflvzqad.onion`
- **i2p**: `ssh kyccliymrfjyumorhpfujsqfvbb2vakajklhw5kfomec7umwhpgq.b32.i2p` (via i2pd SOCKS proxy)

All three channels go to the same SSH server with the same host key —
verify the fingerprint regardless of which channel you choose.
