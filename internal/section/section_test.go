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
	}
}

func TestModelSection(t *testing.T) {
	ctx := testContext(t)
	// normal.json: effort.level = "high" → 모델명 옆에 "· high" 표시
	s := &ModelSection{}
	got := s.Render(ctx)
	if !strings.Contains(got, "Opus") {
		t.Errorf("ModelSection: 모델명 'Opus' 누락 — got %q", got)
	}
	if !strings.Contains(got, "· high") {
		t.Errorf("ModelSection: effort '· high' 누락 — got %q", got)
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
	// normal.json: used_percentage=28 → remaining=72
	s := &ContextSection{}
	got := s.Render(ctx)
	if !strings.Contains(got, "72") {
		t.Errorf("ContextSection: 남은 비율 '72' 누락 — got %q", got)
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

func TestContextSectionNoBar(t *testing.T) {
	ctx := testContext(t)
	ctx.Config.BarWidth = 0 // 바 없이 % 만
	s := &ContextSection{}
	got := s.Render(ctx)
	if strings.ContainsAny(got, "▰▱") {
		t.Errorf("bar_width 0: 바 문자가 없어야 함 — got %q", got)
	}
	if strings.Contains(got, "  ") {
		t.Errorf("bar_width 0: 이중 공백 없어야 함 — got %q", got)
	}
	if !strings.Contains(got, "72%") {
		t.Errorf("bar_width 0: %% 값은 유지해야 함 — got %q", got)
	}
}

func TestRateLimitSectionNoBar(t *testing.T) {
	ctx := testContext(t)
	ctx.Config.BarWidth = 0 // 바 없이 % 만
	// 활성 상태로 만들어 바가 그려질 조건 확보
	ctx.Input.RateLimits.FiveHour.ResetsAt = float64(time.Now().Add(2 * time.Hour).Unix())
	ctx.Input.RateLimits.SevenDay.ResetsAt = float64(time.Now().Add(4 * 24 * time.Hour).Unix())
	s := &RateLimitSection{}
	got := s.Render(ctx)
	if strings.ContainsAny(got, "▰▱") {
		t.Errorf("bar_width 0: 바 문자가 없어야 함 — got %q", got)
	}
	if strings.Contains(got, "  ") {
		t.Errorf("bar_width 0: 이중 공백 없어야 함 — got %q", got)
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

func TestFormatDuration(t *testing.T) {
	noColor := render.NewColors(false)
	tests := []struct {
		name     string
		ms       float64
		contains []string
	}{
		{"초 단위", 45000, []string{"45", "s"}},
		{"분+초", 754000, []string{"12", "m", "34", "s"}},
		{"시+분", 7500000, []string{"2", "h", "5", "m"}},
		{"일+시 (24h 이상)", 367440000, []string{"4", "d", "6", "h"}},
		{"정확히 24h", 86400000, []string{"1", "d", "0", "h"}},
		{"48h", 172800000, []string{"2", "d", "0", "h"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := formatDuration(tc.ms, noColor)
			for _, want := range tc.contains {
				if !strings.Contains(got, want) {
					t.Errorf("formatDuration(%v): %q 누락 — got %q", tc.ms, want, got)
				}
			}
		})
	}
}

func TestFormatDurationNoDaysBelowThreshold(t *testing.T) {
	noColor := render.NewColors(false)
	// 23h 59m → "d" 가 포함되면 안 됨
	got := formatDuration(86340000, noColor)
	if strings.Contains(got, "d") {
		t.Errorf("24h 미만인데 'd' 표시됨 — got %q", got)
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

func TestModelSectionEffortNil(t *testing.T) {
	ctx := testContext(t)
	// effort 미지원 모델(필드 absent) → 모델명만 표시
	ctx.Input.Effort = nil
	s := &ModelSection{}
	got := s.Render(ctx)
	if !strings.Contains(got, "Opus") {
		t.Errorf("ModelSection(effort nil): 모델명 'Opus' 누락 — got %q", got)
	}
	if strings.Contains(got, "·") {
		t.Errorf("ModelSection(effort nil): '·'가 없어야 함 — got %q", got)
	}
}

func TestModelSectionEffortDisabled(t *testing.T) {
	ctx := testContext(t)
	// sections override로 effort를 끄면 모델명만 표시
	ctx.Config.Sections = map[string]bool{"effort": false}
	s := &ModelSection{}
	got := s.Render(ctx)
	if !strings.Contains(got, "Opus") {
		t.Errorf("ModelSection(effort off): 모델명 'Opus' 누락 — got %q", got)
	}
	if strings.Contains(got, "·") {
		t.Errorf("ModelSection(effort off): '·'가 없어야 함 — got %q", got)
	}
}

func TestSessionSection(t *testing.T) {
	ctx := testContext(t)
	// normal.json: session_name = "my-session"
	s := &SessionSection{}
	got := s.Render(ctx)
	if !strings.Contains(got, "my-session") {
		t.Errorf("SessionSection: 세션명 'my-session' 누락 — got %q", got)
	}
}

func TestSessionSectionEmpty(t *testing.T) {
	ctx := testContext(t)
	ctx.Input.SessionName = ""
	s := &SessionSection{}
	if got := s.Render(ctx); got != "" {
		t.Errorf("SessionSection(empty): 빈 문자열 기대, got %q", got)
	}
}

func TestSessionSectionTruncate(t *testing.T) {
	ctx := testContext(t)
	// 실측: CC 2.1.170은 자동 생성된 긴 세션 제목을 session_name으로 보냄
	ctx.Input.SessionName = "Claude 2.1.170 업데이트로 telemetry 고도화 검토"
	s := &SessionSection{}
	got := s.Render(ctx)
	if !strings.Contains(got, "…") {
		t.Errorf("SessionSection(long): 말줄임 '…' 누락 — got %q", got)
	}
	// [ + 내용 + ] 전체가 24컬럼 이하 (내용 최대 20 + 괄호 2 + 말줄임 1)
	if w := render.DisplayWidth(got); w > 24 {
		t.Errorf("SessionSection(long): 표시 폭 %d > 24", w)
	}
}

func TestSessionSectionExactWidth(t *testing.T) {
	ctx := testContext(t)
	ctx.Input.SessionName = "12345678901234567890" // 20 ASCII = 20컬럼, 절단 불필요
	s := &SessionSection{}
	got := s.Render(ctx)
	if strings.Contains(got, "…") {
		t.Errorf("SessionSection(exact20): 절단되면 안 됨 — got %q", got)
	}
}

func TestPRSection(t *testing.T) {
	ctx := testContext(t)
	// normal.json: pr.number = 1234, review_state = "approved"
	s := &PRSection{}
	got := s.Render(ctx)
	if !strings.Contains(got, "PR#1234") {
		t.Errorf("PRSection: 'PR#1234' 누락 — got %q", got)
	}
	if !strings.Contains(got, "✓") {
		t.Errorf("PRSection: approved 마크 '✓' 누락 — got %q", got)
	}
}

func TestPRSectionNil(t *testing.T) {
	ctx := testContext(t)
	ctx.Input.PR = nil
	s := &PRSection{}
	if got := s.Render(ctx); got != "" {
		t.Errorf("PRSection(nil): 빈 문자열 기대, got %q", got)
	}
}

func TestPRSectionMarkBeforeNumber(t *testing.T) {
	// 리뷰 상태 기호(◌/●/✓/✗)가 번호 뒤에 붙으면 0·o와 헷갈리므로
	// 기호는 반드시 번호 앞에 온다: "✓ PR#1234"
	ctx := testContext(t) // normal.json: approved → ✓
	got := render.StripANSI((&PRSection{}).Render(ctx))
	if !strings.HasPrefix(got, "✓") {
		t.Errorf("PRSection: 리뷰 상태 기호가 맨 앞에 와야 함 — got %q", got)
	}
	if strings.Index(got, "✓") > strings.Index(got, "PR#") {
		t.Errorf("PRSection: 기호가 PR 번호 뒤에 있음 — got %q", got)
	}
}

func TestPRSectionNoReviewState(t *testing.T) {
	ctx := testContext(t)
	ctx.Input.PR.ReviewState = ""
	s := &PRSection{}
	got := s.Render(ctx)
	if !strings.Contains(got, "PR#1234") {
		t.Errorf("PRSection(state 없음): 'PR#1234' 누락 — got %q", got)
	}
}

func TestPRSectionReviewStates(t *testing.T) {
	tests := []struct {
		state string
		mark  string
	}{
		{"approved", "✓"},
		{"pending", "●"},
		{"changes_requested", "✗"},
		{"draft", "◌"},
	}
	for _, tc := range tests {
		t.Run(tc.state, func(t *testing.T) {
			ctx := testContext(t)
			ctx.Input.PR.ReviewState = tc.state
			s := &PRSection{}
			got := s.Render(ctx)
			if !strings.Contains(got, tc.mark) {
				t.Errorf("PRSection(%s): 마크 %q 누락 — got %q", tc.state, tc.mark, got)
			}
		})
	}
}

func TestThinkingSection(t *testing.T) {
	ctx := testContext(t)
	// normal.json: thinking.enabled = true
	s := &ThinkingSection{}
	got := s.Render(ctx)
	if !strings.Contains(got, "✦") {
		t.Errorf("ThinkingSection: '✦' 누락 — got %q", got)
	}
}

func TestThinkingSectionDisabled(t *testing.T) {
	ctx := testContext(t)
	ctx.Input.Thinking = &input.Thinking{Enabled: false}
	s := &ThinkingSection{}
	if got := s.Render(ctx); got != "" {
		t.Errorf("ThinkingSection(disabled): 빈 문자열 기대, got %q", got)
	}
}

func TestThinkingSectionNil(t *testing.T) {
	ctx := testContext(t)
	ctx.Input.Thinking = nil
	s := &ThinkingSection{}
	if got := s.Render(ctx); got != "" {
		t.Errorf("ThinkingSection(nil): 빈 문자열 기대, got %q", got)
	}
}

func TestAllSectionsRegistry(t *testing.T) {
	sections := AllSections()
	if len(sections) != 15 {
		t.Errorf("AllSections: 15개 기대, got %d", len(sections))
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
