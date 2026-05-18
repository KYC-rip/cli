# sshwap — Tor + I2P endpoint setup

Bring up real v3 onion + I2P endpoints for the SSH swap service.

Until you run these, the addresses advertised in the CLI / landing page /
README are **placeholders** that resolve to nothing — they were committed
before the hidden services were configured.

## On the `swap.kyc.rip` box

```sh
# 1. Install tor + i2pd, wire torrc + tunnels, start services with random hostnames.
sudo ./deploy/hidden-services-install.sh

# 2. Show what the random hostnames are right now (services are already usable):
sudo ./deploy/publish-hostnames.sh

# 3. Optional but recommended: mine a vanity-prefix onion. Runs for hours.
#    Default prefix is `kyccli` (6 chars, matches the binary brand). Pass a
#    different prefix as an argument if you want.
#
#    Run under tmux so you can detach and come back:
sudo tmux new -s mine
sudo ./deploy/vanity-mine.sh kyccli
# Ctrl-b d to detach. Re-attach with: sudo tmux a -t mine

# 4. When mkp224o reports a match, swap it in (atomic — tor reload, no downtime):
ls /root/vanity-onions/        # find the matched dir
sudo ./deploy/vanity-swap.sh /root/vanity-onions/kyccli<...>.onion/

# 5. Refresh the published endpoints file:
sudo ./deploy/publish-hostnames.sh
```

## What lives where

| Path | Purpose |
|---|---|
| `/etc/tor/torrc.d/sshwap.conf` | tor HiddenServiceDir + HiddenServicePort 22 |
| `/var/lib/tor/kyccli/` | onion private key + `hostname` file |
| `/etc/i2pd/tunnels.d/sshwap.conf` | i2pd server tunnel block (inport 22 → 127.0.0.1:22) |
| `/var/lib/i2pd/sshwap.dat` | i2p destination keys |
| `/etc/sshwap/hidden-endpoints.json` | published hostnames for the app/docs |

## I2P note

I2P b32 addresses are SHA-256 hashes of the destination key — you can't
grind a vanity prefix the way Tor lets you. The closer equivalent is
registering a human-readable name like `sshwap.i2p` via the `stats.i2p`
or `no.i2p` jump services after the tunnel is up. Do that separately;
the random b32 keeps working regardless.

## After the swap

Once the vanity onion is live and `hidden-endpoints.json` reflects it,
update the advertised strings in:

- `README.md`
- `SECURITY.md`
- `CHANGELOG.md`
- `internal/tui/model.go` (the about-screen lines that hard-code the onion)
- `deploy/swap.kyc.rip.conf` (embedded landing HTML)
- `ui/src/pages/CliLanding.tsx` (in `kyc-rip/kyc-rip/ui`)

…then redeploy the website and rebuild the CLI binaries.
