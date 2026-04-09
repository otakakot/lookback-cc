#!/bin/bash
set -euo pipefail

HOOKS_DIR="$HOME/.claude/hooks"
SETTINGS="$HOME/.claude/settings.json"
BINARY="$HOOKS_DIR/debrief"
GOBIN="$(go env GOPATH)/bin"

echo "==> Removing debrief hook..."
if [ -f "$BINARY" ]; then
  rm "$BINARY"
  echo "    Removed: $BINARY"
else
  echo "    Skipped: not found"
fi

echo ""
echo "==> Removing SessionEnd hook from settings..."
if [ -f "$SETTINGS" ]; then
  BACKUP="${SETTINGS}.bak.$(date +%Y%m%d%H%M%S)"
  cp "$SETTINGS" "$BACKUP"
  echo "    Backup: $BACKUP"
fi
RESULT=$(settings uninstall "$BINARY")
case "$RESULT" in
  not_found)
    echo "    Skipped: not found" ;;
  uninstalled)
    echo "    Removed: debrief hook from $SETTINGS" ;;
  *)
    echo "    Error: unexpected result: $RESULT" >&2; exit 1 ;;
esac

echo ""
echo "==> Removing summarize command..."
if [ -f "$GOBIN/summarize" ]; then
  rm "$GOBIN/summarize"
  echo "    Removed: $GOBIN/summarize"
else
  echo "    Skipped: not found"
fi

echo ""
echo "==> Removing report command..."
if [ -f "$GOBIN/report" ]; then
  rm "$GOBIN/report"
  echo "    Removed: $GOBIN/report"
else
  echo "    Skipped: not found"
fi

echo ""
echo "==> Removing settings command..."
if [ -f "$GOBIN/settings" ]; then
  rm "$GOBIN/settings"
  echo "    Removed: $GOBIN/settings"
else
  echo "    Skipped: not found"
fi

echo ""
echo "Done! Summaries in ~/.claude/debrief/ and reports in ~/.claude/report/ are preserved."
