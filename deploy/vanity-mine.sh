#!/usr/bin/env bash
# vanity-mine.sh — mine a v3 onion address with prefix `kyccli` using mkp224o.
# Run as root on the swap.kyc.rip box AFTER hidden-services-install.sh.
# Designed to run for hours under tmux/nohup; check on it occasionally.
#
# Output: /root/vanity-onions/kyccli*/  (one or more matched key dirs)
# Swap in with: ./vanity-swap.sh <matched-dir>

set -euo pipefail

if [ "$(id -u)" -ne 0 ]; then
  echo "must run as root" >&2; exit 1
fi

PREFIX="${1:-kyccli}"
OUT_DIR="/root/vanity-onions"

# Build mkp224o from source if not installed (no Debian package).
if ! command -v mkp224o >/dev/null 2>&1; then
  echo "[mine] installing build deps + cloning mkp224o"
  apt-get install -y -qq build-essential autoconf libsodium-dev git
  cd /usr/local/src
  if [ ! -d mkp224o ]; then
    git clone https://github.com/cathugger/mkp224o.git
  fi
  cd mkp224o
  ./autogen.sh
  ./configure
  make -j"$(nproc)"
  install -m 0755 mkp224o /usr/local/bin/mkp224o
  cd "$OLDPWD"
fi

mkdir -p "$OUT_DIR"
echo "[mine] grinding prefix '$PREFIX' — Ctrl-C to stop"
echo "[mine] expected time: 4ch ~minutes · 5ch ~hour · 6ch ~hours · 7ch+ overnight"
echo "[mine] output dir: $OUT_DIR"
echo ""

# -n 1: stop after first match. Increase for multiple candidates.
# -t $(nproc): use all CPU cores.
mkp224o -d "$OUT_DIR" -t "$(nproc)" -n 1 "$PREFIX"

echo ""
echo "[mine] done — matched dir(s):"
ls -1 "$OUT_DIR" | grep "^$PREFIX"
