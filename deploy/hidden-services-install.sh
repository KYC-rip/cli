#!/usr/bin/env bash
# hidden-services-install.sh — set up Tor + I2P endpoints for swap.kyc.rip's SSH service.
#
# Run as root on the swap.kyc.rip box ONCE.
# Adds:
#   - tor (HiddenServiceDir -> /var/lib/tor/kyccli/, listen 127.0.0.1:22)
#   - i2pd (server tunnel -> 127.0.0.1:22)
#   - mkp224o (vanity-onion miner, run separately via vanity-mine.sh)
#
# Two-phase flow:
#   1. This script: install + start with a random v3 onion + random b32 i2p.
#      Services are usable immediately at random hostnames.
#   2. vanity-mine.sh: mines a `kyccli` prefix in the background. When matched,
#      run vanity-swap.sh to atomically replace the key dir + restart tor.

set -euo pipefail

if [ "$(id -u)" -ne 0 ]; then
  echo "must run as root" >&2; exit 1
fi

echo "[1/5] apt update + install tor + i2pd + tools"
apt-get update -qq
apt-get install -y tor i2pd jq curl

echo "[2/5] writing /etc/tor/torrc.d/sshwap.conf"
mkdir -p /etc/tor/torrc.d
cat > /etc/tor/torrc.d/sshwap.conf <<'EOF'
# sshwap (kyc.rip SSH swap) — v3 hidden service
HiddenServiceDir /var/lib/tor/kyccli/
HiddenServiceVersion 3
HiddenServicePort 22 127.0.0.1:22
EOF

# Some Debian builds don't auto-include /etc/tor/torrc.d/*. Make sure it's wired.
if ! grep -q "%include /etc/tor/torrc.d" /etc/tor/torrc; then
  echo "" >> /etc/tor/torrc
  echo "%include /etc/tor/torrc.d/" >> /etc/tor/torrc
fi

echo "[3/5] writing /etc/i2pd/tunnels.d/sshwap.conf"
mkdir -p /etc/i2pd/tunnels.d
cat > /etc/i2pd/tunnels.d/sshwap.conf <<'EOF'
[sshwap]
type = server
host = 127.0.0.1
port = 22
inport = 22
keys = sshwap.dat
gzip = false
EOF

echo "[4/5] enabling + restarting services"
systemctl enable --now tor
systemctl restart tor
systemctl enable --now i2pd
systemctl restart i2pd

echo "[5/5] waiting up to 30s for endpoints to publish"
for i in $(seq 1 30); do
  if [ -f /var/lib/tor/kyccli/hostname ] && \
     [ -f /var/lib/i2pd/sshwap.dat ]; then
    break
  fi
  sleep 1
done

echo ""
echo "=== installed ==="
if [ -f /var/lib/tor/kyccli/hostname ]; then
  echo "tor onion (random for now): $(cat /var/lib/tor/kyccli/hostname)"
else
  echo "tor onion: not yet generated — check 'journalctl -u tor'"
fi

if [ -f /var/lib/i2pd/sshwap.dat ]; then
  # i2pd doesn't expose b32 directly; pull it from the running webconsole.
  B32=$(curl -s http://127.0.0.1:7070/?page=local_destinations 2>/dev/null \
    | grep -oE '[a-z2-7]{52}\.b32\.i2p' | head -1 || true)
  if [ -n "$B32" ]; then
    echo "i2p b32: $B32"
  else
    echo "i2p b32: tunnel up but b32 not yet visible — try 'curl http://127.0.0.1:7070/?page=i2p_tunnels' in a minute"
  fi
fi

echo ""
echo "next: run ./vanity-mine.sh to grind a 'kyccli'-prefix onion in the background"
