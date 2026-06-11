#!/bin/bash
# v2: Go 바이너리 우선, 없으면 v1(jq) 폴백
SL_DIR="${HOME}/.claude/statusline"
BIN="${SL_DIR}/bin/claude-telemetry"
MARKER="${SL_DIR}/.managed-by-plugin"
STAMP="${SL_DIR}/.removal-detected"

# 플러그인 설치본(setup이 마커 기록) 한정: 플러그인 제거 감지 시 자가 정리.
# 업데이트 중 캐시가 일시적으로 비는 오탐을 막기 위해 60초 유예 후 정리한다.
if [ -f "$MARKER" ]; then
    if ls "${HOME}"/.claude/plugins/cache/*/claude-telemetry >/dev/null 2>&1; then
        rm -f "$STAMP"
    else
        NOW=$(date +%s)
        FIRST=$(cat "$STAMP" 2>/dev/null)
        if [ -z "$FIRST" ]; then
            echo "$NOW" > "$STAMP"
            FIRST=$NOW
        fi
        if [ $((NOW - FIRST)) -ge 60 ] && [ -x "$BIN" ]; then
            "$BIN" --self-uninstall >/dev/null 2>&1
            exit 0
        fi
        echo "⚠ claude-telemetry: plugin removed — cleaning up automatically (reinstall to keep)"
        exit 0
    fi
fi

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
