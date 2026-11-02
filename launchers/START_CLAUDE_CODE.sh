#!/usr/bin/env bash
set -e
cd "$(dirname "$0")/.."
if ! command -v claude >/dev/null 2>&1; then
  echo "Claude Code was not found on PATH" >&2
  exit 1
fi
exec claude
