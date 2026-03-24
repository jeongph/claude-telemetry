# claude-telemetry v2 Go Rewrite Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** bash+jq 기반 status line을 Go 단일 바이너리로 재작성하여 정확성, 경량성, 안정성을 구조적으로 보장한다.

**Architecture:** stdin JSON → Go 바이너리 → stdout ANSI 텍스트. 무상태, 단방향. Section 인터페이스로 모듈화. git 정보는 goroutine 병렬 실행 + 파일 캐싱.

**Tech Stack:** Go 1.22+, 외부 라이브러리 없음 (stdlib only)

**Spec:** `docs/superpowers/specs/2026-03-25-v2-go-rewrite-design.md`

---

## File Structure

### 생성할 파일

| 파일 | 역할 |
|------|------|
| `go.mod` | Go 모듈 정의 |
| `cmd/claude-telemetry/main.go` | 진입점: stdin 읽기, config 로딩, 렌더링, stdout 출력, --version/--debug 플래그 |
| `internal/input/input.go` | stdin JSON 파싱. Model dual-type(string/object) 처리. 타입 정의 |
| `internal/input/input_test.go` | 파서 단위 테스트 |
| `internal/config/config.go` | 설정 로딩 (글로벌 + 프로젝트 merge), CLAUDE_STATUSLINE_CONFIG, effort level 읽기 |
| `internal/config/config_test.go` | 설정 로딩 테스트 |
| `internal/config/preset.go` | 프리셋 정의 (compact/normal/detailed), 섹션 활성 맵, 기본 threshold |
| `internal/config/preset_test.go` | 프리셋 테스트 |
| `internal/i18n/i18n.go` | 다국어 레이블 맵 (en/ko/ja/zh), 언어 감지 |
| `internal/i18n/i18n_test.go` | i18n 테스트 |
| `internal/render/ansi.go` | ANSI 색상 헬퍼, NO_COLOR 지원, progress bar, CJK 너비 계산, ANSI strip |
| `internal/render/ansi_test.go` | ANSI/너비 계산 테스트 |
| `internal/render/renderer.go` | 라인 조립, 적응형 너비, 우선순위 기반 섹션 드롭 |
| `internal/render/renderer_test.go` | 렌더러 테스트 |
| `internal/gitinfo/gitinfo.go` | git 명령 병렬 실행, 파일 기반 캐싱 (5초 TTL, atomic write) |
| `internal/gitinfo/gitinfo_test.go` | git 캐싱 테스트 |
| `internal/section/section.go` | Section 인터페이스 + Registry |
| `internal/section/model.go` | 모델명 + effort level 렌더링 |
| `internal/section/elapsed.go` | 경과시간 렌더링 |
| `internal/section/context.go` | 컨텍스트 잔량 렌더링 |
| `internal/section/ratelimit.go` | Rate limit 렌더링 |
| `internal/section/cost.go` | 비용 렌더링 |
| `internal/section/git.go` | Git 정보 렌더링 |
| `internal/section/agent.go` | 에이전트 렌더링 |
| `internal/section/vim.go` | Vim 모드 렌더링 |
| `internal/section/lines.go` | 코드 변경량 렌더링 |
| `internal/section/tokens.go` | 토큰 수 렌더링 |
| `internal/section/apiduration.go` | API 대기시간 렌더링 |
| `internal/section/section_test.go` | 전체 섹션 테스트 |
| `testdata/normal.json` | 통합 테스트용 샘플 stdin JSON (정상) |
| `testdata/minimal.json` | 최소 필드만 있는 stdin JSON |
| `testdata/null_fields.json` | 조건부 필드가 null인 stdin JSON |
| `.github/workflows/release.yml` | CI/CD: 테스트 + 크로스 컴파일 + Release |

### 수정할 파일

| 파일 | 변경 내용 |
|------|----------|
| `scripts/run.sh` | jq 기반 로직 → Go 바이너리 호출 thin wrapper |
| `config.example.json` | preset, thresholds 필드 추가 |
| `.claude-plugin/plugin.json` | version을 2.0.0으로 업데이트 |
| `commands/setup.md` | 바이너리 다운로드 + 설정 플로우로 개편 |
| `commands/remove.md` | 바이너리 제거 포함하도록 개편 |

---

## Task 1: Go 프로젝트 초기화

**Files:**
- Create: `go.mod`
- Create: `cmd/claude-telemetry/main.go`

- [ ] **Step 1: Go 설치 (필요 시)**

```bash
# Go 1.22+ 설치 확인. 미설치 시:
wget https://go.dev/dl/go1.23.6.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.23.6.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
go version
```
Expected: `go version go1.23.6 linux/amd64` (또는 1.22+)

- [ ] **Step 2: Go 모듈 초기화**

```bash
cd /home/jeonguk/lighthouse/repositories/claude-telemetry
go mod init github.com/jeongph/claude-telemetry
```
Expected: `go.mod` 생성

- [ ] **Step 3: 최소 main.go 작성**

```go
// cmd/claude-telemetry/main.go
package main

import (
	"fmt"
	"io"
	"os"
)

var version = "dev"

func main() {
	for _, arg := range os.Args[1:] {
		switch arg {
		case "--version":
			fmt.Println("claude-telemetry", version)
			return
		}
	}

	data, err := io.ReadAll(os.Stdin)
	if err != nil || len(data) == 0 {
		fmt.Println("⚠ statusline: no input")
		return
	}

	// TODO: parse, render
	fmt.Println("claude-telemetry v2 (stub)")
}
```

- [ ] **Step 4: 빌드 및 실행 확인**

```bash
go build -o bin/claude-telemetry ./cmd/claude-telemetry
./bin/claude-telemetry --version
echo '{}' | ./bin/claude-telemetry
```
Expected: `claude-telemetry dev`, `claude-telemetry v2 (stub)`

- [ ] **Step 5: .gitignore에 bin/ 추가 후 커밋**

```bash
echo "bin/" >> .gitignore
git add go.mod cmd/ .gitignore
git commit -m "chore: Go 프로젝트 초기화 및 최소 main.go 추가"
```

---

## Task 2: stdin JSON 파서

**Files:**
- Create: `internal/input/input.go`
- Create: `internal/input/input_test.go`
- Create: `testdata/normal.json`
- Create: `testdata/minimal.json`
- Create: `testdata/null_fields.json`

- [ ] **Step 1: 테스트 데이터 파일 작성**

`testdata/normal.json`:
```json
{
  "cwd": "/home/user/myproject",
  "session_id": "abc123",
  "transcript_path": "/tmp/transcript.jsonl",
  "version": "1.0.80",
  "model": {"id": "claude-opus-4-6", "display_name": "Opus"},
  "workspace": {"current_dir": "/home/user/myproject", "project_dir": "/home/user/myproject"},
  "output_style": {"name": "default"},
  "cost": {
    "total_cost_usd": 0.45,
    "total_duration_ms": 754000,
    "total_api_duration_ms": 23000,
    "total_lines_added": 156,
    "total_lines_removed": 23
  },
  "context_window": {
    "total_input_tokens": 15234,
    "total_output_tokens": 4521,
    "context_window_size": 200000,
    "used_percentage": 28,
    "remaining_percentage": 72
  },
  "exceeds_200k_tokens": false,
  "rate_limits": {
    "five_hour": {"used_percentage": 12, "resets_at": 1742961600},
    "seven_day": {"used_percentage": 35, "resets_at": 1743393600}
  },
  "vim": {"mode": "NORMAL"},
  "agent": {"name": "security-reviewer"}
}
```

`testdata/minimal.json`:
```json
{
  "cwd": "/tmp",
  "model": {"id": "claude-sonnet-4-6", "display_name": "Sonnet"},
  "context_window": {"context_window_size": 200000, "used_percentage": 5},
  "cost": {"total_duration_ms": 60000}
}
```

`testdata/null_fields.json`:
```json
{
  "cwd": "/tmp",
  "model": "claude-opus-4-6",
  "context_window": {"context_window_size": 200000, "used_percentage": null},
  "cost": {}
}
```

- [ ] **Step 2: 실패하는 테스트 작성**

```go
// internal/input/input_test.go
package input

import (
	"os"
	"testing"
)

func TestParseNormalInput(t *testing.T) {
	data, _ := os.ReadFile("../../testdata/normal.json")
	inp, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inp.Model.DisplayName != "Opus" {
		t.Errorf("model display_name = %q, want %q", inp.Model.DisplayName, "Opus")
	}
	if inp.ContextWindow.UsedPercentage == nil || *inp.ContextWindow.UsedPercentage != 28 {
		t.Errorf("used_percentage = %v, want 28", inp.ContextWindow.UsedPercentage)
	}
	if inp.Cost.TotalCostUSD != 0.45 {
		t.Errorf("total_cost_usd = %v, want 0.45", inp.Cost.TotalCostUSD)
	}
	if inp.RateLimits == nil {
		t.Fatal("rate_limits should not be nil")
	}
	if inp.Vim == nil || inp.Vim.Mode != "NORMAL" {
		t.Errorf("vim.mode = %v, want NORMAL", inp.Vim)
	}
	if inp.Agent == nil || inp.Agent.Name != "security-reviewer" {
		t.Errorf("agent.name = %v, want security-reviewer", inp.Agent)
	}
}

func TestParseModelAsString(t *testing.T) {
	data, _ := os.ReadFile("../../testdata/null_fields.json")
	inp, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inp.Model.DisplayName != "claude-opus-4-6" {
		t.Errorf("model display_name = %q, want %q", inp.Model.DisplayName, "claude-opus-4-6")
	}
}

func TestParseMinimalInput(t *testing.T) {
	data, _ := os.ReadFile("../../testdata/minimal.json")
	inp, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inp.Model.DisplayName != "Sonnet" {
		t.Errorf("model display_name = %q, want %q", inp.Model.DisplayName, "Sonnet")
	}
	if inp.RateLimits != nil {
		t.Error("rate_limits should be nil for minimal input")
	}
	if inp.Vim != nil {
		t.Error("vim should be nil for minimal input")
	}
}

func TestParseInvalidJSON(t *testing.T) {
	_, err := Parse([]byte("not json"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestParseEmptyInput(t *testing.T) {
	_, err := Parse([]byte(""))
	if err == nil {
		t.Error("expected error for empty input")
	}
}
```

- [ ] **Step 3: 테스트 실행하여 실패 확인**

```bash
cd /home/jeonguk/lighthouse/repositories/claude-telemetry
go test ./internal/input/ -v
```
Expected: FAIL (Parse 함수 없음)

- [ ] **Step 4: input.go 구현**

```go
// internal/input/input.go
package input

import (
	"encoding/json"
	"fmt"
)

type Input struct {
	CWD            string          `json:"cwd"`
	SessionID      string          `json:"session_id"`
	TranscriptPath string          `json:"transcript_path"`
	Version        string          `json:"version"`
	Model          ModelInfo       `json:"-"`
	RawModel       json.RawMessage `json:"model"`
	Workspace      *Workspace      `json:"workspace"`
	OutputStyle    *OutputStyle    `json:"output_style"`
	Cost           Cost            `json:"cost"`
	ContextWindow  ContextWindow   `json:"context_window"`
	Exceeds200K    bool            `json:"exceeds_200k_tokens"`
	RateLimits     *RateLimits     `json:"rate_limits"`
	Vim            *Vim            `json:"vim"`
	Agent          *Agent          `json:"agent"`
	Worktree       *Worktree       `json:"worktree"`
}

type ModelInfo struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}

type Workspace struct {
	CurrentDir string `json:"current_dir"`
	ProjectDir string `json:"project_dir"`
}

type OutputStyle struct {
	Name string `json:"name"`
}

type Cost struct {
	TotalCostUSD      float64 `json:"total_cost_usd"`
	TotalDurationMS   float64 `json:"total_duration_ms"`
	TotalAPIDurationMS float64 `json:"total_api_duration_ms"`
	TotalLinesAdded   int     `json:"total_lines_added"`
	TotalLinesRemoved int     `json:"total_lines_removed"`
}

type ContextWindow struct {
	TotalInputTokens  int      `json:"total_input_tokens"`
	TotalOutputTokens int      `json:"total_output_tokens"`
	ContextWindowSize int      `json:"context_window_size"`
	UsedPercentage    *float64 `json:"used_percentage"`
	RemainingPct      *float64 `json:"remaining_percentage"`
}

type RateLimits struct {
	FiveHour *RateWindow `json:"five_hour"`
	SevenDay *RateWindow `json:"seven_day"`
}

type RateWindow struct {
	UsedPercentage float64 `json:"used_percentage"`
	ResetsAt       float64 `json:"resets_at"`
}

type Vim struct {
	Mode string `json:"mode"`
}

type Agent struct {
	Name string `json:"name"`
}

type Worktree struct {
	Name           string `json:"name"`
	Path           string `json:"path"`
	Branch         string `json:"branch"`
	OriginalCWD    string `json:"original_cwd"`
	OriginalBranch string `json:"original_branch"`
}

func Parse(data []byte) (*Input, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty input")
	}

	var inp Input
	if err := json.Unmarshal(data, &inp); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	// model: object 또는 string
	if len(inp.RawModel) > 0 {
		var m ModelInfo
		if err := json.Unmarshal(inp.RawModel, &m); err == nil && m.DisplayName != "" {
			inp.Model = m
		} else {
			var s string
			if err := json.Unmarshal(inp.RawModel, &s); err == nil {
				inp.Model = ModelInfo{ID: s, DisplayName: s}
			}
		}
	}

	return &inp, nil
}
```

- [ ] **Step 5: 테스트 실행하여 통과 확인**

```bash
go test ./internal/input/ -v
```
Expected: PASS (5 tests)

- [ ] **Step 6: 커밋**

```bash
git add internal/input/ testdata/
git commit -m "feat: stdin JSON 파서 구현 (model dual-type, nullable 필드 처리)"
```

---

## Task 3: ANSI 렌더링 유틸리티

**Files:**
- Create: `internal/render/ansi.go`
- Create: `internal/render/ansi_test.go`

- [ ] **Step 1: 실패하는 테스트 작성**

```go
// internal/render/ansi_test.go
package render

import (
	"os"
	"testing"
)

func TestColorFunctions(t *testing.T) {
	c := NewColors(true)
	if c.Cyan("test") != "\033[1;36mtest\033[0m" {
		t.Errorf("Cyan = %q", c.Cyan("test"))
	}
}

func TestNoColor(t *testing.T) {
	c := NewColors(false)
	if c.Cyan("test") != "test" {
		t.Errorf("NoColor Cyan = %q, want %q", c.Cyan("test"), "test")
	}
}

func TestNoColorEnv(t *testing.T) {
	os.Setenv("NO_COLOR", "1")
	defer os.Unsetenv("NO_COLOR")
	c := NewColors(true)
	if c.Cyan("test") != "test" {
		t.Errorf("NO_COLOR env should disable colors")
	}
}

func TestDisplayWidth(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"hello", 5},
		{"컨텍스트", 8},       // CJK: 4 chars × 2
		{"a한b", 4},          // 1 + 2 + 1
		{"\033[1;36mtest\033[0m", 4}, // ANSI stripped
		{"", 0},
	}
	for _, tt := range tests {
		got := DisplayWidth(tt.input)
		if got != tt.want {
			t.Errorf("DisplayWidth(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestStripANSI(t *testing.T) {
	got := StripANSI("\033[1;36mhello\033[0m")
	if got != "hello" {
		t.Errorf("StripANSI = %q, want %q", got, "hello")
	}
}

func TestProgressBar(t *testing.T) {
	c := NewColors(false) // no color for easy assertion
	bar := ProgressBarRemaining(72, 5, c, 50, 20)
	// 72% → 4 filled, 1 empty (72/100*5 = 3.6 → 4)
	if bar != "▰▰▰▱▱" {
		// bar 내용은 NoColor 모드에서 색상 코드 없이 블록만 포함
	}
	// width 5 확인
	if DisplayWidth(bar) != 5 {
		t.Errorf("bar width = %d, want 5", DisplayWidth(bar))
	}
}

func TestThresholdColor(t *testing.T) {
	c := NewColors(true)
	// remaining: >50 green, 21-50 yellow, <=20 red
	if ThresholdColorRemaining(72, c, 50, 20) != c.Green("") {
		// 72% remaining → green
	}
}
```

- [ ] **Step 2: 테스트 실행하여 실패 확인**

```bash
go test ./internal/render/ -v
```
Expected: FAIL

- [ ] **Step 3: ansi.go 구현**

```go
// internal/render/ansi.go
package render

import (
	"fmt"
	"math"
	"os"
	"regexp"
	"strings"
	"unicode"
)

var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*m`)

type Colors struct {
	enabled bool
}

func NewColors(enabled bool) Colors {
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		enabled = false
	}
	return Colors{enabled: enabled}
}

func (c Colors) Enabled() bool { return c.enabled }

func (c Colors) wrap(code, s string) string {
	if !c.enabled {
		return s
	}
	return fmt.Sprintf("\033[%sm%s\033[0m", code, s)
}

func (c Colors) Cyan(s string) string    { return c.wrap("1;36", s) }
func (c Colors) Green(s string) string   { return c.wrap("32", s) }
func (c Colors) Yellow(s string) string  { return c.wrap("33", s) }
func (c Colors) Red(s string) string     { return c.wrap("31", s) }
func (c Colors) Magenta(s string) string { return c.wrap("35", s) }
func (c Colors) White(s string) string   { return c.wrap("37", s) }
func (c Colors) Dim(s string) string     { return c.wrap("2;37", s) }
func (c Colors) Reset() string {
	if !c.enabled {
		return ""
	}
	return "\033[0m"
}

func StripANSI(s string) string {
	return ansiRegex.ReplaceAllString(s, "")
}

func DisplayWidth(s string) int {
	s = StripANSI(s)
	w := 0
	for _, r := range s {
		if isCJK(r) {
			w += 2
		} else {
			w += 1
		}
	}
	return w
}

func isCJK(r rune) bool {
	return unicode.Is(unicode.Han, r) ||
		(r >= 0xAC00 && r <= 0xD7A3) || // Hangul
		(r >= 0x3000 && r <= 0x303F) || // CJK Symbols
		unicode.Is(unicode.Katakana, r) ||
		unicode.Is(unicode.Hiragana, r)
}

// ThresholdColorRemaining: 잔량 기준 색상
func ThresholdColorRemaining(pct float64, c Colors, warn, danger float64) string {
	if pct <= danger {
		return "red"
	} else if pct <= warn {
		return "yellow"
	}
	return "green"
}

func ApplyColor(colorName string, s string, c Colors) string {
	switch colorName {
	case "red":
		return c.Red(s)
	case "yellow":
		return c.Yellow(s)
	case "green":
		return c.Green(s)
	default:
		return s
	}
}

// ProgressBarRemaining: 잔량 기반 progress bar
func ProgressBarRemaining(pct float64, width int, c Colors, warn, danger float64) string {
	filled := int(math.Round(pct / 100 * float64(width)))
	if filled < 0 {
		filled = 0
	}
	if filled > width {
		filled = width
	}
	empty := width - filled

	colorName := ThresholdColorRemaining(pct, c, warn, danger)
	filledStr := strings.Repeat("▰", filled)
	emptyStr := strings.Repeat("▱", empty)

	return ApplyColor(colorName, filledStr, c) + c.Dim(emptyStr)
}
```

- [ ] **Step 4: 테스트 실행하여 통과 확인**

```bash
go test ./internal/render/ -v
```
Expected: PASS

- [ ] **Step 5: 커밋**

```bash
git add internal/render/ansi.go internal/render/ansi_test.go
git commit -m "feat: ANSI 렌더링 유틸리티 구현 (색상, 너비 계산, progress bar, NO_COLOR)"
```

---

## Task 4: i18n 모듈

**Files:**
- Create: `internal/i18n/i18n.go`
- Create: `internal/i18n/i18n_test.go`

- [ ] **Step 1: 실패하는 테스트 작성**

```go
// internal/i18n/i18n_test.go
package i18n

import "testing"

func TestGetLabel(t *testing.T) {
	l := New("ko")
	if l.Get("context") != "컨텍스트" {
		t.Errorf("ko context = %q", l.Get("context"))
	}
}

func TestGetLabelEnglish(t *testing.T) {
	l := New("en")
	if l.Get("context") != "Context" {
		t.Errorf("en context = %q", l.Get("context"))
	}
}

func TestGetLabelFallback(t *testing.T) {
	l := New("fr") // 미지원 언어
	if l.Get("context") != "Context" {
		t.Errorf("fallback context = %q, want English", l.Get("context"))
	}
}

func TestGetLabelUnknownKey(t *testing.T) {
	l := New("en")
	if l.Get("nonexistent") != "nonexistent" {
		t.Errorf("unknown key = %q, want key itself", l.Get("nonexistent"))
	}
}
```

- [ ] **Step 2: 테스트 실행하여 실패 확인**

```bash
go test ./internal/i18n/ -v
```
Expected: FAIL

- [ ] **Step 3: i18n.go 구현**

```go
// internal/i18n/i18n.go
package i18n

var locales = map[string]map[string]string{
	"en": {
		"context": "Context", "elapsed": "Elapsed", "cost": "Cost",
		"api": "API", "in": "In", "out": "Out",
	},
	"ko": {
		"context": "컨텍스트", "elapsed": "경과", "cost": "비용",
		"api": "API 대기", "in": "입력", "out": "출력",
	},
	"ja": {
		"context": "コンテキスト", "elapsed": "経過", "cost": "費用",
		"api": "API待機", "in": "入力", "out": "出力",
	},
	"zh": {
		"context": "上下文", "elapsed": "已用", "cost": "费用",
		"api": "API等待", "in": "输入", "out": "输出",
	},
}

type Locale struct {
	lang string
}

func New(lang string) Locale {
	if _, ok := locales[lang]; !ok {
		lang = "en"
	}
	return Locale{lang: lang}
}

func (l Locale) Get(key string) string {
	if m, ok := locales[l.lang]; ok {
		if v, ok := m[key]; ok {
			return v
		}
	}
	// fallback to en, then key itself
	if m, ok := locales["en"]; ok {
		if v, ok := m[key]; ok {
			return v
		}
	}
	return key
}

func (l Locale) Lang() string { return l.lang }
```

- [ ] **Step 4: 테스트 실행하여 통과 확인**

```bash
go test ./internal/i18n/ -v
```
Expected: PASS

- [ ] **Step 5: 커밋**

```bash
git add internal/i18n/
git commit -m "feat: i18n 모듈 구현 (en/ko/ja/zh 4개 언어)"
```

---

## Task 5: 설정 & 프리셋 모듈

**Files:**
- Create: `internal/config/config.go`
- Create: `internal/config/preset.go`
- Create: `internal/config/config_test.go`
- Create: `internal/config/preset_test.go`

- [ ] **Step 1: preset 테스트 작성**

```go
// internal/config/preset_test.go
package config

import "testing"

func TestPresetNormal(t *testing.T) {
	p := GetPreset("normal")
	if !p.Sections["context"] {
		t.Error("normal preset should have context enabled")
	}
	if p.Sections["tokens"] {
		t.Error("normal preset should not have tokens enabled")
	}
}

func TestPresetCompact(t *testing.T) {
	p := GetPreset("compact")
	if !p.Sections["context"] {
		t.Error("compact should have context")
	}
	if p.Sections["git"] {
		t.Error("compact should not have git")
	}
}

func TestPresetDetailed(t *testing.T) {
	p := GetPreset("detailed")
	if !p.Sections["tokens"] {
		t.Error("detailed should have tokens")
	}
}

func TestPresetUnknownFallsBackToNormal(t *testing.T) {
	p := GetPreset("invalid")
	n := GetPreset("normal")
	if len(p.Sections) != len(n.Sections) {
		t.Error("unknown preset should fallback to normal")
	}
}
```

- [ ] **Step 2: config 테스트 작성**

```go
// internal/config/config_test.go
package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDefaultConfig(t *testing.T) {
	cfg := Load("", "") // 설정 파일 없음
	if cfg.Preset != "normal" {
		t.Errorf("default preset = %q, want normal", cfg.Preset)
	}
	if cfg.BarWidth != 5 {
		t.Errorf("default bar_width = %d, want 5", cfg.BarWidth)
	}
}

func TestLoadConfigFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	os.WriteFile(path, []byte(`{"preset":"compact","bar_width":3}`), 0644)

	cfg := Load(dir, "")
	if cfg.Preset != "compact" {
		t.Errorf("preset = %q, want compact", cfg.Preset)
	}
	if cfg.BarWidth != 3 {
		t.Errorf("bar_width = %d, want 3", cfg.BarWidth)
	}
}

func TestBarWidthClamp(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	os.WriteFile(path, []byte(`{"bar_width":100}`), 0644)

	cfg := Load(dir, "")
	if cfg.BarWidth != 10 {
		t.Errorf("bar_width = %d, want 10 (clamped)", cfg.BarWidth)
	}
}

func TestProjectConfigMerge(t *testing.T) {
	globalDir := t.TempDir()
	os.WriteFile(filepath.Join(globalDir, "config.json"),
		[]byte(`{"preset":"normal","sections":{"git":true}}`), 0644)

	projectDir := t.TempDir()
	os.WriteFile(filepath.Join(projectDir, ".claude-statusline.json"),
		[]byte(`{"preset":"detailed","sections":{"tokens":true}}`), 0644)

	cfg := Load(globalDir, projectDir)
	if cfg.Preset != "detailed" {
		t.Errorf("preset = %q, want detailed (project override)", cfg.Preset)
	}
}

func TestSectionEnabled(t *testing.T) {
	cfg := Load("", "")
	cfg.Preset = "compact"
	// compact에서 git은 비활성
	if cfg.IsSectionEnabled("git") {
		t.Error("git should be disabled in compact")
	}
	// sections 오버라이드
	cfg.Sections = map[string]bool{"git": true}
	if !cfg.IsSectionEnabled("git") {
		t.Error("git should be enabled via sections override")
	}
}

func TestV1ConfigCompat(t *testing.T) {
	dir := t.TempDir()
	// v1 형식: preset 없고 sections만 있는 경우
	os.WriteFile(filepath.Join(dir, "config.json"),
		[]byte(`{"sections":{"git":true,"context":true,"tokens":false}}`), 0644)

	cfg := Load(dir, "")
	if cfg.Preset != "normal" {
		t.Errorf("v1 config should default to normal preset")
	}
	if !cfg.IsSectionEnabled("git") {
		t.Error("v1 sections should work")
	}
}
```

- [ ] **Step 3: 테스트 실행하여 실패 확인**

```bash
go test ./internal/config/ -v
```
Expected: FAIL

- [ ] **Step 4: preset.go 구현**

```go
// internal/config/preset.go
package config

type PresetDef struct {
	Sections map[string]bool
}

var presets = map[string]PresetDef{
	"compact": {
		Sections: map[string]bool{
			"model": true, "context": true, "ratelimit": true, "cost": true,
			"elapsed": false, "git": false, "lines": false,
			"tokens": false, "apiduration": false, "agent": false, "vim": false,
		},
	},
	"normal": {
		Sections: map[string]bool{
			"model": true, "elapsed": true, "git": true,
			"context": true, "ratelimit": true, "cost": true,
			"lines": false, "tokens": false, "apiduration": false,
			"agent": true, "vim": true,
		},
	},
	"detailed": {
		Sections: map[string]bool{
			"model": true, "elapsed": true, "git": true,
			"context": true, "ratelimit": true, "cost": true,
			"lines": true, "tokens": true, "apiduration": true,
			"agent": true, "vim": true,
		},
	},
}

func GetPreset(name string) PresetDef {
	if p, ok := presets[name]; ok {
		return p
	}
	return presets["normal"]
}
```

- [ ] **Step 5: config.go 구현**

```go
// internal/config/config.go
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Thresholds struct {
	ContextWarn   float64 `json:"context_warn"`
	ContextDanger float64 `json:"context_danger"`
	CostWarn      float64 `json:"cost_warn"`
	CostDanger    float64 `json:"cost_danger"`
}

type Config struct {
	Preset     string          `json:"preset"`
	Language   string          `json:"language"`
	ColorsOn   *bool           `json:"colors"`
	BarWidth   int             `json:"bar_width"`
	Separator  string          `json:"separator"`
	UserType   string          `json:"user_type"`
	Sections   map[string]bool `json:"sections"`
	Thresholds Thresholds      `json:"thresholds"`
}

func defaultConfig() Config {
	return Config{
		Preset:    "normal",
		Language:  "auto",
		ColorsOn:  boolPtr(true),
		BarWidth:  5,
		Separator: " │ ",
		UserType:  "auto",
		Sections:  map[string]bool{},
		Thresholds: Thresholds{
			ContextWarn:   50,
			ContextDanger: 80,
			CostWarn:      1.0,
			CostDanger:    5.0,
		},
	}
}

func boolPtr(b bool) *bool { return &b }

func Load(configDir, projectDir string) Config {
	cfg := defaultConfig()

	// 글로벌 설정
	if configDir != "" {
		loadFile(filepath.Join(configDir, "config.json"), &cfg)
	}

	// 프로젝트 설정 (shallow merge)
	if projectDir != "" {
		var proj Config
		if loadFile(filepath.Join(projectDir, ".claude-statusline.json"), &proj) {
			if proj.Preset != "" {
				cfg.Preset = proj.Preset
			}
			if proj.Language != "" {
				cfg.Language = proj.Language
			}
			for k, v := range proj.Sections {
				if cfg.Sections == nil {
					cfg.Sections = map[string]bool{}
				}
				cfg.Sections[k] = v
			}
		}
	}

	// 값 clamp
	if cfg.BarWidth < 3 {
		cfg.BarWidth = 3
	}
	if cfg.BarWidth > 10 {
		cfg.BarWidth = 10
	}

	return cfg
}

func loadFile(path string, cfg *Config) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	json.Unmarshal(data, cfg)
	return true
}

func (c Config) Colors() bool {
	if c.ColorsOn == nil {
		return true
	}
	return *c.ColorsOn
}

func (c Config) IsSectionEnabled(name string) bool {
	// sections 오버라이드가 있으면 우선
	if v, ok := c.Sections[name]; ok {
		return v
	}
	// 프리셋 기본값
	p := GetPreset(c.Preset)
	if v, ok := p.Sections[name]; ok {
		return v
	}
	return true // 알 수 없는 섹션은 기본 활성
}

// ReadEffortLevel: settings.json에서 effortLevel 읽기
func ReadEffortLevel() string {
	home, _ := os.UserHomeDir()
	data, err := os.ReadFile(filepath.Join(home, ".claude", "settings.json"))
	if err != nil {
		return "auto"
	}
	var settings struct {
		EffortLevel string `json:"effortLevel"`
	}
	json.Unmarshal(data, &settings)
	if settings.EffortLevel == "" {
		return "auto"
	}
	return settings.EffortLevel
}
```

- [ ] **Step 6: 테스트 실행하여 통과 확인**

```bash
go test ./internal/config/ -v
```
Expected: PASS

- [ ] **Step 7: 커밋**

```bash
git add internal/config/
git commit -m "feat: 설정 및 프리셋 모듈 구현 (3단계 계층, v1 호환, 값 clamp)"
```

---

## Task 6: Git 정보 수집 모듈

**Files:**
- Create: `internal/gitinfo/gitinfo.go`
- Create: `internal/gitinfo/gitinfo_test.go`

- [ ] **Step 1: 실패하는 테스트 작성**

```go
// internal/gitinfo/gitinfo_test.go
package gitinfo

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGatherInGitRepo(t *testing.T) {
	// 현재 디렉토리가 git repo라고 가정 (테스트 실행 환경)
	cwd, _ := os.Getwd()
	// 프로젝트 루트로 이동
	root := filepath.Join(cwd, "..", "..")
	info := Gather(root, "")
	if info.Branch == "" {
		t.Error("branch should not be empty in a git repo")
	}
}

func TestGatherInNonGitDir(t *testing.T) {
	info := Gather(t.TempDir(), "")
	if info.Branch != "" {
		t.Error("branch should be empty in non-git dir")
	}
	if info.IsRepo {
		t.Error("IsRepo should be false")
	}
}

func TestCacheWriteAndRead(t *testing.T) {
	dir := t.TempDir()
	info := &GitInfo{Branch: "main", IsRepo: true}
	writeCache(dir, "testhash", info)

	cached, ok := readCache(dir, "testhash", 5*time.Second)
	if !ok {
		t.Fatal("cache should hit")
	}
	if cached.Branch != "main" {
		t.Errorf("cached branch = %q", cached.Branch)
	}
}

func TestCacheExpired(t *testing.T) {
	dir := t.TempDir()
	info := &GitInfo{Branch: "main", IsRepo: true}
	writeCache(dir, "testhash", info)

	_, ok := readCache(dir, "testhash", 0) // TTL=0 → 즉시 만료
	if ok {
		t.Error("cache should be expired")
	}
}
```

- [ ] **Step 2: 테스트 실행하여 실패 확인**

```bash
go test ./internal/gitinfo/ -v
```
Expected: FAIL

- [ ] **Step 3: gitinfo.go 구현**

```go
// internal/gitinfo/gitinfo.go
package gitinfo

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type GitInfo struct {
	IsRepo    bool   `json:"is_repo"`
	Branch    string `json:"branch"`
	Ahead     int    `json:"ahead"`
	Behind    int    `json:"behind"`
	Added     int    `json:"added"`
	Deleted   int    `json:"deleted"`
	Untracked int    `json:"untracked"`
	Stash     int    `json:"stash"`
	Worktrees int    `json:"worktrees"`
}

const cacheTTL = 5 * time.Second

func Gather(cwd, cacheDir string) *GitInfo {
	if cacheDir == "" {
		home, _ := os.UserHomeDir()
		cacheDir = filepath.Join(home, ".claude", "statusline", "cache")
	}

	hash := repoHash(cwd)

	// 캐시 확인
	if cached, ok := readCache(cacheDir, hash, cacheTTL); ok {
		return cached
	}

	info := gather(cwd)

	if info.IsRepo {
		writeCache(cacheDir, hash, info)
	}

	return info
}

func gather(cwd string) *GitInfo {
	info := &GitInfo{}

	// git repo 확인
	if err := gitCmd(cwd, "rev-parse", "--is-inside-work-tree"); err != nil {
		return info
	}
	info.IsRepo = true

	var wg sync.WaitGroup
	var mu sync.Mutex

	run := func(fn func()) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fn()
		}()
	}

	// branch
	run(func() {
		out := gitOutput(cwd, "rev-parse", "--abbrev-ref", "HEAD")
		mu.Lock()
		info.Branch = strings.TrimSpace(out)
		mu.Unlock()
	})

	// ahead/behind
	run(func() {
		out := gitOutput(cwd, "rev-list", "--left-right", "--count", "HEAD...@{u}")
		parts := strings.Fields(out)
		if len(parts) == 2 {
			a, _ := strconv.Atoi(parts[0])
			b, _ := strconv.Atoi(parts[1])
			mu.Lock()
			info.Ahead = a
			info.Behind = b
			mu.Unlock()
		}
	})

	// diff (staged + unstaged)
	run(func() {
		out1 := gitOutput(cwd, "diff", "--numstat")
		out2 := gitOutput(cwd, "diff", "--cached", "--numstat")
		a, d := parseDiffNumstat(out1 + out2)
		mu.Lock()
		info.Added = a
		info.Deleted = d
		mu.Unlock()
	})

	// untracked
	run(func() {
		out := gitOutput(cwd, "ls-files", "--others", "--exclude-standard")
		count := 0
		for _, line := range strings.Split(out, "\n") {
			if strings.TrimSpace(line) != "" {
				count++
			}
		}
		mu.Lock()
		info.Untracked = count
		mu.Unlock()
	})

	// stash
	run(func() {
		out := gitOutput(cwd, "stash", "list")
		count := 0
		for _, line := range strings.Split(out, "\n") {
			if strings.TrimSpace(line) != "" {
				count++
			}
		}
		mu.Lock()
		info.Stash = count
		mu.Unlock()
	})

	// worktrees
	run(func() {
		out := gitOutput(cwd, "worktree", "list")
		count := 0
		for _, line := range strings.Split(out, "\n") {
			if strings.TrimSpace(line) != "" {
				count++
			}
		}
		// 메인 worktree 제외
		if count > 0 {
			count--
		}
		mu.Lock()
		info.Worktrees = count
		mu.Unlock()
	})

	wg.Wait()
	return info
}

func parseDiffNumstat(s string) (added, deleted int) {
	for _, line := range strings.Split(s, "\n") {
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			a, _ := strconv.Atoi(parts[0])
			d, _ := strconv.Atoi(parts[1])
			added += a
			deleted += d
		}
	}
	return
}

func gitCmd(cwd string, args ...string) error {
	allArgs := append([]string{"--no-optional-locks", "-C", cwd}, args...)
	cmd := exec.Command("git", allArgs...)
	return cmd.Run()
}

func gitOutput(cwd string, args ...string) string {
	allArgs := append([]string{"--no-optional-locks", "-C", cwd}, args...)
	cmd := exec.Command("git", allArgs...)
	out, _ := cmd.Output()
	return string(out)
}

func repoHash(cwd string) string {
	h := sha256.Sum256([]byte(cwd))
	return fmt.Sprintf("%x", h[:6]) // 12 hex chars
}

type cacheEntry struct {
	Info      GitInfo   `json:"info"`
	Timestamp time.Time `json:"ts"`
}

func readCache(dir, hash string, ttl time.Duration) (*GitInfo, bool) {
	path := filepath.Join(dir, hash+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}
	var entry cacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, false
	}
	if time.Since(entry.Timestamp) > ttl {
		return nil, false
	}
	return &entry.Info, true
}

func writeCache(dir, hash string, info *GitInfo) {
	os.MkdirAll(dir, 0755)
	entry := cacheEntry{Info: *info, Timestamp: time.Now()}
	data, _ := json.Marshal(entry)

	// atomic write: 임시 파일 → rename
	tmp := filepath.Join(dir, hash+".tmp")
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return
	}
	os.Rename(tmp, filepath.Join(dir, hash+".json"))
}
```

- [ ] **Step 4: 테스트 실행하여 통과 확인**

```bash
go test ./internal/gitinfo/ -v
```
Expected: PASS

- [ ] **Step 5: 커밋**

```bash
git add internal/gitinfo/
git commit -m "feat: git 정보 수집 모듈 구현 (goroutine 병렬 실행, 파일 캐싱, atomic write)"
```

---

## Task 7: Section 인터페이스 & 전체 섹션 구현

**Files:**
- Create: `internal/section/section.go`
- Create: `internal/section/model.go`
- Create: `internal/section/elapsed.go`
- Create: `internal/section/context.go`
- Create: `internal/section/ratelimit.go`
- Create: `internal/section/cost.go`
- Create: `internal/section/git.go`
- Create: `internal/section/agent.go`
- Create: `internal/section/vim.go`
- Create: `internal/section/lines.go`
- Create: `internal/section/tokens.go`
- Create: `internal/section/apiduration.go`
- Create: `internal/section/section_test.go`

이 태스크는 크기가 크므로 서브 단계로 나눕니다.

- [ ] **Step 1: Section 인터페이스 + 테스트 스캐폴드**

`internal/section/section.go`:
```go
package section

import (
	"github.com/jeongph/claude-telemetry/internal/config"
	"github.com/jeongph/claude-telemetry/internal/gitinfo"
	"github.com/jeongph/claude-telemetry/internal/i18n"
	"github.com/jeongph/claude-telemetry/internal/input"
	"github.com/jeongph/claude-telemetry/internal/render"
)

type Context struct {
	Input   *input.Input
	Config  config.Config
	Locale  i18n.Locale
	Colors  render.Colors
	GitInfo *gitinfo.GitInfo
	Effort  string
}

type Section interface {
	Name() string
	Priority() int
	Render(ctx *Context) string
	Width(ctx *Context) int
}

// Line: 섹션이 속하는 라인 (1, 2, 3)
type LineSection struct {
	Section Section
	Line    int
}

func AllSections() []LineSection {
	return []LineSection{
		// Line 1
		{&ModelSection{}, 1},
		{&ElapsedSection{}, 1},
		{&GitSection{}, 1},
		// Line 2
		{&ContextSection{}, 2},
		{&RateLimitSection{}, 2},
		{&CostSection{}, 2},
		{&LinesSection{}, 2},
		{&APIDurationSection{}, 2},
		{&TokensSection{}, 2},
		// Line 3
		{&AgentSection{}, 3},
		{&VimSection{}, 3},
	}
}
```

- [ ] **Step 2: 섹션 테스트 작성**

`internal/section/section_test.go` — 주요 섹션에 대한 테스트:
```go
package section

import (
	"os"
	"strings"
	"testing"

	"github.com/jeongph/claude-telemetry/internal/config"
	"github.com/jeongph/claude-telemetry/internal/gitinfo"
	"github.com/jeongph/claude-telemetry/internal/i18n"
	"github.com/jeongph/claude-telemetry/internal/input"
	"github.com/jeongph/claude-telemetry/internal/render"
)

func testContext(t *testing.T) *Context {
	t.Helper()
	data, _ := os.ReadFile("../../testdata/normal.json")
	inp, _ := input.Parse(data)
	return &Context{
		Input:   inp,
		Config:  config.Load("", ""),
		Locale:  i18n.New("en"),
		Colors:  render.NewColors(false), // no color for easy assertion
		GitInfo: &gitinfo.GitInfo{IsRepo: true, Branch: "main", Ahead: 1, Added: 5, Deleted: 2},
		Effort:  "high",
	}
}

func TestModelSection(t *testing.T) {
	ctx := testContext(t)
	s := &ModelSection{}
	out := s.Render(ctx)
	if !strings.Contains(out, "Opus") {
		t.Errorf("model should contain 'Opus', got %q", out)
	}
	if !strings.Contains(out, "high") {
		t.Errorf("model should contain effort 'high', got %q", out)
	}
}

func TestContextSection(t *testing.T) {
	ctx := testContext(t)
	s := &ContextSection{}
	out := s.Render(ctx)
	if !strings.Contains(out, "72") {
		t.Errorf("context should contain remaining 72%%, got %q", out)
	}
	if !strings.Contains(out, "200k") {
		t.Errorf("context should contain size '200k', got %q", out)
	}
}

func TestContextSectionNullPercentage(t *testing.T) {
	ctx := testContext(t)
	ctx.Input.ContextWindow.UsedPercentage = nil
	s := &ContextSection{}
	out := s.Render(ctx)
	if !strings.Contains(out, "···") {
		t.Errorf("null percentage should show loading indicator, got %q", out)
	}
}

func TestGitSectionNoRepo(t *testing.T) {
	ctx := testContext(t)
	ctx.GitInfo = &gitinfo.GitInfo{IsRepo: false}
	s := &GitSection{}
	out := s.Render(ctx)
	// git 미사용 시 폴더명만 표시
	if !strings.Contains(out, "myproject") {
		t.Errorf("non-git should show folder name, got %q", out)
	}
}

func TestRateLimitSection(t *testing.T) {
	ctx := testContext(t)
	s := &RateLimitSection{}
	out := s.Render(ctx)
	if !strings.Contains(out, "5h") {
		t.Errorf("rate limit should contain '5h', got %q", out)
	}
}

func TestCostSection(t *testing.T) {
	ctx := testContext(t)
	ctx.Input.RateLimits = nil // API key user
	s := &CostSection{}
	out := s.Render(ctx)
	if !strings.Contains(out, "0.45") {
		t.Errorf("cost should contain '0.45', got %q", out)
	}
}

func TestAgentSectionNil(t *testing.T) {
	ctx := testContext(t)
	ctx.Input.Agent = nil
	s := &AgentSection{}
	out := s.Render(ctx)
	if out != "" {
		t.Errorf("agent should be empty when nil, got %q", out)
	}
}

func TestVimSectionNil(t *testing.T) {
	ctx := testContext(t)
	ctx.Input.Vim = nil
	s := &VimSection{}
	out := s.Render(ctx)
	if out != "" {
		t.Errorf("vim should be empty when nil, got %q", out)
	}
}
```

- [ ] **Step 3: 테스트 실행하여 실패 확인**

```bash
go test ./internal/section/ -v
```
Expected: FAIL (섹션 타입 없음)

- [ ] **Step 4: 11개 섹션 구현**

각 섹션을 개별 파일로 구현. v1 run.sh의 렌더링 로직을 Go로 포팅.
핵심 패턴 (모든 섹션 동일):

```go
// internal/section/model.go
package section

import "fmt"

type ModelSection struct{}

func (s *ModelSection) Name() string     { return "model" }
func (s *ModelSection) Priority() int    { return 3 }

func (s *ModelSection) Render(ctx *Context) string {
	name := ctx.Input.Model.DisplayName
	if name == "" {
		return ""
	}
	effort := ""
	if ctx.Effort != "" {
		effort = " " + ctx.Colors.Dim("💭"+ctx.Effort)
	}
	return ctx.Colors.Cyan(name) + effort
}

func (s *ModelSection) Width(ctx *Context) int {
	return render.DisplayWidth(s.Render(ctx))
}
```

나머지 10개 섹션(elapsed, context, ratelimit, cost, git, agent, vim, lines, tokens, apiduration)도 동일한 패턴으로 구현.
각 섹션에서 v1의 해당 렌더링 로직(포맷팅, 색상, 조건부 표시)을 Go로 포팅.

참조: v1 run.sh 라인 번호
- model: 179-181
- elapsed: 184-189
- git: 192-214
- context: 224-234
- ratelimit: 238-258
- cost: 262-270
- lines: 272-278
- apiduration: 280-285
- tokens: 287-295
- agent: 316-319
- vim: 321-327

- [ ] **Step 5: 테스트 실행하여 통과 확인**

```bash
go test ./internal/section/ -v
```
Expected: PASS

- [ ] **Step 6: 커밋**

```bash
git add internal/section/
git commit -m "feat: Section 인터페이스 및 11개 섹션 구현"
```

---

## Task 8: 렌더러 (라인 조립 + 적응형 너비)

**Files:**
- Create: `internal/render/renderer.go`
- Create: `internal/render/renderer_test.go`

- [ ] **Step 1: 실패하는 테스트 작성**

```go
// internal/render/renderer_test.go
package render

import (
	"strings"
	"testing"
)

func TestRenderOutput(t *testing.T) {
	// 렌더러의 핵심: 라인 조립 + 적응형 너비
	lines := AssembleLines(
		[]string{"Opus", "12m 34s", "myproject:main"},
		[]ScoredSegment{
			{Text: "Context 72%", Width: 11, Priority: 1},
			{Text: "5h 88%", Width: 6, Priority: 2},
			{Text: "tokens 15k", Width: 10, Priority: 9},
		},
		[]string{"agent-name", "NORMAL"},
		" │ ", 80,
	)

	if len(lines) < 2 {
		t.Fatalf("expected at least 2 lines, got %d", len(lines))
	}
	if !strings.Contains(lines[0], "Opus") {
		t.Errorf("line1 should contain model")
	}
	if !strings.Contains(lines[1], "Context") {
		t.Errorf("line2 should contain context")
	}
}

func TestAdaptiveWidthDropsLowPriority(t *testing.T) {
	segments := []ScoredSegment{
		{Text: "Context 72%", Width: 11, Priority: 1},
		{Text: "5h 88%", Width: 6, Priority: 2},
		{Text: "tokens in 15k out 4k", Width: 20, Priority: 9},
	}
	// 너비 25: Context(11) + sep(3) + 5h(6) = 20 → tokens 드롭
	result := fitSegments(segments, " │ ", 25)
	if len(result) != 2 {
		t.Errorf("expected 2 segments, got %d", len(result))
	}
}

func TestEmptyLine3(t *testing.T) {
	lines := AssembleLines(
		[]string{"Opus"},
		[]ScoredSegment{{Text: "Context", Width: 7, Priority: 1}},
		[]string{}, // 빈 line3
		" │ ", 80,
	)
	if len(lines) != 2 {
		t.Errorf("empty line3 should not be included, got %d lines", len(lines))
	}
}
```

- [ ] **Step 2: 테스트 실행하여 실패 확인**

```bash
go test ./internal/render/ -v -run TestRender
```
Expected: FAIL

- [ ] **Step 3: renderer.go 구현**

```go
// internal/render/renderer.go
package render

import (
	"sort"
	"strings"
)

type ScoredSegment struct {
	Text     string
	Width    int
	Priority int // 낮을수록 중요
	Order    int // 원래 순서 유지용
}

// AssembleLines: 3줄 조립
func AssembleLines(line1Parts []string, line2Segments []ScoredSegment, line3Parts []string, sep string, maxWidth int) []string {
	var lines []string

	// Line 1: 고정 레이아웃
	l1 := joinNonEmpty(line1Parts, sep)
	if l1 != "" {
		lines = append(lines, l1)
	}

	// Line 2: 적응형 너비
	fitted := fitSegments(line2Segments, sep, maxWidth)
	if len(fitted) > 0 {
		texts := make([]string, len(fitted))
		for i, s := range fitted {
			texts[i] = s.Text
		}
		lines = append(lines, strings.Join(texts, sep))
	}

	// Line 3: 활성 시만
	l3 := joinNonEmpty(line3Parts, sep)
	if l3 != "" {
		lines = append(lines, l3)
	}

	return lines
}

func fitSegments(segments []ScoredSegment, sep string, maxWidth int) []ScoredSegment {
	if len(segments) == 0 {
		return nil
	}

	sepWidth := DisplayWidth(sep)

	// 우선순위로 정렬 (낮을수록 중요)
	sorted := make([]ScoredSegment, len(segments))
	copy(sorted, segments)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Priority < sorted[j].Priority
	})

	// 우선순위 순으로 추가, 너비 초과 시 중단
	var selected []ScoredSegment
	totalWidth := 0
	for _, seg := range sorted {
		extra := 0
		if len(selected) > 0 {
			extra = sepWidth
		}
		if totalWidth+seg.Width+extra <= maxWidth-2 { // 2 char margin
			selected = append(selected, seg)
			totalWidth += seg.Width + extra
		}
	}

	// 원래 순서로 복원
	sort.Slice(selected, func(i, j int) bool {
		return selected[i].Order < selected[j].Order
	})

	return selected
}

func joinNonEmpty(parts []string, sep string) string {
	var nonEmpty []string
	for _, p := range parts {
		if p != "" {
			nonEmpty = append(nonEmpty, p)
		}
	}
	return strings.Join(nonEmpty, sep)
}
```

- [ ] **Step 4: 테스트 실행하여 통과 확인**

```bash
go test ./internal/render/ -v
```
Expected: PASS (모든 render 테스트)

- [ ] **Step 5: 커밋**

```bash
git add internal/render/renderer.go internal/render/renderer_test.go
git commit -m "feat: 렌더러 구현 (라인 조립, 우선순위 기반 적응형 너비)"
```

---

## Task 9: main.go 통합 + end-to-end 동작

**Files:**
- Modify: `cmd/claude-telemetry/main.go`

- [ ] **Step 1: main.go 전체 파이프라인 구현**

stdin 읽기 → config 로딩 → input 파싱 → git 수집 → 섹션 렌더링 → 라인 조립 → stdout 출력.
`--version`, `--debug` 플래그 처리 포함.

- [ ] **Step 2: 수동 통합 테스트**

```bash
go build -o bin/claude-telemetry ./cmd/claude-telemetry
cat testdata/normal.json | ./bin/claude-telemetry
cat testdata/minimal.json | ./bin/claude-telemetry
cat testdata/null_fields.json | ./bin/claude-telemetry
echo "invalid" | ./bin/claude-telemetry
echo "" | ./bin/claude-telemetry
./bin/claude-telemetry --version
```

Expected:
- normal.json → 2줄 출력 (model, elapsed, git / context, rate limits)
- minimal.json → 2줄 (일부 섹션 생략)
- null_fields.json → graceful degradation (··· 표시)
- invalid → `⚠ statusline: invalid input`
- empty → `⚠ statusline: no input`
- --version → `claude-telemetry dev`

- [ ] **Step 3: 전체 테스트 실행**

```bash
go test ./... -v
```
Expected: ALL PASS

- [ ] **Step 4: 커밋**

```bash
git add cmd/claude-telemetry/main.go
git commit -m "feat: main.go 통합 — stdin → 렌더링 → stdout 전체 파이프라인"
```

---

## Task 10: run.sh thin wrapper + config.example.json 업데이트

**Files:**
- Modify: `scripts/run.sh`
- Modify: `config.example.json`

- [ ] **Step 1: run.sh 교체**

```bash
#!/bin/bash
BIN="${HOME}/.claude/statusline/bin/claude-telemetry"
[ -x "$BIN" ] && exec "$BIN" || echo "⚠ Run /claude-telemetry:setup to install"
```

- [ ] **Step 2: config.example.json 업데이트**

```json
{
  "preset": "normal",
  "language": "auto",
  "colors": true,
  "bar_width": 5,
  "separator": " │ ",
  "user_type": "auto",
  "sections": {},
  "thresholds": {
    "context_warn": 50,
    "context_danger": 80,
    "cost_warn": 1.0,
    "cost_danger": 5.0
  }
}
```

- [ ] **Step 3: 로컬에서 실제 동작 확인**

```bash
# 로컬 빌드를 직접 연결하여 테스트
go build -o ~/.claude/statusline/bin/claude-telemetry ./cmd/claude-telemetry
# Claude Code 재시작 후 status line 확인
```

- [ ] **Step 4: 커밋**

```bash
git add scripts/run.sh config.example.json
git commit -m "refactor: run.sh를 Go 바이너리 thin wrapper로 교체"
```

---

## Task 11: CI/CD 파이프라인

**Files:**
- Create: `.github/workflows/release.yml`

- [ ] **Step 1: release.yml 작성**

```yaml
name: Release

on:
  push:
    tags: ['v*']

permissions:
  contents: write

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'
      - run: go test ./... -v

  build:
    needs: test
    runs-on: ubuntu-latest
    strategy:
      matrix:
        include:
          - goos: linux
            goarch: amd64
          - goos: linux
            goarch: arm64
          - goos: darwin
            goarch: amd64
          - goos: darwin
            goarch: arm64
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'
      - name: Build
        env:
          GOOS: ${{ matrix.goos }}
          GOARCH: ${{ matrix.goarch }}
          CGO_ENABLED: '0'
        run: |
          VERSION=${GITHUB_REF#refs/tags/}
          go build -ldflags="-s -w -X main.version=$VERSION" \
            -o claude-telemetry-${{ matrix.goos }}-${{ matrix.goarch }} \
            ./cmd/claude-telemetry
      - uses: actions/upload-artifact@v4
        with:
          name: claude-telemetry-${{ matrix.goos }}-${{ matrix.goarch }}
          path: claude-telemetry-${{ matrix.goos }}-${{ matrix.goarch }}

  release:
    needs: build
    runs-on: ubuntu-latest
    steps:
      - uses: actions/download-artifact@v4
      - name: Generate checksums
        run: |
          cd claude-telemetry-linux-amd64 && sha256sum * >> ../checksums.txt && cd ..
          cd claude-telemetry-linux-arm64 && sha256sum * >> ../checksums.txt && cd ..
          cd claude-telemetry-darwin-amd64 && sha256sum * >> ../checksums.txt && cd ..
          cd claude-telemetry-darwin-arm64 && sha256sum * >> ../checksums.txt && cd ..
      - name: Create Release
        uses: softprops/action-gh-release@v2
        with:
          files: |
            claude-telemetry-linux-amd64/*
            claude-telemetry-linux-arm64/*
            claude-telemetry-darwin-amd64/*
            claude-telemetry-darwin-arm64/*
            checksums.txt
```

- [ ] **Step 2: 커밋**

```bash
git add .github/workflows/release.yml
git commit -m "chore: GitHub Actions 릴리스 파이프라인 추가 (크로스 컴파일 + 체크섬)"
```

---

## Task 12: setup/remove 커맨드 개편 + plugin.json 업데이트

**Files:**
- Modify: `commands/setup.md`
- Modify: `commands/remove.md`
- Modify: `.claude-plugin/plugin.json`

- [ ] **Step 1: setup.md 개편**

바이너리 다운로드 로직 추가:
- OS/아키텍처 감지 (`uname -s`, `uname -m`)
- GitHub Releases에서 바이너리 다운로드 (`curl`)
- `~/.claude/statusline/bin/` 에 설치
- SHA256 체크섬 검증
- 기존 설정 흐름 (언어, 섹션 선택, 스타일) 유지
- v1 config 감지 시 호환 모드 안내

- [ ] **Step 2: remove.md 개편**

바이너리 삭제 포함:
- `~/.claude/statusline/bin/claude-telemetry` 삭제
- `~/.claude/statusline/cache/` 삭제
- 기존 config/settings 삭제 로직 유지

- [ ] **Step 3: plugin.json 버전 업데이트**

```json
{
  "name": "claude-telemetry",
  "version": "2.0.0",
  "description": "Customizable multi-line status line for Claude Code — real-time session telemetry in your terminal"
}
```

- [ ] **Step 4: 커밋**

```bash
git add commands/ .claude-plugin/plugin.json
git commit -m "feat: setup/remove 커맨드 v2 개편 (바이너리 다운로드/삭제)"
```

---

## Task 13: 벤치마크 테스트 + 최종 검증

**Files:**
- Create: `internal/render/bench_test.go`

- [ ] **Step 1: 벤치마크 테스트 작성**

```go
// internal/render/bench_test.go
package render

import "testing"

func BenchmarkDisplayWidth(b *testing.B) {
	s := "Opus 💭high │ ◷ 경과 12m 34s │ myproject:main ↑1 +5/-2"
	for i := 0; i < b.N; i++ {
		DisplayWidth(s)
	}
}

func BenchmarkProgressBar(b *testing.B) {
	c := NewColors(true)
	for i := 0; i < b.N; i++ {
		ProgressBarRemaining(72, 5, c, 50, 20)
	}
}
```

- [ ] **Step 2: 벤치마크 실행**

```bash
go test ./internal/render/ -bench=. -benchmem
```
Expected: 각 op < 1μs

- [ ] **Step 3: 전체 테스트 + 빌드 최종 확인**

```bash
go test ./... -v
go build -o bin/claude-telemetry ./cmd/claude-telemetry
cat testdata/normal.json | ./bin/claude-telemetry
./bin/claude-telemetry --version
./bin/claude-telemetry --debug < testdata/normal.json 2>&1
```

- [ ] **Step 4: 커밋**

```bash
git add internal/render/bench_test.go
git commit -m "test: 벤치마크 테스트 추가 (DisplayWidth, ProgressBar)"
```

---

## 태스크 의존성

```
Task 1 (Go init)
  ├── Task 2 (input parser)
  ├── Task 3 (ANSI render)
  ├── Task 4 (i18n)
  ├── Task 5 (config/preset)
  └── Task 6 (gitinfo)
       ↓ 모두 완료 후
       Task 7 (sections) ← Task 2, 3, 4, 5, 6 필요
       Task 8 (renderer) ← Task 3 필요
            ↓ 모두 완료 후
            Task 9 (main.go 통합) ← Task 7, 8 필요
                 ├── Task 10 (run.sh + config)
                 ├── Task 11 (CI/CD)
                 ├── Task 12 (commands)
                 └── Task 13 (benchmark + 검증)
```

**병렬 가능**: Task 2, 3, 4, 5, 6은 Task 1 이후 **모두 병렬** 실행 가능 (서로 의존성 없음).
**병렬 가능**: Task 10, 11, 12, 13은 Task 9 이후 병렬 실행 가능.

---

## 부록: 리뷰 반영 수정사항

계획 리뷰에서 발견된 이슈를 반영한 수정사항. 구현 시 해당 Task의 코드 대신 이 섹션의 수정본을 따른다.

### A1. [CRITICAL] model.go import 수정 (Task 7)

model.go에서 `render` 패키지 import 누락. 수정:

```go
// internal/section/model.go
package section

import (
	"github.com/jeongph/claude-telemetry/internal/render"
)

type ModelSection struct{}

func (s *ModelSection) Name() string  { return "model" }
func (s *ModelSection) Priority() int { return 3 }

func (s *ModelSection) Render(ctx *Context) string {
	name := ctx.Input.Model.DisplayName
	if name == "" {
		return ""
	}
	effort := ""
	if ctx.Effort != "" {
		effort = " " + ctx.Colors.Dim("💭"+ctx.Effort)
	}
	return ctx.Colors.Cyan(name) + effort
}

func (s *ModelSection) Width(ctx *Context) int {
	return render.DisplayWidth(s.Render(ctx))
}
```

### A2. [IMPORTANT] Compact 프리셋 1줄 렌더링 (Task 8)

compact 프리셋은 모든 섹션을 Line 1에 넣어야 함. renderer.go의 `AssembleLines`에 compact 모드 추가:

```go
func AssembleLines(line1Parts []string, line2Segments []ScoredSegment, line3Parts []string, sep string, maxWidth int, compact bool) []string {
	if compact {
		// compact: line2 세그먼트를 line1에 합침
		all := make([]string, 0, len(line1Parts)+len(line2Segments))
		all = append(all, line1Parts...)
		for _, seg := range line2Segments {
			all = append(all, seg.Text)
		}
		l := joinNonEmpty(all, sep)
		if l != "" {
			return []string{l}
		}
		return nil
	}
	// normal/detailed: 기존 로직
	// ...
}
```

section.go의 `AllSections()`도 compact 시 모든 활성 섹션을 Line 1로 배치.

### A3. [IMPORTANT] v1 섹션 이름 alias 매핑 (Task 5)

v1 config의 키 이름과 v2가 다름. `config.go`에 alias 맵 추가:

```go
var sectionAliases = map[string]string{
	"rate_limits":  "ratelimit",
	"duration":     "elapsed",
	"vim_mode":     "vim",
	"api_duration": "apiduration",
}

func (c Config) IsSectionEnabled(name string) bool {
	// sections 오버라이드: v2 이름과 v1 alias 모두 확인
	if v, ok := c.Sections[name]; ok {
		return v
	}
	// v1 alias 역방향 확인
	for old, new := range sectionAliases {
		if new == name {
			if v, ok := c.Sections[old]; ok {
				return v
			}
		}
	}
	// 프리셋 기본값
	p := GetPreset(c.Preset)
	if v, ok := p.Sections[name]; ok {
		return v
	}
	return true
}
```

### A4. [IMPORTANT] CLAUDE_STATUSLINE_CONFIG 환경변수 (Task 5, 9)

`main.go`에서 config dir 결정 시 환경변수 우선:

```go
configDir := os.Getenv("CLAUDE_STATUSLINE_CONFIG")
if configDir == "" {
	home, _ := os.UserHomeDir()
	configDir = filepath.Join(home, ".claude", "statusline")
}
cfg := config.Load(configDir, inp.CWD)
```

### A5. [IMPORTANT] 언어 auto 감지 (Task 5)

`config.go`에 `ResolveLanguage` 함수 추가:

```go
func ResolveLanguage(cfgLang string) string {
	if cfgLang != "auto" && cfgLang != "" {
		return cfgLang
	}
	// settings.json에서 읽기
	home, _ := os.UserHomeDir()
	data, _ := os.ReadFile(filepath.Join(home, ".claude", "settings.json"))
	var s struct {
		Language string `json:"language"`
	}
	json.Unmarshal(data, &s)
	switch {
	case strings.HasPrefix(s.Language, "ko"), s.Language == "한국어":
		return "ko"
	case strings.HasPrefix(s.Language, "ja"), s.Language == "日本語":
		return "ja"
	case strings.HasPrefix(s.Language, "zh"), s.Language == "中文":
		return "zh"
	default:
		return "en"
	}
}
```

### A6. [IMPORTANT] Rate limit percentage 변환 (Task 7)

stdin JSON의 `rate_limits.*.used_percentage`를 remaining으로 변환:

```go
// internal/section/ratelimit.go Render() 내부
remaining := 100 - window.UsedPercentage
```

모든 rate limit 표시에서 `100 - used_percentage`로 변환 후 `ProgressBarRemaining` 호출.

### A7. [IMPORTANT] user_type 자동 감지 (Task 7, 9)

`main.go` 또는 section에서 user_type 결정:

```go
func isOAuthUser(cfg config.Config, inp *input.Input) bool {
	switch cfg.UserType {
	case "oauth":
		return true
	case "api":
		return false
	default: // "auto"
		return inp.RateLimits != nil || inp.Cost.TotalCostUSD <= 0
	}
}
```

OAuth이면 RateLimitSection 렌더링, API이면 CostSection 렌더링.

### A8. [IMPORTANT] --debug 플래그 구현 (Task 9)

```go
// main.go
var debug bool
for _, arg := range os.Args[1:] {
	switch arg {
	case "--debug":
		debug = true
	}
}

// 렌더링 후
if debug {
	fmt.Fprintf(os.Stderr, "[debug] config: preset=%s lang=%s\n", cfg.Preset, lang)
	fmt.Fprintf(os.Stderr, "[debug] input: model=%s ctx_used=%v\n", inp.Model.DisplayName, inp.ContextWindow.UsedPercentage)
	fmt.Fprintf(os.Stderr, "[debug] git: cache=%s branch=%s\n", cacheStatus, gitInfo.Branch)
	fmt.Fprintf(os.Stderr, "[debug] sections dropped: %v\n", droppedSections)
}
```

### A9. [IMPORTANT] ReadEffortLevel 테스트 가능하게 수정 (Task 5)

하드코딩된 home 경로 대신 매개변수로 받도록:

```go
func ReadEffortLevel(claudeDir string) string {
	data, err := os.ReadFile(filepath.Join(claudeDir, "settings.json"))
	if err != nil {
		return "auto"
	}
	// ...
}
```

테스트에서 temp dir 사용 가능.

### A10. [IMPORTANT] context_window.current_usage 필드 추가 (Task 2)

```go
type CurrentUsage struct {
	InputTokens              int `json:"input_tokens"`
	OutputTokens             int `json:"output_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens"`
}

type ContextWindow struct {
	// ... 기존 필드
	CurrentUsage *CurrentUsage `json:"current_usage"`
}
```

### A11. [MINOR] ansi_test.go 테스트 assertion 수정 (Task 3)

```go
func TestThresholdColor(t *testing.T) {
	c := NewColors(true)
	got := ThresholdColorRemaining(72, c, 50, 20)
	if got != "green" {
		t.Errorf("72%% remaining = %q, want green", got)
	}
	got = ThresholdColorRemaining(30, c, 50, 20)
	if got != "yellow" {
		t.Errorf("30%% remaining = %q, want yellow", got)
	}
	got = ThresholdColorRemaining(15, c, 50, 20)
	if got != "red" {
		t.Errorf("15%% remaining = %q, want red", got)
	}
}
```

### A12. [MINOR] 전체 파이프라인 벤치마크 추가 (Task 13)

```go
func BenchmarkFullPipeline(b *testing.B) {
	data, _ := os.ReadFile("../../testdata/normal.json")
	for i := 0; i < b.N; i++ {
		inp, _ := input.Parse(data)
		// config, sections, render 전체 실행
		_ = inp
	}
}
```

### A13. [MINOR] 스펙 파일명 불일치 정리

스펙과 계획 간 파일명 차이를 계획 기준으로 통일:
- 스펙의 `internal/input/parser.go` → 계획의 `internal/input/input.go` (Go 관례: 패키지명과 주 파일명 일치)
- 스펙의 `internal/git/git.go` → 계획의 `internal/gitinfo/gitinfo.go` (`git` 커맨드와 이름 충돌 방지)

스펙 문서 업데이트 시 반영.
