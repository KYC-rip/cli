#!/usr/bin/env bash
# vanity-swap.sh — atomically replace the tor HiddenServiceDir keys with a
# mined vanity match. Run after vanity-mine.sh produces a kyccli* dir.
#
# Usage:
#   ./vanity-swap.sh /root/vanity-onions/kyccliXXXXXX.onion/

set -euo pipefail

if [ "$(id -u)" -ne 0 ]; then
  echo "must run as root" >&2; exit 1
fi

SRC="${1:?usage: $0 <mined-key-dir>}"
DEST="/var/lib/tor/kyccli"

if [ ! -d "$SRC" ]; then
  echo "no such dir: $SRC" >&2; exit 1
fi
if [ ! -f "$SRC/hs_ed25519_secret_key" ] || [ ! -f "$SRC/hostname" ]; then
  echo "missing required files in $SRC (hs_ed25519_secret_key, hostname)" >&2; exit 1
fi

echo "[swap] new hostname: $(cat "$SRC/hostname")"
echo "[swap] backing up current keys"
if [ -d "$DEST" ]; then
  mv "$DEST" "${DEST}.bak.$(date +%s)"
fi

mkdir -p "$DEST"
cp "$SRC/hs_ed25519_secret_key" "$DEST/"
cp "$SRC/hs_ed25519_public_key" "$DEST/" 2>/dev/null || true
cp "$SRC/hostname"              "$DEST/"
chown -R debian-tor:debian-tor "$DEST"
chmod 700 "$DEST"
chmod 600 "$DEST"/*

echo "[swap] reloading tor"
systemctl reload tor

echo "[swap] done — new endpoint live:"
echo "  $(cat "$DEST/hostname")"
echo ""
echo "next: update sshwap docs/landing/TUI with the new address (see publish-hostnames.sh)"
