#!/bin/bash
set -euo pipefail

HOOKS_DIR="$HOME/.claude/hooks"
SETTINGS="$HOME/.claude/settings.json"
BINARY="$HOOKS_DIR/debrief"

echo "==> Removing debrief hook binary..."
if [ -f "$BINARY" ]; then
  rm "$BINARY"
  echo "    Removed: $BINARY"
else
  echo "    Not found, skipping."
fi

echo ""
echo "==> Removing SessionEnd hook from settings..."
if [ -f "$SETTINGS" ] && python3 -c "
import json, sys
with open(sys.argv[1]) as f:
    cfg = json.load(f)
hooks = cfg.get('hooks', {}).get('SessionEnd', [])
for rule in hooks:
    for h in rule.get('hooks', []):
        if 'debrief' in h.get('command', ''):
            sys.exit(0)
sys.exit(1)
" "$SETTINGS" 2>/dev/null; then
  BACKUP="${SETTINGS}.bak.$(date +%Y%m%d%H%M%S)"
  cp "$SETTINGS" "$BACKUP"
  echo "    Backup: $BACKUP"
  python3 -c "
import json, sys

path = sys.argv[1]

with open(path) as f:
    cfg = json.load(f)

session_end = cfg.get('hooks', {}).get('SessionEnd', [])
cfg['hooks']['SessionEnd'] = [
    rule for rule in session_end
    if not any('debrief' in h.get('command', '') for h in rule.get('hooks', []))
]

if not cfg['hooks']['SessionEnd']:
    del cfg['hooks']['SessionEnd']
if not cfg.get('hooks'):
    del cfg['hooks']

with open(path, 'w') as f:
    json.dump(cfg, f, indent=2, ensure_ascii=False)
    f.write('\n')
" "$SETTINGS"
  echo "    Removed debrief hook from: $SETTINGS"
else
  echo "    No debrief hook found, skipping."
fi

echo ""
echo "==> Removing cccall command..."
CCCALL="$(go env GOPATH)/bin/cccall"
if [ -f "$CCCALL" ]; then
  rm "$CCCALL"
  echo "    Removed: $CCCALL"
else
  echo "    Not found, skipping."
fi

echo ""
echo "==> Removing report command..."
REPORT="$(go env GOPATH)/bin/report"
if [ -f "$REPORT" ]; then
  rm "$REPORT"
  echo "    Removed: $REPORT"
else
  echo "    Not found, skipping."
fi

echo ""
echo "Done!"
echo "Note: debrief summaries in ~/.claude/debrief/ and reports in ~/.claude/report/ are preserved."
