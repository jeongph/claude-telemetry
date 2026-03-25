package input_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/jeongph/claude-telemetry/internal/input"
)

func testdataPath(name string) string {
	_, filename, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(filename), "..", "..", "testdata")
	return filepath.Join(root, name)
}

func readTestdata(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(testdataPath(name))
	if err != nil {
		t.Fatalf("testdata 파일 읽기 실패 (%s): %v", name, err)
	}
	return data
}

func TestParseNormalInput(t *testing.T) {
	data := readTestdata(t, "normal.json")
	inp, err := input.Parse(data)
	if err != nil {
		t.Fatalf("Parse 실패: %v", err)
	}

	// 기본 필드
	if inp.CWD != "/home/user/myproject" {
		t.Errorf("CWD = %q, want %q", inp.CWD, "/home/user/myproject")
	}
	if inp.SessionID != "abc123" {
		t.Errorf("SessionID = %q, want %q", inp.SessionID, "abc123")
	}
	if inp.TranscriptPath != "/tmp/transcript.jsonl" {
		t.Errorf("TranscriptPath = %q, want %q", inp.TranscriptPath, "/tmp/transcript.jsonl")
	}
	if inp.Version != "1.0.80" {
		t.Errorf("Version = %q, want %q", inp.Version, "1.0.80")
	}

	// Model (object 형태)
	if inp.Model.ID != "claude-opus-4-6" {
		t.Errorf("Model.ID = %q, want %q", inp.Model.ID, "claude-opus-4-6")
	}
	if inp.Model.DisplayName != "Opus" {
		t.Errorf("Model.DisplayName = %q, want %q", inp.Model.DisplayName, "Opus")
	}

	// Workspace
	if inp.Workspace == nil {
		t.Fatal("Workspace가 nil이면 안 됨")
	}
	if inp.Workspace.CurrentDir != "/home/user/myproject" {
		t.Errorf("Workspace.CurrentDir = %q, want %q", inp.Workspace.CurrentDir, "/home/user/myproject")
	}

	// OutputStyle
	if inp.OutputStyle == nil {
		t.Fatal("OutputStyle가 nil이면 안 됨")
	}
	if inp.OutputStyle.Name != "default" {
		t.Errorf("OutputStyle.Name = %q, want %q", inp.OutputStyle.Name, "default")
	}

	// Cost
	if inp.Cost.TotalCostUSD != 0.45 {
		t.Errorf("Cost.TotalCostUSD = %v, want 0.45", inp.Cost.TotalCostUSD)
	}
	if inp.Cost.TotalDurationMS != 754000 {
		t.Errorf("Cost.TotalDurationMS = %v, want 754000", inp.Cost.TotalDurationMS)
	}
	if inp.Cost.TotalLinesAdded != 156 {
		t.Errorf("Cost.TotalLinesAdded = %v, want 156", inp.Cost.TotalLinesAdded)
	}
	if inp.Cost.TotalLinesRemoved != 23 {
		t.Errorf("Cost.TotalLinesRemoved = %v, want 23", inp.Cost.TotalLinesRemoved)
	}

	// ContextWindow
	if inp.ContextWindow.TotalInputTokens != 15234 {
		t.Errorf("ContextWindow.TotalInputTokens = %v, want 15234", inp.ContextWindow.TotalInputTokens)
	}
	if inp.ContextWindow.TotalOutputTokens != 4521 {
		t.Errorf("ContextWindow.TotalOutputTokens = %v, want 4521", inp.ContextWindow.TotalOutputTokens)
	}
	if inp.ContextWindow.ContextWindowSize != 200000 {
		t.Errorf("ContextWindow.ContextWindowSize = %v, want 200000", inp.ContextWindow.ContextWindowSize)
	}
	if inp.ContextWindow.UsedPercentage == nil {
		t.Fatal("ContextWindow.UsedPercentage가 nil이면 안 됨")
	}
	if *inp.ContextWindow.UsedPercentage != 28 {
		t.Errorf("ContextWindow.UsedPercentage = %v, want 28", *inp.ContextWindow.UsedPercentage)
	}
	if inp.ContextWindow.RemainingPct == nil {
		t.Fatal("ContextWindow.RemainingPct가 nil이면 안 됨")
	}
	if *inp.ContextWindow.RemainingPct != 72 {
		t.Errorf("ContextWindow.RemainingPct = %v, want 72", *inp.ContextWindow.RemainingPct)
	}

	// Exceeds200K
	if inp.Exceeds200K != false {
		t.Errorf("Exceeds200K = %v, want false", inp.Exceeds200K)
	}

	// RateLimits
	if inp.RateLimits == nil {
		t.Fatal("RateLimits가 nil이면 안 됨")
	}
	if inp.RateLimits.FiveHour == nil {
		t.Fatal("RateLimits.FiveHour가 nil이면 안 됨")
	}
	if inp.RateLimits.FiveHour.UsedPercentage != 12 {
		t.Errorf("RateLimits.FiveHour.UsedPercentage = %v, want 12", inp.RateLimits.FiveHour.UsedPercentage)
	}
	if inp.RateLimits.SevenDay == nil {
		t.Fatal("RateLimits.SevenDay가 nil이면 안 됨")
	}
	if inp.RateLimits.SevenDay.UsedPercentage != 35 {
		t.Errorf("RateLimits.SevenDay.UsedPercentage = %v, want 35", inp.RateLimits.SevenDay.UsedPercentage)
	}

	// Vim
	if inp.Vim == nil {
		t.Fatal("Vim이 nil이면 안 됨")
	}
	if inp.Vim.Mode != "NORMAL" {
		t.Errorf("Vim.Mode = %q, want %q", inp.Vim.Mode, "NORMAL")
	}

	// Agent
	if inp.Agent == nil {
		t.Fatal("Agent가 nil이면 안 됨")
	}
	if inp.Agent.Name != "security-reviewer" {
		t.Errorf("Agent.Name = %q, want %q", inp.Agent.Name, "security-reviewer")
	}
}

func TestParseModelAsString(t *testing.T) {
	data := readTestdata(t, "null_fields.json")
	inp, err := input.Parse(data)
	if err != nil {
		t.Fatalf("Parse 실패: %v", err)
	}

	// model 필드가 문자열인 경우 ID와 DisplayName 모두 해당 문자열로 설정
	if inp.Model.ID != "claude-opus-4-6" {
		t.Errorf("Model.ID = %q, want %q", inp.Model.ID, "claude-opus-4-6")
	}
	if inp.Model.DisplayName != "claude-opus-4-6" {
		t.Errorf("Model.DisplayName = %q, want %q", inp.Model.DisplayName, "claude-opus-4-6")
	}

	// used_percentage가 null인 경우
	if inp.ContextWindow.UsedPercentage != nil {
		t.Errorf("ContextWindow.UsedPercentage가 nil이어야 함, got %v", *inp.ContextWindow.UsedPercentage)
	}

	// 선택적 필드가 nil인지 확인
	if inp.RateLimits != nil {
		t.Error("RateLimits가 nil이어야 함")
	}
	if inp.Vim != nil {
		t.Error("Vim이 nil이어야 함")
	}
	if inp.Agent != nil {
		t.Error("Agent가 nil이어야 함")
	}
}

func TestParseMinimalInput(t *testing.T) {
	data := readTestdata(t, "minimal.json")
	inp, err := input.Parse(data)
	if err != nil {
		t.Fatalf("Parse 실패: %v", err)
	}

	if inp.CWD != "/tmp" {
		t.Errorf("CWD = %q, want %q", inp.CWD, "/tmp")
	}
	if inp.Model.ID != "claude-sonnet-4-6" {
		t.Errorf("Model.ID = %q, want %q", inp.Model.ID, "claude-sonnet-4-6")
	}
	if inp.Model.DisplayName != "Sonnet" {
		t.Errorf("Model.DisplayName = %q, want %q", inp.Model.DisplayName, "Sonnet")
	}
	if inp.Cost.TotalDurationMS != 60000 {
		t.Errorf("Cost.TotalDurationMS = %v, want 60000", inp.Cost.TotalDurationMS)
	}
	if inp.ContextWindow.ContextWindowSize != 200000 {
		t.Errorf("ContextWindow.ContextWindowSize = %v, want 200000", inp.ContextWindow.ContextWindowSize)
	}

	// 선택적 필드들이 nil인지 확인
	if inp.Workspace != nil {
		t.Error("Workspace가 nil이어야 함")
	}
	if inp.OutputStyle != nil {
		t.Error("OutputStyle이 nil이어야 함")
	}
	if inp.RateLimits != nil {
		t.Error("RateLimits가 nil이어야 함")
	}
	if inp.Vim != nil {
		t.Error("Vim이 nil이어야 함")
	}
	if inp.Agent != nil {
		t.Error("Agent가 nil이어야 함")
	}
	if inp.Worktree != nil {
		t.Error("Worktree가 nil이어야 함")
	}
}

func TestParseInvalidJSON(t *testing.T) {
	_, err := input.Parse([]byte(`{invalid json`))
	if err == nil {
		t.Error("잘못된 JSON에서 에러가 반환되어야 함")
	}
}

func TestParseEmptyInput(t *testing.T) {
	_, err := input.Parse([]byte{})
	if err == nil {
		t.Error("빈 입력에서 에러가 반환되어야 함")
	}
}

func TestParseCurrentUsage(t *testing.T) {
	data := readTestdata(t, "normal.json")
	inp, err := input.Parse(data)
	if err != nil {
		t.Fatalf("Parse 실패: %v", err)
	}

	cu := inp.ContextWindow.CurrentUsage
	if cu == nil {
		t.Fatal("CurrentUsage가 nil이면 안 됨")
	}
	if cu.InputTokens != 8500 {
		t.Errorf("InputTokens = %v, want 8500", cu.InputTokens)
	}
	if cu.OutputTokens != 1200 {
		t.Errorf("OutputTokens = %v, want 1200", cu.OutputTokens)
	}
	if cu.CacheCreationInputTokens != 5000 {
		t.Errorf("CacheCreationInputTokens = %v, want 5000", cu.CacheCreationInputTokens)
	}
	if cu.CacheReadInputTokens != 2000 {
		t.Errorf("CacheReadInputTokens = %v, want 2000", cu.CacheReadInputTokens)
	}
}
