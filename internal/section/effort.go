package section

import "github.com/jeongph/claude-telemetry/internal/render"

// EffortSection displays the reasoning effort level (Line 1, Priority 3).
// Claude Code 2.1.141+ 의 effort.level 입력을 사용한다 (라이브 세션 값).
type EffortSection struct{}

func (s *EffortSection) Name() string  { return "effort" }
func (s *EffortSection) Priority() int { return 3 }

func (s *EffortSection) Render(ctx *Context) string {
	eff := ctx.Input.Effort
	if eff == nil || eff.Level == "" {
		return ""
	}
	return ctx.Colors.Dim("↯") + effortColor(ctx.Colors, eff.Level)(eff.Level)
}

// effortColor는 effort 레벨별 색상 함수를 반환한다 (높을수록 강조).
func effortColor(c render.Colors, level string) func(string) string {
	switch level {
	case "low":
		return c.Dim
	case "medium":
		return c.White
	case "high":
		return c.Green
	case "xhigh":
		return c.Yellow
	case "max":
		return c.Red
	default:
		return c.White
	}
}

func (s *EffortSection) Width(ctx *Context) int {
	return render.DisplayWidth(s.Render(ctx))
}
