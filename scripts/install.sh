#!/usr/bin/env bash
# install.sh — fetches the latest kyc-cli release from GitHub and
# drops the binary into ~/.local/bin (or /usr/local/bin if writable).
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/kyc-rip/cli/master/scripts/install.sh | sh
#
# or pin a version:
#   curl -fsSL https://raw.githubusercontent.com/kyc-rip/cli/master/scripts/install.sh | sh -s v0.1.2
#
# Set INSTALL_DIR to override the install location:
#   INSTALL_DIR=/opt/bin curl -fsSL .../install.sh | sh
set -eu

REPO="kyc-rip/cli"
BIN_NAME="kyc-cli"
VERSION="${1:-latest}"
INSTALL_DIR="${INSTALL_DIR:-}"

# Detect OS / arch
case "$(uname -s)" in
  Linux)  OS=linux ;;
  Darwin) OS=darwin ;;
  *) echo "[install.sh] unsupported OS: $(uname -s)" >&2; exit 1 ;;
esac
case "$(uname -m)" in
  x86_64|amd64)  ARCH=amd64 ;;
  arm64|aarch64) ARCH=arm64 ;;
  *) echo "[install.sh] unsupported arch: $(uname -m)" >&2; exit 1 ;;
esac

# Resolve version → tag
if [ "$VERSION" = "latest" ]; then
  TAG=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | sed -nE 's/.*"tag_name": *"([^"]+)".*/\1/p' | head -1)
  if [ -z "$TAG" ]; then
    echo "[install.sh] could not resolve latest release tag" >&2
    exit 1
  fi
else
  TAG="$VERSION"
fi
VER_NUMERIC="${TAG#v}"

# Pick install dir
if [ -z "$INSTALL_DIR" ]; then
  if [ -w /usr/local/bin ]; then
    INSTALL_DIR=/usr/local/bin
  else
    INSTALL_DIR="$HOME/.local/bin"
    mkdir -p "$INSTALL_DIR"
  fi
fi

URL="https://github.com/${REPO}/releases/download/${TAG}/${BIN_NAME}_${VER_NUMERIC}_${OS}_${ARCH}.tar.gz"
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

echo "[install.sh] downloading ${URL}"
curl -fsSL "$URL" -o "$TMP/kyc-cli.tar.gz"

# Verify checksum (best-effort)
if curl -fsSL "https://github.com/${REPO}/releases/download/${TAG}/checksums.txt" -o "$TMP/checksums.txt" 2>/dev/null; then
  EXPECTED=$(grep "${BIN_NAME}_${VER_NUMERIC}_${OS}_${ARCH}.tar.gz" "$TMP/checksums.txt" | awk '{print $1}')
  ACTUAL=$(sha256sum "$TMP/kyc-cli.tar.gz" 2>/dev/null | awk '{print $1}')
  [ -z "$ACTUAL" ] && ACTUAL=$(shasum -a 256 "$TMP/kyc-cli.tar.gz" 2>/dev/null | awk '{print $1}') || true
  if [ -n "$EXPECTED" ] && [ -n "$ACTUAL" ] && [ "$EXPECTED" != "$ACTUAL" ]; then
    echo "[install.sh] checksum mismatch! expected ${EXPECTED}, got ${ACTUAL}" >&2
    exit 1
  fi
fi

tar -xzf "$TMP/kyc-cli.tar.gz" -C "$TMP"
install -m 0755 "$TMP/${BIN_NAME}" "${INSTALL_DIR}/${BIN_NAME}"

echo "[install.sh] installed ${BIN_NAME} ${TAG} → ${INSTALL_DIR}/${BIN_NAME}"
case ":$PATH:" in
  *":${INSTALL_DIR}:"*) ;;
  *) echo "[install.sh] note: ${INSTALL_DIR} is not in your PATH; add it or move the binary" ;;
esac

"${INSTALL_DIR}/${BIN_NAME}" --version || true
