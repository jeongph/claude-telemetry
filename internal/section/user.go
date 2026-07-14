package section

import "github.com/jeongph/claude-telemetry/internal/render"

// UserSection은 로그인 계정의 이메일과 플랜을 표시한다 (전용 줄, Line 4).
// 이메일은 상태 라인 stdin JSON에 없어 ~/.claude.json에서 읽어오며(Context.Account),
// 민감 정보이므로 기본 off(opt-in)다 — config에서 "user" 섹션을 켜야 나타난다.
// 예: "◉ user@example.com · Max"
type UserSection struct{}

func (s *UserSection) Name() string  { return "user" }
func (s *UserSection) Priority() int { return 8 }

func (s *UserSection) Render(ctx *Context) string {
	if ctx.Account == nil || ctx.Account.Email == "" {
		return ""
	}
	out := ctx.Colors.Dim("◉") + " " + ctx.Colors.Cyan(ctx.Account.Email)
	if plan := planLabel(ctx.Account.PlanType); plan != "" {
		out += " " + ctx.Colors.Dim("·") + " " + ctx.Colors.Magenta(plan)
	}
	return out
}

func (s *UserSection) Width(ctx *Context) int {
	return render.DisplayWidth(s.Render(ctx))
}

// planLabel은 organizationType 원시 값을 표시용 짧은 이름으로 매핑한다.
// 알 수 없는 값은 빈 문자열을 반환해 raw 값이 노출되지 않도록 한다.
func planLabel(orgType string) string {
	switch orgType {
	case "claude_max":
		return "Max"
	case "claude_pro":
		return "Pro"
	case "claude_team":
		return "Team"
	case "claude_enterprise":
		return "Enterprise"
	default:
		return ""
	}
}
