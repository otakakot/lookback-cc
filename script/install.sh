#!/bin/bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"

echo "==> Checking prerequisites..."
for cmd in go claude; do
  if ! command -v "$cmd" >/dev/null 2>&1; then
    echo "    Error: '$cmd' not found. Please install it first." >&2
    exit 1
  fi
done
echo "    go: $(go version | awk '{print $3}')"
echo "    claude: $(claude --version 2>/dev/null || echo 'ok')"

HOOKS_DIR="$HOME/.claude/hooks"
SETTINGS="$HOME/.claude/settings.json"
OUTPUT_DIR="$HOME/.claude/debrief"
BINARY="$HOOKS_DIR/debrief"

echo "==> Building debrief..."
go build -o "$BINARY" "$ROOT_DIR/cmd/debrief"
echo "    Installed: $BINARY"

echo "==> Creating output directory..."
mkdir -p "$OUTPUT_DIR"
echo "    Directory: $OUTPUT_DIR"

echo "==> Configuring SessionEnd hook..."
if [ -f "$SETTINGS" ]; then
  BACKUP="${SETTINGS}.bak.$(date +%Y%m%d%H%M%S)"
  cp "$SETTINGS" "$BACKUP"
  echo "    Backup: $BACKUP"
fi
if [ ! -f "$SETTINGS" ]; then
  cat > "$SETTINGS" <<'EOF'
{
  "hooks": {
    "SessionEnd": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "BINARY_PLACEHOLDER"
          }
        ]
      }
    ]
  }
}
EOF
  sed -i '' "s|BINARY_PLACEHOLDER|$BINARY|" "$SETTINGS"
  echo "    Created: $SETTINGS"
elif python3 -c "
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
  echo "    Already configured, skipping."
else
  python3 -c "
import json, sys

path = sys.argv[1]
binary = sys.argv[2]

with open(path) as f:
    cfg = json.load(f)

cfg.setdefault('hooks', {})
cfg['hooks'].setdefault('SessionEnd', [])
cfg['hooks']['SessionEnd'].append({
    'hooks': [{
        'type': 'command',
        'command': binary
    }]
})

with open(path, 'w') as f:
    json.dump(cfg, f, indent=2, ensure_ascii=False)
    f.write('\n')
" "$SETTINGS" "$BINARY"
  echo "    Added SessionEnd hook to: $SETTINGS"
fi

echo ""
echo "==> Installing cccall command..."
go install "$ROOT_DIR/cmd/cccall"
echo "    Installed: $(which cccall 2>/dev/null || echo '$(go env GOPATH)/bin/cccall')"

echo ""
echo "==> Installing report command..."
go install "$ROOT_DIR/cmd/report"
echo "    Installed: $(which report 2>/dev/null || echo '$(go env GOPATH)/bin/report')"

echo ""
echo "Done!"
echo "Summaries will be saved to: $OUTPUT_DIR"
