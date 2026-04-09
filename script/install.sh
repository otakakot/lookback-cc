#!/bin/bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"

HOOKS_DIR="$HOME/.claude/hooks"
SETTINGS="$HOME/.claude/settings.json"
OUTPUT_DIR="$HOME/.claude/debrief"
BINARY="$HOOKS_DIR/debrief"

echo "==> Checking prerequisites..."
for cmd in go claude; do
  if ! command -v "$cmd" >/dev/null 2>&1; then
    echo "    Error: '$cmd' not found. Please install it first." >&2
    exit 1
  fi
done
echo "    go: $(go version | awk '{print $3}')"
echo "    claude: $(claude --version 2>/dev/null || echo 'ok')"

echo ""
echo "==> Installing debrief hook..."
go build -o "$BINARY" "$ROOT_DIR/cmd/debrief"
echo "    Installed: $BINARY"

echo ""
echo "==> Creating output directory..."
mkdir -p "$OUTPUT_DIR"
echo "    Created: $OUTPUT_DIR"

echo ""
echo "==> Installing settings command..."
go install "$ROOT_DIR/cmd/settings"
echo "    Installed: $(which settings 2>/dev/null || echo '$(go env GOPATH)/bin/settings')"

echo ""
echo "==> Configuring SessionEnd hook..."
if [ -f "$SETTINGS" ]; then
  BACKUP="${SETTINGS}.bak.$(date +%Y%m%d%H%M%S)"
  cp "$SETTINGS" "$BACKUP"
  echo "    Backup: $BACKUP"
fi
RESULT=$(settings install "$BINARY")
case "$RESULT" in
  already_configured)
    echo "    Skipped: already configured" ;;
  installed)
    echo "    Configured: $SETTINGS" ;;
  *)
    echo "    Error: unexpected result: $RESULT" >&2; exit 1 ;;
esac

echo ""
echo "==> Installing summarize command..."
go install "$ROOT_DIR/cmd/summarize"
echo "    Installed: $(which summarize 2>/dev/null || echo '$(go env GOPATH)/bin/summarize')"

echo ""
echo "==> Installing report command..."
go install "$ROOT_DIR/cmd/report"
echo "    Installed: $(which report 2>/dev/null || echo '$(go env GOPATH)/bin/report')"

echo ""
echo "Done! Summaries will be saved to: $OUTPUT_DIR"
