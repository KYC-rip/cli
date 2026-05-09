#!/usr/bin/env bash
# One-shot installer for the sshwap host on a fresh Debian/Ubuntu VPS.
# Run as root. Idempotent.
set -euo pipefail

VER=${1:-latest}
GOOS=linux
GOARCH=$(uname -m)
case "$GOARCH" in
  x86_64) GOARCH=amd64 ;;
  aarch64|arm64) GOARCH=arm64 ;;
  *) echo "unsupported arch: $GOARCH" >&2; exit 1 ;;
esac

# 1. user
id -u sshwap >/dev/null 2>&1 || useradd --system --no-create-home --shell /usr/sbin/nologin sshwap

# 2. dirs
install -d -o sshwap -g sshwap -m 0700 /var/lib/sshwap
install -d -m 0755 /etc/sshwap

# 3. binary
# (Caller is expected to scp the binary to /tmp/sshwap before running this,
#  or to replace this block with a release-download URL once we cut tags.)
if [ -f /tmp/sshwap ]; then
  install -o root -g root -m 0755 /tmp/sshwap /usr/local/bin/sshwap
elif [ -f ./bin/sshwap ]; then
  install -o root -g root -m 0755 ./bin/sshwap /usr/local/bin/sshwap
else
  echo "binary not found at /tmp/sshwap or ./bin/sshwap" >&2
  exit 1
fi

# 4. systemd
install -m 0644 ./deploy/sshwap.service /etc/systemd/system/sshwap.service
systemctl daemon-reload
systemctl enable --now sshwap

# 5. firewall hint (optional; idempotent)
if command -v ufw >/dev/null && ufw status | grep -q "Status: active"; then
  ufw allow 2222/tcp || true
fi

# 6. fingerprint
sleep 1
journalctl -u sshwap --no-pager -n 20 | grep -i fingerprint || true

echo
echo "[ok] sshwap installed. Test:  ssh -p 2222 anyuser@$(hostname -I | awk '{print $1}')"
echo "    Then create a DNS-only A record swap.kyc.rip -> this host."
