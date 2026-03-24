#!/bin/bash
BIN="${HOME}/.claude/statusline/bin/claude-telemetry"
[ -x "$BIN" ] && exec "$BIN" || echo "⚠ Run /claude-telemetry:setup to install"
