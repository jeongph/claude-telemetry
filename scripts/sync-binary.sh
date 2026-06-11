#!/bin/bash
# SessionStart 훅: 플러그인 버전과 바이너리 버전이 다르면 릴리즈 바이너리로 자동 동기화.
# - 바이너리 미설치(셋업 전)나 dev 빌드는 건드리지 않는다.
# - 다운로드는 백그라운드 — 세션 시작을 차단하지 않는다.
# - SessionStart stdout은 세션 컨텍스트에 주입되므로 아무것도 출력하지 않는다.
set -u

BIN="${HOME}/.claude/statusline/bin/claude-telemetry"
[ -x "$BIN" ] || exit 0

PLUGIN_VER=$(sed -n 's/.*"version"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' "${CLAUDE_PLUGIN_ROOT}/.claude-plugin/plugin.json" | head -1)
[ -n "$PLUGIN_VER" ] || exit 0

BIN_VER=$("$BIN" --version 2>/dev/null | sed 's/^claude-telemetry v\{0,1\}//')
[ -n "$BIN_VER" ] || exit 0
[ "$BIN_VER" = "dev" ] && exit 0
[ "$BIN_VER" = "$PLUGIN_VER" ] && exit 0

LOCK="${HOME}/.claude/statusline/.sync.lock"
# 10분 이상 된 stale 락 정리 (이전 실행이 비정상 종료한 경우)
if [ -d "$LOCK" ] && [ -n "$(find "$LOCK" -maxdepth 0 -mmin +10 2>/dev/null)" ]; then
    rmdir "$LOCK" 2>/dev/null
fi
mkdir "$LOCK" 2>/dev/null || exit 0

(
    trap 'rmdir "$LOCK" 2>/dev/null' EXIT
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    case "$ARCH" in
        x86_64) ARCH="amd64" ;;
        aarch64|arm64) ARCH="arm64" ;;
        *) exit 0 ;;
    esac
    NAME="claude-telemetry-${OS}-${ARCH}"
    BASE="https://github.com/jeongph/claude-telemetry/releases/download/v${PLUGIN_VER}"
    TMP=$(mktemp -p "$(dirname "$BIN")") || exit 0
    if curl -fsSL --max-time 60 "${BASE}/${NAME}" -o "$TMP"; then
        SUM=$(curl -fsSL --max-time 15 "${BASE}/checksums.txt" | awk -v n="$NAME" '$2 == n {print $1}')
        if command -v sha256sum >/dev/null 2>&1; then
            ACTUAL=$(sha256sum "$TMP" | awk '{print $1}')
        else
            ACTUAL=$(shasum -a 256 "$TMP" | awk '{print $1}')
        fi
        if [ -n "$SUM" ] && [ "$ACTUAL" = "$SUM" ]; then
            chmod +x "$TMP" && mv -f "$TMP" "$BIN"
        fi
    fi
    rm -f "$TMP"
) >/dev/null 2>&1 &

exit 0
