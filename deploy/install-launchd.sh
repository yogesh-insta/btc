#!/bin/bash
# Install btc as a macOS launchd service (auto-start on boot).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
PLIST_SRC="$ROOT/deploy/com.btc.daemon.plist"
PLIST_DST="$HOME/Library/LaunchAgents/com.btc.daemon.plist"

if [[ ! -f "$ROOT/.credentials" ]]; then
  echo "error: $ROOT/.credentials not found — copy .credentials.example and fill keys first"
  exit 1
fi

mkdir -p "$ROOT/logs"
sed "s|/Users/ym/mws26/btc|$ROOT|g" "$PLIST_SRC" > "$PLIST_DST"

launchctl unload "$PLIST_DST" 2>/dev/null || true
launchctl load "$PLIST_DST"

echo "Installed: $PLIST_DST"
echo "Status:    launchctl list | grep btc"
echo "Logs:      tail -f $ROOT/logs/daemon.stdout.log"
