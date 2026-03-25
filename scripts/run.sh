#!/bin/bash
# v2: Go 바이너리 우선, 없으면 v1(jq) 폴백
BIN="${HOME}/.claude/statusline/bin/claude-telemetry"
if [ -x "$BIN" ]; then
    exec "$BIN"
else
    SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    LEGACY="${SCRIPT_DIR}/run-legacy.sh"
    if [ -f "$LEGACY" ]; then
        exec bash "$LEGACY"
    else
        echo "⚠ Run /claude-telemetry:setup to install"
    fi
fi
