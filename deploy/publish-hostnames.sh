#!/usr/bin/env bash
# publish-hostnames.sh — read the live tor/i2pd hostnames and write them to a
# JSON file the sshwap service (and operator) can consume.
#
# Output: /etc/sshwap/hidden-endpoints.json
#   { "tor": "...onion", "i2p": "...b32.i2p", "updated": "<iso>" }

set -euo pipefail

if [ "$(id -u)" -ne 0 ]; then
  echo "must run as root" >&2; exit 1
fi

OUT=/etc/sshwap/hidden-endpoints.json
mkdir -p "$(dirname "$OUT")"

TOR=""
if [ -f /var/lib/tor/kyccli/hostname ]; then
  TOR=$(cat /var/lib/tor/kyccli/hostname)
fi

I2P=""
# Try the i2pd webconsole for the b32. (Default listens on 127.0.0.1:7070)
I2P=$(curl -s http://127.0.0.1:7070/?page=i2p_tunnels 2>/dev/null \
  | grep -oE '[a-z2-7]{52}\.b32\.i2p' | head -1 || true)
if [ -z "$I2P" ]; then
  # Fallback: pull from local_destinations page.
  I2P=$(curl -s http://127.0.0.1:7070/?page=local_destinations 2>/dev/null \
    | grep -oE '[a-z2-7]{52}\.b32\.i2p' | head -1 || true)
fi

cat > "$OUT" <<EOF
{
  "tor": "${TOR}",
  "i2p": "${I2P}",
  "updated": "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
}
EOF

echo "wrote $OUT:"
cat "$OUT"
