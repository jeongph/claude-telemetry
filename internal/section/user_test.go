package section

import (
	"strings"
	"testing"

	"github.com/jeongph/claude-telemetry/internal/account"
	"github.com/jeongph/claude-telemetry/internal/config"
	"github.com/jeongph/claude-telemetry/internal/i18n"
	"github.com/jeongph/claude-telemetry/internal/input"
	"github.com/jeongph/claude-telemetry/internal/render"
)

func userContext(acc *account.Account) *Context {
	return &Context{
		Input:   &input.Input{},
		Config:  config.Load("", ""),
		Locale:  i18n.New("en"),
		Colors:  render.NewColors(false), // 색 없이 문자열만 검증
		Account: acc,
	}
}

func TestUserSectionRendersEmailAndPlan(t *testing.T) {
	ctx := userContext(&account.Account{Email: "user@example.com", PlanType: "claude_max"})
	got := (&UserSection{}).Render(ctx)
	if !strings.Contains(got, "user@example.com") {
		t.Errorf("이메일 누락 — got %q", got)
	}
	if !strings.Contains(got, "Max") {
		t.Errorf("플랜 축약 'Max' 누락 — got %q", got)
	}
	if strings.Contains(got, "claude_max") {
		t.Errorf("raw 플랜 타입이 그대로 노출됨 — got %q", got)
	}
}

func TestUserSectionEmptyWhenNoAccount(t *testing.T) {
	if got := (&UserSection{}).Render(userContext(nil)); got != "" {
		t.Errorf("계정 없음인데 렌더됨 — got %q", got)
	}
}

func TestUserSectionRendersEmailWhenPlanUnknown(t *testing.T) {
	// 미지의 플랜 타입은 생략하고 이메일만 표시한다.
	ctx := userContext(&account.Account{Email: "a@b.com", PlanType: "weird_unknown"})
	got := (&UserSection{}).Render(ctx)
	if !strings.Contains(got, "a@b.com") {
		t.Errorf("이메일 누락 — got %q", got)
	}
	if strings.Contains(got, "weird_unknown") {
		t.Errorf("미지 플랜 타입이 노출됨 — got %q", got)
	}
}

func TestUserSectionName(t *testing.T) {
	if got := (&UserSection{}).Name(); got != "user" {
		t.Errorf("Name: got %q, want %q", got, "user")
	}
}

func TestPlanLabel(t *testing.T) {
	cases := map[string]string{
		"claude_max":        "Max",
		"claude_pro":        "Pro",
		"claude_team":       "Team",
		"claude_enterprise": "Enterprise",
		"":                  "",
		"weird_unknown":     "",
	}
	for in, want := range cases {
		if got := planLabel(in); got != want {
			t.Errorf("planLabel(%q) = %q, want %q", in, got, want)
		}
	}
}
