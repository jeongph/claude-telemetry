# v2.4.0 바이너리 자동 동기화 및 제거 시 자가 정리 구현 계획

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 플러그인 업데이트 시 바이너리가 자동으로 동기화되고(SessionStart 훅), 플러그인 제거 시 statusline이 스스로 정리되도록(run.sh 감지 + `--self-uninstall`) 만든다.

**Architecture:** (1) SessionStart 커맨드 훅이 플러그인 버전과 바이너리 버전을 비교해 불일치 시 버전 핀 릴리즈를 백그라운드 다운로드+sha256 검증 후 원자적 교체. (2) setup이 마커 파일을 남기고, run.sh가 매 렌더링마다 플러그인 캐시 존재를 확인 — 60초 이상 부재 시 Go 바이너리의 `--self-uninstall`을 호출해 settings.json의 statusLine 키 제거(백업+원자적 쓰기)와 파일 정리를 수행. JSON 조작은 전부 테스트 가능한 Go 코드가 담당한다.

**Tech Stack:** Go 1.22 (encoding/json), bash (감지·다운로드), Claude Code plugin hooks (SessionStart)

---

## 배경 (Why)

- **바이너리-플러그인 버전 분리 문제**: `/plugin`의 "Update now"는 플러그인 파일만 갱신하고 렌더링 바이너리는 그대로라, 사용자가 "2.3.0인데 effort가 안 뜬다"는 혼란 발생 (실사용 보고). CC에 플러그인 업데이트 이벤트 훅은 없지만 SessionStart 훅으로 다음 세션 시작 시 동기화 가능.
- **제거 시 유령 UI**: 플러그인을 제거해도 settings.json의 statusLine 등록과 바이너리는 남아 UI가 계속 표시됨 (실사용 보고). 제거 시점 훅이 없으므로, 제거 후에도 유일하게 계속 실행되는 run.sh가 감지·정리를 담당해야 함. 안내줄만 남기는 방안은 그 안내줄을 지울 방법이 없어 기각 — 완전 자가 정리로 결정 (사용자 합의).
- **jq 리스크 회피**: settings.json 조작을 bash+jq가 아닌 Go 바이너리(`--self-uninstall`)가 수행 — 표준 라이브러리 JSON 처리, 백업, 원자적 쓰기, 단위 테스트 가능.

## 실측 확인 사항 (2026-06-11)

- 플러그인 캐시 구조: `~/.claude/plugins/cache/<마켓플레이스명>/claude-telemetry/<버전>/` — 마켓플레이스명이 다를 수 있어 글롭 `*/claude-telemetry`로 감지
- 바이너리 버전 출력: `claude-telemetry v2.3.0` (v 접두사 포함), plugin.json은 `"2.3.0"` (v 없음) → 정규화 필요
- 릴리즈 checksums.txt 형식: `<sha256>  claude-telemetry-<os>-<arch>` (sha256sum 출력)
- macOS에는 sha256sum이 없음 → `shasum -a 256` 폴백 필요
- 훅 형식: plugin `hooks/hooks.json`은 `{"hooks": {"SessionStart": [...]}}` 래퍼 형식, `${CLAUDE_PLUGIN_ROOT}` 사용. SessionStart stdout은 세션 컨텍스트에 주입되므로 **침묵 필수**. 훅은 세션 시작 시 로드 — 변경 반영은 다음 세션부터.

## setup/remove 크로스체크 결과 (이번 범위에 반영)

| 항목 | 현재 | 문제 | 조치 |
|------|------|------|------|
| setup Step 2 | `releases/latest` 다운로드, 체크섬 없음 | 자동 동기화 훅은 버전 핀인데 setup은 latest — 정책 불일치. 무결성 미검증 | 플러그인 버전 핀 다운로드 + sha256 검증 (Task 6) |
| setup Step 2 | 기설치 시 업데이트 여부 질문 | 핀 정책에선 질문 무의미 — 버전 다르면 동기화가 정답 | 같으면 스킵, 다르면 질문 없이 동기화 (Task 6) |
| setup Step 6 | 마커 파일 없음 | 자가 정리가 setup 설치본임을 식별 못 함 | `.managed-by-plugin` 마커 기록 (Task 6) |
| remove "Remove all" | Claude가 Edit 도구로 settings.json 수동 편집 + python3 검증 | 수작업 JSON 편집은 자가 정리와 코드 경로 이원화, 오편집 위험 | `--self-uninstall` 단일 경로로 교체, 실패 시에만 수동 폴백 (Task 7) |

## 파일 구조

- Create: `internal/selfuninstall/selfuninstall.go` — settings.json statusLine 제거 + 파일 정리 (단일 책임)
- Create: `internal/selfuninstall/selfuninstall_test.go`
- Modify: `cmd/claude-telemetry/main.go` — `--self-uninstall` 플래그
- Modify: `scripts/run.sh` — 플러그인 부재 감지 (마커 기반, 60초 유예)
- Create: `hooks/hooks.json` — SessionStart 훅 등록
- Create: `scripts/sync-binary.sh` — 버전 동기화 (백그라운드 다운로드 + 체크섬)
- Modify: `commands/setup.md`, `commands/remove.md`, `README.md`
- Modify: `.claude-plugin/plugin.json`, `.claude-plugin/marketplace.json` — 2.4.0

**스코프 제외:** 마커 없는 기존 사용자(셋업 미재실행)의 자가 정리 — 마커가 없으면 로직 미작동(의도), setup 재실행 시부터 적용. 수동 설치 사용자는 마커가 없어 영원히 미적용(의도).

---

### Task 1: feature 브랜치 생성 및 계획 문서 커밋

- [ ] **Step 1: 브랜치 생성**

```bash
cd /home/jeonguk/dev/lighthouse/repositories/claude-telemetry
git checkout main && git pull origin main
git checkout -b feature/v2-4-0-auto-sync-cleanup
```

- [ ] **Step 2: 계획 문서 커밋**

```bash
git add docs/superpowers/plans/2026-06-11-v2-4-0-auto-sync-cleanup.md
git commit -m "docs: v2.4.0 자동 동기화·자가 정리 계획 추가"
```

---

### Task 2: internal/selfuninstall 패키지 (TDD)

**Files:**
- Create: `internal/selfuninstall/selfuninstall.go`
- Create: `internal/selfuninstall/selfuninstall_test.go`

- [ ] **Step 1: 실패하는 테스트 작성** — `internal/selfuninstall/selfuninstall_test.go`:

```go
package selfuninstall

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("파일 쓰기 실패: %v", err)
	}
}

func TestRemoveStatusLine(t *testing.T) {
	dir := t.TempDir()
	settings := filepath.Join(dir, "settings.json")
	writeFile(t, settings, `{
  "model": "opus",
  "statusLine": {"type": "command", "command": "bash /x/run.sh"},
  "permissions": {"allow": ["Bash(ls:*)"]}
}`)

	if err := RemoveStatusLine(settings); err != nil {
		t.Fatalf("RemoveStatusLine 실패: %v", err)
	}

	data, _ := os.ReadFile(settings)
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("결과가 유효한 JSON이 아님: %v", err)
	}
	if _, ok := m["statusLine"]; ok {
		t.Error("statusLine 키가 제거되지 않음")
	}
	if _, ok := m["model"]; !ok {
		t.Error("model 키가 보존되지 않음")
	}
	if _, ok := m["permissions"]; !ok {
		t.Error("permissions 키가 보존되지 않음")
	}

	// 백업 존재 + 원본 내용 보존
	bak, err := os.ReadFile(settings + ".claude-telemetry.bak")
	if err != nil {
		t.Fatalf("백업 파일 없음: %v", err)
	}
	if !strings.Contains(string(bak), "statusLine") {
		t.Error("백업에 원본 statusLine이 없음")
	}
}

func TestRemoveStatusLineIdempotent(t *testing.T) {
	dir := t.TempDir()
	settings := filepath.Join(dir, "settings.json")
	original := `{"model": "opus"}`
	writeFile(t, settings, original)

	if err := RemoveStatusLine(settings); err != nil {
		t.Fatalf("statusLine 없는 파일에서 에러: %v", err)
	}
	data, _ := os.ReadFile(settings)
	if string(data) != original {
		t.Error("statusLine 없는 파일이 변경됨 — no-op이어야 함")
	}
	if _, err := os.Stat(settings + ".claude-telemetry.bak"); err == nil {
		t.Error("no-op인데 백업이 생성됨")
	}
}

func TestRemoveStatusLineMissingFile(t *testing.T) {
	dir := t.TempDir()
	if err := RemoveStatusLine(filepath.Join(dir, "settings.json")); err != nil {
		t.Errorf("파일 없음은 에러가 아니어야 함: %v", err)
	}
}

func TestRemoveStatusLineInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	settings := filepath.Join(dir, "settings.json")
	broken := `{invalid json`
	writeFile(t, settings, broken)

	if err := RemoveStatusLine(settings); err == nil {
		t.Error("깨진 JSON에서 에러가 반환되어야 함")
	}
	data, _ := os.ReadFile(settings)
	if string(data) != broken {
		t.Error("깨진 JSON 파일을 건드리면 안 됨")
	}
}

func TestCleanupFiles(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "bin"), 0o755)
	os.MkdirAll(filepath.Join(dir, "cache"), 0o755)
	writeFile(t, filepath.Join(dir, "bin", "claude-telemetry"), "binary")
	writeFile(t, filepath.Join(dir, "config.json"), "{}")
	writeFile(t, filepath.Join(dir, ".managed-by-plugin"), "")
	writeFile(t, filepath.Join(dir, ".removal-detected"), "123")
	writeFile(t, filepath.Join(dir, "run.sh"), "#!/bin/bash\nexec something")

	if err := CleanupFiles(dir); err != nil {
		t.Fatalf("CleanupFiles 실패: %v", err)
	}

	for _, gone := range []string{"bin", "cache", "config.json", ".managed-by-plugin", ".removal-detected"} {
		if _, err := os.Stat(filepath.Join(dir, gone)); err == nil {
			t.Errorf("%s가 제거되지 않음", gone)
		}
	}
	// run.sh는 삭제가 아니라 무해화 스텁으로 교체 (현재 세션이 계속 호출하므로)
	stub, err := os.ReadFile(filepath.Join(dir, "run.sh"))
	if err != nil {
		t.Fatal("run.sh 스텁이 없음")
	}
	if !strings.Contains(string(stub), "exit 0") {
		t.Errorf("run.sh가 무해화 스텁이 아님: %q", string(stub))
	}
}

func TestRunWithLock(t *testing.T) {
	claudeDir := t.TempDir()
	slDir := filepath.Join(claudeDir, "statusline")
	os.MkdirAll(slDir, 0o755)
	settings := filepath.Join(claudeDir, "settings.json")
	writeFile(t, settings, `{"statusLine": {"type": "command"}}`)
	// 락 선점 → Run은 조용히 no-op
	writeFile(t, filepath.Join(slDir, ".uninstall.lock"), "")

	if err := Run(claudeDir, slDir); err != nil {
		t.Fatalf("락 존재 시 에러 없이 종료해야 함: %v", err)
	}
	data, _ := os.ReadFile(settings)
	if !strings.Contains(string(data), "statusLine") {
		t.Error("락 존재 시 settings를 건드리면 안 됨")
	}
}

func TestRunFull(t *testing.T) {
	claudeDir := t.TempDir()
	slDir := filepath.Join(claudeDir, "statusline")
	os.MkdirAll(filepath.Join(slDir, "bin"), 0o755)
	settings := filepath.Join(claudeDir, "settings.json")
	writeFile(t, settings, `{"model": "opus", "statusLine": {"type": "command"}}`)
	writeFile(t, filepath.Join(slDir, "run.sh"), "#!/bin/bash\nexec x")

	if err := Run(claudeDir, slDir); err != nil {
		t.Fatalf("Run 실패: %v", err)
	}
	data, _ := os.ReadFile(settings)
	if strings.Contains(string(data), "statusLine") {
		t.Error("statusLine이 제거되지 않음")
	}
	if _, err := os.Stat(filepath.Join(slDir, "bin")); err == nil {
		t.Error("bin 디렉토리가 제거되지 않음")
	}
}
```

- [ ] **Step 2: 실패 확인**

Run: `go test ./internal/selfuninstall/ -v`
Expected: FAIL — 컴파일 에러 (패키지 없음)

- [ ] **Step 3: 구현** — `internal/selfuninstall/selfuninstall.go`:

```go
// Package selfuninstall은 플러그인 제거 시 statusline 구성을 안전하게 정리한다.
// settings.json 조작을 bash가 아닌 테스트 가능한 Go 코드로 수행하기 위한 패키지다.
package selfuninstall

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// runShStub은 정리 후 run.sh에 남기는 무해화 스텁이다.
// 현재 떠 있는 세션이 statusline 명령을 계속 호출하므로 빈 출력으로 응답해야 한다.
// 다음 세션부터는 settings.json에 등록이 없어 호출되지 않는다.
const runShStub = "#!/bin/bash\n# claude-telemetry removed — this stub is safe to delete\nexit 0\n"

// RemoveStatusLine은 settings.json에서 statusLine 키만 제거한다.
// 다른 키는 보존하며, 원본은 settings.json.claude-telemetry.bak으로 백업한다.
// statusLine 키가 없거나 파일이 없으면 아무것도 하지 않는다 (멱등).
// 깨진 JSON이면 파일을 건드리지 않고 에러를 반환한다.
func RemoveStatusLine(settingsPath string) error {
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return err
	}
	var settings map[string]json.RawMessage
	if err := json.Unmarshal(data, &settings); err != nil {
		return fmt.Errorf("settings.json 파싱 실패 (수정하지 않음): %w", err)
	}
	if _, ok := settings["statusLine"]; !ok {
		return nil
	}
	delete(settings, "statusLine")

	out, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	out = append(out, '\n')

	mode := fs.FileMode(0o600)
	if fi, err := os.Stat(settingsPath); err == nil {
		mode = fi.Mode()
	}
	if err := os.WriteFile(settingsPath+".claude-telemetry.bak", data, mode); err != nil {
		return err
	}
	tmp := settingsPath + ".claude-telemetry.tmp"
	if err := os.WriteFile(tmp, out, mode); err != nil {
		return err
	}
	return os.Rename(tmp, settingsPath)
}

// CleanupFiles는 statusline 디렉토리의 구성 요소를 정리하고 run.sh를 스텁으로 교체한다.
// run.sh 자체는 삭제하지 않는다 — 현재 세션이 계속 호출하기 때문.
func CleanupFiles(statuslineDir string) error {
	for _, f := range []string{"config.json", ".managed-by-plugin", ".removal-detected"} {
		if err := os.Remove(filepath.Join(statuslineDir, f)); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return err
		}
	}
	for _, d := range []string{"bin", "cache"} {
		if err := os.RemoveAll(filepath.Join(statuslineDir, d)); err != nil {
			return err
		}
	}
	runSh := filepath.Join(statuslineDir, "run.sh")
	if _, err := os.Stat(runSh); err == nil {
		if err := os.WriteFile(runSh, []byte(runShStub), 0o755); err != nil {
			return err
		}
	}
	return nil
}

// Run은 self-uninstall 전체 절차를 수행한다.
// 락 파일(O_EXCL)로 동시 실행을 방지하며, 락 선점 실패는 에러가 아니다 (이미 진행 중).
func Run(claudeDir, statuslineDir string) error {
	lock := filepath.Join(statuslineDir, ".uninstall.lock")
	f, err := os.OpenFile(lock, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return nil
	}
	f.Close()

	if err := RemoveStatusLine(filepath.Join(claudeDir, "settings.json")); err != nil {
		os.Remove(lock)
		return err
	}
	if err := CleanupFiles(statuslineDir); err != nil {
		os.Remove(lock)
		return err
	}
	return os.Remove(lock)
}
```

- [ ] **Step 4: 테스트 통과 확인**

Run: `go test ./internal/selfuninstall/ -v && go test ./... -count=1`
Expected: 전체 PASS

- [ ] **Step 5: 커밋**

```bash
git add internal/selfuninstall/
git commit -m "feat(selfuninstall): settings.json statusLine 제거 및 파일 정리 패키지 추가

- 플러그인 제거 시 자가 정리를 위한 테스트 가능한 Go 경로
- 백업 + 원자적 쓰기, 깨진 JSON 보호, 락 기반 동시 실행 방지, 멱등"
```

---

### Task 3: main.go에 --self-uninstall 플래그 연결

**Files:**
- Modify: `cmd/claude-telemetry/main.go`

- [ ] **Step 1: 플래그 추가** — main()의 인자 루프(`switch arg`)에 케이스 추가:

```go
		case "--self-uninstall":
			home, _ := os.UserHomeDir()
			claudeDir := filepath.Join(home, ".claude")
			slDir := os.Getenv("CLAUDE_STATUSLINE_CONFIG")
			if slDir == "" {
				slDir = filepath.Join(claudeDir, "statusline")
			}
			if err := selfuninstall.Run(claudeDir, slDir); err != nil {
				fmt.Fprintln(os.Stderr, "self-uninstall:", err)
				os.Exit(1)
			}
			return
```

import에 `"github.com/jeongph/claude-telemetry/internal/selfuninstall"` 추가.

- [ ] **Step 2: 빌드·수동 확인**

```bash
go build ./... && go test ./... -count=1
go build -o /tmp/ct-su ./cmd/claude-telemetry
HOME=$(mktemp -d) && mkdir -p $HOME/.claude/statusline/bin && echo '{"statusLine":{"a":1},"model":"x"}' > $HOME/.claude/settings.json && HOME=$HOME /tmp/ct-su --self-uninstall && cat $HOME/.claude/settings.json
```

Expected: 출력 JSON에 statusLine 없음, model 보존. (주의: 위 HOME 변수는 서브셸용 임시 디렉토리 — 실제 홈 보호)

실행 후 `unset HOME` 또는 새 셸 확인 — 실제로는 `env HOME=$(mktemp -d) bash -c '...'` 형태로 실행하여 현재 셸 HOME을 오염시키지 말 것:

```bash
T=$(mktemp -d) && mkdir -p "$T/.claude/statusline/bin" && echo '{"statusLine":{"a":1},"model":"x"}' > "$T/.claude/settings.json" && env HOME="$T" /tmp/ct-su --self-uninstall && cat "$T/.claude/settings.json" && rm -rf "$T"
```

- [ ] **Step 3: 커밋**

```bash
git add cmd/claude-telemetry/main.go
git commit -m "feat(cli): --self-uninstall 플래그 추가"
```

---

### Task 4: run.sh 플러그인 부재 감지

**Files:**
- Modify: `scripts/run.sh` (전체 교체)

- [ ] **Step 1: run.sh를 다음 내용으로 교체**

```bash
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
```

- [ ] **Step 2: 가짜 HOME으로 시나리오 검증** (실제 홈 비오염)

```bash
go build -o /tmp/ct-su ./cmd/claude-telemetry
T=$(mktemp -d)
mkdir -p "$T/.claude/statusline/bin" "$T/.claude/plugins/cache/mkt/claude-telemetry/2.4.0"
cp /tmp/ct-su "$T/.claude/statusline/bin/claude-telemetry"
echo '{"statusLine":{"a":1},"model":"x"}' > "$T/.claude/settings.json"
touch "$T/.claude/statusline/.managed-by-plugin"

# 1) 플러그인 존재 → 정상 렌더링 (echo 입력으로 모델명 출력 확인)
echo '{"model":{"id":"m","display_name":"M"}}' | env HOME="$T" bash scripts/run.sh

# 2) 플러그인 제거 → 안내줄
rm -rf "$T/.claude/plugins/cache/mkt"
env HOME="$T" bash scripts/run.sh </dev/null   # → "⚠ ... cleaning up" + STAMP 생성 확인
cat "$T/.claude/statusline/.removal-detected"

# 3) 60초 경과 시뮬레이션 → 자가 정리
echo 1 > "$T/.claude/statusline/.removal-detected"
env HOME="$T" bash scripts/run.sh </dev/null
grep statusLine "$T/.claude/settings.json" || echo "CLEANED"
cat "$T/.claude/statusline/run.sh" 2>/dev/null   # 스텁인지? (주의: 이 테스트의 run.sh는 레포 파일이라 스텁 교체 대상은 $T가 아님 — CleanupFiles는 $T/.claude/statusline/run.sh를 대상으로 함. $T에 run.sh를 복사해두고 검증)

# 4) 오탐 복구: 플러그인 재생성 후 STAMP 정리 확인
rm -rf "$T"
```

Expected: 1) 모델명 출력 2) 안내줄 + STAMP 파일 3) "CLEANED" 4) 정상.
주의: 시나리오 3 전에 `cp scripts/run.sh "$T/.claude/statusline/run.sh"` 해두고 스텁 교체를 확인하라.

- [ ] **Step 3: 커밋**

```bash
git add scripts/run.sh
git commit -m "feat(run): 플러그인 제거 감지 시 자가 정리 트리거 추가

- setup이 남긴 마커가 있을 때만 작동 (수동 설치 사용자 미영향)
- 60초 유예로 플러그인 업데이트 중 캐시 공백 오탐 방지
- 감지~정리 사이에는 흐린 안내줄 표시 후 스스로 사라짐"
```

---

### Task 5: SessionStart 훅 — 바이너리 자동 동기화

**Files:**
- Create: `hooks/hooks.json`
- Create: `scripts/sync-binary.sh`

- [ ] **Step 1: hooks/hooks.json 생성**

```json
{
  "description": "플러그인-바이너리 버전 자동 동기화 (세션 시작 시)",
  "hooks": {
    "SessionStart": [
      {
        "matcher": "*",
        "hooks": [
          {
            "type": "command",
            "command": "bash ${CLAUDE_PLUGIN_ROOT}/scripts/sync-binary.sh",
            "timeout": 15
          }
        ]
      }
    ]
  }
}
```

- [ ] **Step 2: scripts/sync-binary.sh 생성**

```bash
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
    esac
    NAME="claude-telemetry-${OS}-${ARCH}"
    BASE="https://github.com/jeongph/claude-telemetry/releases/download/v${PLUGIN_VER}"
    TMP=$(mktemp) || exit 0
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
```

- [ ] **Step 3: 실 다운로드 검증 (v2.3.0 릴리즈 자산 활용, 가짜 HOME)**

```bash
chmod +x scripts/sync-binary.sh
T=$(mktemp -d)
mkdir -p "$T/.claude/statusline/bin" "$T/plugin/.claude-plugin"
go build -o "$T/.claude/statusline/bin/claude-telemetry" ./cmd/claude-telemetry  # dev 빌드

# 1) dev 가드: dev 빌드는 건드리지 않음
env HOME="$T" CLAUDE_PLUGIN_ROOT="$T/plugin" bash -c 'echo "{\"version\": \"2.3.0\"}" > "$CLAUDE_PLUGIN_ROOT/.claude-plugin/plugin.json"; bash scripts/sync-binary.sh'
"$T/.claude/statusline/bin/claude-telemetry" --version   # → "claude-telemetry dev" 유지

# 2) 실제 동기화: 구버전 흉내(버전 문자열만 다른 가짜 스크립트) → v2.3.0 실 다운로드+체크섬 검증
printf '#!/bin/bash\necho "claude-telemetry v0.0.1"\n' > "$T/.claude/statusline/bin/claude-telemetry"
chmod +x "$T/.claude/statusline/bin/claude-telemetry"
env HOME="$T" CLAUDE_PLUGIN_ROOT="$T/plugin" bash scripts/sync-binary.sh
sleep 8   # 백그라운드 다운로드 대기
"$T/.claude/statusline/bin/claude-telemetry" --version   # → "claude-telemetry v2.3.0"
rm -rf "$T"
```

Expected: 1) dev 유지 2) v2.3.0으로 교체됨.

- [ ] **Step 4: 커밋**

```bash
git add hooks/hooks.json scripts/sync-binary.sh
git commit -m "feat(hooks): SessionStart 바이너리 자동 동기화 훅 추가

- 플러그인 버전 핀 다운로드 + sha256 검증 + 원자적 교체
- 백그라운드 실행으로 세션 시작 비차단, dev 빌드 보호, stale 락 정리"
```

---

### Task 6: setup.md 갱신 (크로스체크 반영)

**Files:**
- Modify: `commands/setup.md`

- [ ] **Step 1: Step 2의 스크립트 블록 전체 교체** (버전 핀 + sha256 + 질문 제거):

```bash
PLUGIN_VER=$(sed -n 's/.*"version"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' "${CLAUDE_PLUGIN_ROOT}/.claude-plugin/plugin.json" | head -1)
EXISTING_VER=$(~/.claude/statusline/bin/claude-telemetry --version 2>/dev/null | sed 's/^claude-telemetry v\{0,1\}//')
if [ "$EXISTING_VER" = "$PLUGIN_VER" ]; then
  echo "UP_TO_DATE: v$EXISTING_VER"
else
  OS=$(uname -s | tr '[:upper:]' '[:lower:]')
  ARCH=$(uname -m)
  case "$ARCH" in
    x86_64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
  esac
  NAME="claude-telemetry-${OS}-${ARCH}"
  BASE="https://github.com/jeongph/claude-telemetry/releases/download/v${PLUGIN_VER}"
  mkdir -p ~/.claude/statusline/bin
  TMP=$(mktemp)
  echo "Downloading: ${BASE}/${NAME}"
  if curl -fsSL "${BASE}/${NAME}" -o "$TMP"; then
    SUM=$(curl -fsSL "${BASE}/checksums.txt" | awk -v n="$NAME" '$2 == n {print $1}')
    if command -v sha256sum >/dev/null 2>&1; then ACTUAL=$(sha256sum "$TMP" | awk '{print $1}'); else ACTUAL=$(shasum -a 256 "$TMP" | awk '{print $1}'); fi
    if [ -n "$SUM" ] && [ "$ACTUAL" = "$SUM" ]; then
      chmod +x "$TMP" && mv -f "$TMP" ~/.claude/statusline/bin/claude-telemetry && \
      ~/.claude/statusline/bin/claude-telemetry --version && echo "SUCCESS"
    else
      rm -f "$TMP"; echo "CHECKSUM_FAILED"
    fi
  else
    rm -f "$TMP"; echo "DOWNLOAD_FAILED"
  fi
fi
```

출력 처리 지시문도 교체:
- `UP_TO_DATE` → 사용자에게 현재 버전 안내 후 Step 3으로
- `SUCCESS` → Step 3으로
- `CHECKSUM_FAILED` / `DOWNLOAD_FAILED` → 실패 안내 후 중단

(기존 "INSTALLED → 업데이트 여부 질문" 분기 삭제 — 버전 핀 정책에서는 플러그인 버전과 다르면 동기화가 정답)

- [ ] **Step 2: Step 6의 복사 명령에 마커 추가**

기존 bash 블록:
```bash
mkdir -p ~/.claude/statusline
cp "${CLAUDE_PLUGIN_ROOT}/scripts/run.sh" ~/.claude/statusline/run.sh
```
를 다음으로 교체:
```bash
mkdir -p ~/.claude/statusline
cp "${CLAUDE_PLUGIN_ROOT}/scripts/run.sh" ~/.claude/statusline/run.sh
touch ~/.claude/statusline/.managed-by-plugin
```
해당 항목 설명에 한 줄 추가: "(the marker file enables automatic cleanup when the plugin is uninstalled)"

- [ ] **Step 3: Step 7 완료 메시지에 한 줄 추가**

`Restart Claude Code to apply.` 뒤에:
```
From now on the binary stays in sync with the plugin automatically (checked at session start).
```

- [ ] **Step 4: 커밋**

```bash
git add commands/setup.md
git commit -m "docs(setup): 버전 핀 다운로드·sha256 검증·자가 정리 마커 반영

- latest 대신 플러그인 버전 핀으로 훅 동기화 정책과 일치
- 기설치 업데이트 질문 제거 (버전 다르면 동기화가 정답)"
```

---

### Task 7: remove.md 갱신 (크로스체크 반영)

**Files:**
- Modify: `commands/remove.md`

- [ ] **Step 1: "Remove all" 절차 교체**

기존 1~5 (rm 바이너리 → rm -rf → Edit로 statusLine 제거 → python3 검증)를 다음으로 교체:

````markdown
### "Remove all"

1. Run the tested cleanup path in the binary (removes statusLine from settings.json with backup, deletes config/binary/marker, neutralizes run.sh):
   ```bash
   ~/.claude/statusline/bin/claude-telemetry --self-uninstall && echo "CLEANED" || echo "FALLBACK"
   ```
2. If `CLEANED` → proceed to Step 4. A backup was saved at `~/.claude/settings.json.claude-telemetry.bak`.
3. If `FALLBACK` (binary missing or failed) → do it manually:
   1. Remove files:
      ```bash
      rm -rf ~/.claude/statusline/
      ```
   2. Read `~/.claude/settings.json`
   3. Remove the `"statusLine"` entry from the JSON using Edit tool
   4. Verify settings.json is still valid JSON:
      ```bash
      python3 -c "import json; json.load(open('$HOME/.claude/settings.json'))" 2>/dev/null || echo "INVALID"
      ```
      If invalid, warn the user and suggest manual fix.
````

- [ ] **Step 2: Step 4 완료 메시지에 한 줄 추가**

기존 템플릿 끝에:
```
A harmless stub may remain at ~/.claude/statusline/run.sh — you can delete the directory after restarting.
```

- [ ] **Step 3: 커밋**

```bash
git add commands/remove.md
git commit -m "docs(remove): --self-uninstall 단일 경로로 교체

- 수동 JSON 편집 대신 테스트된 Go 코드 경로 사용, 실패 시에만 수동 폴백"
```

---

### Task 8: README 갱신

**Files:**
- Modify: `README.md`

- [ ] **Step 1: Features 목록에 추가** (적절한 위치, 예: NO_COLOR 항목 위):

```markdown
- **Auto binary sync** — a SessionStart hook keeps the binary matched to the plugin version (pinned download + sha256 verification)
- **Self-cleanup on uninstall** — if you uninstall the plugin, the status line removes its own settings entry and files within a minute (setup-managed installs only)
```

- [ ] **Step 2: Removal 섹션 교체**

기존:
````markdown
## Removal

```
/claude-telemetry:remove
```
````
를 다음으로 교체:
````markdown
## Removal

```
/claude-telemetry:remove
```

If you uninstall the plugin without running remove first, the status line detects the missing plugin and cleans itself up automatically within about a minute (settings entry removed from the next session). This applies to installs managed by `/claude-telemetry:setup`; manual installs are never touched.
````

- [ ] **Step 3: "Upgrading from v1" 섹션 위에 "Upgrading" 섹션 추가**

````markdown
## Upgrading

- **Plugin users (v2.4.0+):** update the plugin (`/plugin` → Update), then restart Claude Code. The SessionStart hook syncs the binary to the plugin version automatically.
- **Plugin users (older):** run `/claude-telemetry:setup` once after updating the plugin — it downloads the matching binary and migrates your settings to the version-independent launcher path.
- **Manual installs:** re-run the curl command from Manual setup; the binary is all that matters.
````

- [ ] **Step 4: 커밋**

```bash
git add README.md
git commit -m "docs: 자동 동기화·자가 정리·업그레이드 안내 추가"
```

---

### Task 9: 로컬 통합 검증 (가짜 HOME 시나리오)

**Files:** 없음

- [ ] **Step 1: 전체 테스트 + 빌드**

```bash
go test ./... -count=1 && go vet ./... && gofmt -l internal/ cmd/ | (! grep .)
```

- [ ] **Step 2: Task 4 Step 2의 4개 시나리오 + Task 5 Step 3의 2개 시나리오 재실행** — 모두 기대대로인지 확인

- [ ] **Step 3: 이 머신 실환경에는 마커를 만들지 않음에 유의** (이 머신은 setup 재실행 전까지 자가 정리 비활성 — 의도된 동작). 로컬 바이너리는 dev 빌드로 교체하지 않고 v2.3.0 릴리즈본 유지 (sync 훅 dev 가드 테스트는 가짜 HOME에서 이미 수행).

---

### Task 10: v2.4.0 버전 bump

**Files:**
- Modify: `.claude-plugin/plugin.json` (`"version": "2.3.0"` → `"2.4.0"`)
- Modify: `.claude-plugin/marketplace.json` (`"version": "2.3.0"` → `"2.4.0"`)

- [ ] **Step 1: 두 파일 수정 후 커밋**

```bash
git add .claude-plugin/
git commit -m "chore: v2.4.0 버전 갱신"
```

---

### ◆ 승인 게이트 — 여기서 정지

사용자에게 검증 결과를 보고하고 승인을 받은 뒤에만 이후 태스크 진행 (푸시·PR·릴리즈는 외부 반영).

---

### Task 11: PR·머지·릴리즈

- [ ] **Step 1: 푸시 및 PR 생성·머지**

```bash
git push -u origin feature/v2-4-0-auto-sync-cleanup
gh pr create --title "feat: 바이너리 자동 동기화 및 제거 시 자가 정리" --body "(요약·검증 포함, 🤖 푸터)"
gh pr merge --merge
```

- [ ] **Step 2: 태그 전 원격 태그 확인 후 태그 푸시** (v2.3.0 교훈)

```bash
git checkout main && git pull origin main && git fetch --tags
git ls-remote --tags origin | grep v2.4.0 || (git tag v2.4.0 && git push origin v2.4.0)
```

- [ ] **Step 3: release.yml 완료 대기 및 자산 확인**

```bash
gh run watch $(gh run list --workflow=release.yml --limit 1 --json databaseId -q '.[0].databaseId') --exit-status
gh release view v2.4.0
```

- [ ] **Step 4: 릴리즈 바이너리로 로컬 교체 검증** (`--version` → v2.4.0, statusline 정상)

---

### Task 12: 중심 레포(claude-plugins) 최신화

- [ ] **Step 1: SHA 갱신 + 푸시**

```bash
cd /home/jeonguk/dev/lighthouse/repositories/claude-plugins && git pull origin main
NEW_SHA=$(git ls-remote https://github.com/jeongph/claude-telemetry.git HEAD | awk '{print $1}')
jq --arg sha "$NEW_SHA" '(.plugins[] | select(.name == "claude-telemetry") | .source.sha) = $sha' .claude-plugin/marketplace.json > /tmp/mp.json && mv /tmp/mp.json .claude-plugin/marketplace.json
git add .claude-plugin/marketplace.json && git commit -m "chore: claude-telemetry v2.4.0 반영 (SHA 갱신)" && git push origin main
```

(README 한 줄 설명은 v2.3.0에서 이미 갱신됨 — 이번 기능은 비표시 동작이라 설명 변경 불필요. 단, 실행 시점에 자동 동기화 언급 추가가 자연스러우면 한 줄 보강 가능)

---

### Task 13: Notion 작업 히스토리 기록

- [ ] `~/CLAUDE.md` 컨벤션대로 기록 — 제목 "claude-telemetry v2.4.0 — 바이너리 자동 동기화 및 제거 시 자가 정리", 위치 "jeongph/claude-telemetry", 6개 섹션 (원복: 릴리즈 삭제 + revert + 중심 레포 SHA 복원, settings.json은 .bak에서 복원 가능)

---

## 리스크 및 참고

- **오탐 방어**: 자가 정리는 마커(setup 설치본) + 60초 연속 부재 + 락의 3중 가드. 플러그인 업데이트 중 캐시 공백은 수 초라 60초 유예로 충분.
- **훅 반영 시점**: 훅은 세션 시작 시 로드 — 플러그인 업데이트 직후 현재 세션은 미적용, 다음 세션부터 동기화 (구조상 불가피, README에 명시).
- **CLAUDE_CONFIG_DIR 비표준 환경**: run.sh·sync-binary.sh는 `$HOME/.claude` 고정 — 비표준 설정 디렉토리 사용자는 미지원 (기존 동작과 동일한 한계).
- **sync 실패 시**: 네트워크 부재 등에서는 조용히 스킵 — 구버전 바이너리로 계속 동작 (신규 필드만 미표시).
