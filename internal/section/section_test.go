package section

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jeongph/claude-telemetry/internal/config"
	"github.com/jeongph/claude-telemetry/internal/gitinfo"
	"github.com/jeongph/claude-telemetry/internal/i18n"
	"github.com/jeongph/claude-telemetry/internal/input"
	"github.com/jeongph/claude-telemetry/internal/render"
)

func testContext(t *testing.T) *Context {
	t.Helper()
	data, err := os.ReadFile("../../testdata/normal.json")
	if err != nil {
		t.Fatalf("testdata 읽기 실패: %v", err)
	}
	inp, err := input.Parse(data)
	if err != nil {
		t.Fatalf("input 파싱 실패: %v", err)
	}
	return &Context{
		Input:   inp,
		Config:  config.Load("", ""),
		Locale:  i18n.New("en"),
		Colors:  render.NewColors(false), // no color for assertion
		GitInfo: &gitinfo.GitInfo{IsRepo: true, Branch: "main", Ahead: 1, Added: 5, Deleted: 2},
		Effort:  "high",
	}
}

func TestModelSection(t *testing.T) {
	ctx := testContext(t)
	s := &ModelSection{}
	got := s.Render(ctx)
	if !strings.Contains(got, "Opus") {
		t.Errorf("ModelSection: 모델명 'Opus' 누락 — got %q", got)
	}
	if strings.Contains(got, "high") {
		t.Errorf("ModelSection: effort는 표시하지 않아야 함 — got %q", got)
	}
}

func TestElapsedSection(t *testing.T) {
	ctx := testContext(t)
	// normal.json has total_duration_ms: 754000 → 12m 34s
	s := &ElapsedSection{}
	got := s.Render(ctx)
	if !strings.Contains(got, "12m") {
		t.Errorf("ElapsedSection: '12m' 누락 — got %q", got)
	}
	if !strings.Contains(got, "34s") {
		t.Errorf("ElapsedSection: '34s' 누락 — got %q", got)
	}
}

func TestContextSection(t *testing.T) {
	ctx := testContext(t)
	// normal.json: used_percentage=28 → remaining=72, size=200000 → "200k"
	s := &ContextSection{}
	got := s.Render(ctx)
	if !strings.Contains(got, "72") {
		t.Errorf("ContextSection: 남은 비율 '72' 누락 — got %q", got)
	}
	if !strings.Contains(got, "200k") {
		t.Errorf("ContextSection: 컨텍스트 크기 '200k' 누락 — got %q", got)
	}
}

func TestContextSectionNull(t *testing.T) {
	ctx := testContext(t)
	// nil UsedPercentage → 새 세션, 100% 표시
	ctx.Input.ContextWindow.UsedPercentage = nil
	s := &ContextSection{}
	got := s.Render(ctx)
	if !strings.Contains(got, "100%") {
		t.Errorf("ContextSection(nil): 새 세션은 100%% 표시해야 함 — got %q", got)
	}
}

func TestRateLimitSectionExpired(t *testing.T) {
	ctx := testContext(t)
	// normal.json의 resets_at은 과거 → 리셋됨 표시
	s := &RateLimitSection{}
	got := s.Render(ctx)
	if !strings.Contains(got, "5h") {
		t.Errorf("RateLimitSection: '5h' 누락 — got %q", got)
	}
	if !strings.Contains(got, "···") {
		t.Errorf("RateLimitSection: 리셋 후 '···' 표시 필요 — got %q", got)
	}
}

func TestRateLimitSectionActive(t *testing.T) {
	ctx := testContext(t)
	// resets_at을 미래로 설정
	futureTS := float64(time.Now().Add(2 * time.Hour).Unix())
	ctx.Input.RateLimits.FiveHour.ResetsAt = futureTS
	ctx.Input.RateLimits.SevenDay.ResetsAt = float64(time.Now().Add(4 * 24 * time.Hour).Unix())
	s := &RateLimitSection{}
	got := s.Render(ctx)
	if !strings.Contains(got, "%") {
		t.Errorf("RateLimitSection: 활성 상태에서 '%%' 필요 — got %q", got)
	}
}

func TestCostSection(t *testing.T) {
	ctx := testContext(t)
	// set rate_limits = nil to show cost (API key user)
	ctx.Input.RateLimits = nil
	s := &CostSection{}
	got := s.Render(ctx)
	if !strings.Contains(got, "0.45") {
		t.Errorf("CostSection: '$0.45' 누락 — got %q", got)
	}
}

func TestGitSection(t *testing.T) {
	ctx := testContext(t)
	s := &GitSection{}
	got := s.Render(ctx)
	if !strings.Contains(got, "main") {
		t.Errorf("GitSection: branch 'main' 누락 — got %q", got)
	}
	if !strings.Contains(got, "+5") {
		t.Errorf("GitSection: '+5' 누락 — got %q", got)
	}
}

func TestGitSectionNoRepo(t *testing.T) {
	ctx := testContext(t)
	ctx.GitInfo = &gitinfo.GitInfo{IsRepo: false}
	// normal.json cwd = "/home/user/myproject" → folder "myproject"
	s := &GitSection{}
	got := s.Render(ctx)
	if !strings.Contains(got, "myproject") {
		t.Errorf("GitSection(NoRepo): 폴더명 'myproject' 누락 — got %q", got)
	}
}

func TestAgentSection(t *testing.T) {
	ctx := testContext(t)
	// normal.json has agent.name = "security-reviewer"
	s := &AgentSection{}
	got := s.Render(ctx)
	if !strings.Contains(got, "security-reviewer") {
		t.Errorf("AgentSection: 에이전트명 'security-reviewer' 누락 — got %q", got)
	}
}

func TestAgentSectionNil(t *testing.T) {
	ctx := testContext(t)
	ctx.Input.Agent = nil
	s := &AgentSection{}
	got := s.Render(ctx)
	if got != "" {
		t.Errorf("AgentSection(nil): 빈 문자열 기대, got %q", got)
	}
}

func TestVimSection(t *testing.T) {
	ctx := testContext(t)
	// normal.json has vim.mode = "NORMAL"
	s := &VimSection{}
	got := s.Render(ctx)
	if !strings.Contains(got, "NORMAL") {
		t.Errorf("VimSection: 'NORMAL' 누락 — got %q", got)
	}
}

func TestVimSectionNil(t *testing.T) {
	ctx := testContext(t)
	ctx.Input.Vim = nil
	s := &VimSection{}
	got := s.Render(ctx)
	if got != "" {
		t.Errorf("VimSection(nil): 빈 문자열 기대, got %q", got)
	}
}

func TestLinesSection(t *testing.T) {
	ctx := testContext(t)
	// normal.json: total_lines_added=156, total_lines_removed=23
	s := &LinesSection{}
	got := s.Render(ctx)
	if !strings.Contains(got, "+156") {
		t.Errorf("LinesSection: '+156' 누락 — got %q", got)
	}
	if !strings.Contains(got, "-23") {
		t.Errorf("LinesSection: '-23' 누락 — got %q", got)
	}
}

func TestTokensSection(t *testing.T) {
	ctx := testContext(t)
	// normal.json: input=15234 → "15k", output=4521 → "4k"
	s := &TokensSection{}
	got := s.Render(ctx)
	if !strings.Contains(got, "15k") {
		t.Errorf("TokensSection: '15k' 누락 — got %q", got)
	}
	if !strings.Contains(got, "4k") {
		t.Errorf("TokensSection: '4k' 누락 — got %q", got)
	}
}

func TestAllSectionsRegistry(t *testing.T) {
	sections := AllSections()
	if len(sections) != 11 {
		t.Errorf("AllSections: 11개 기대, got %d", len(sections))
	}
}

func TestWidthEqualsDisplayWidth(t *testing.T) {
	ctx := testContext(t)
	s := &ModelSection{}
	rendered := s.Render(ctx)
	if s.Width(ctx) != render.DisplayWidth(rendered) {
		t.Errorf("Width: DisplayWidth와 불일치")
	}
}
