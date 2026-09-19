#!/usr/bin/env bash
set -euo pipefail
DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OS="$(uname -s 2>/dev/null || echo unknown)"
ARCH="$(uname -m 2>/dev/null || echo unknown)"
case "$OS" in
  MINGW*|MSYS*|CYGWIN*)
    case "$ARCH" in
      aarch64|arm64) BIN="$DIR/expertcall-windows-arm64.exe" ;;
      *) BIN="$DIR/expertcall-windows-amd64.exe" ;;
    esac
    ;;
  Darwin)
    case "$ARCH" in
      arm64|aarch64) BIN="$DIR/expertcall-darwin-arm64" ;;
      *) BIN="$DIR/expertcall-darwin-amd64" ;;
    esac
    ;;
  Linux)
    case "$ARCH" in
      arm64|aarch64) BIN="$DIR/expertcall-linux-arm64" ;;
      *) BIN="$DIR/expertcall-linux-amd64" ;;
    esac
    ;;
  *)
    echo "Unsupported operating system: $OS / $ARCH" >&2
    exit 2
    ;;
esac
if [[ ! -f "$BIN" ]]; then
  echo "Packaged helper not found: $BIN" >&2
  exit 2
fi
chmod +x "$BIN" 2>/dev/null || true
exec "$BIN" "$@"
