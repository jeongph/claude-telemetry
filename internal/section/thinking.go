package section

import "github.com/jeongph/claude-telemetry/internal/render"

// ThinkingSection displays the extended thinking indicator (Line 3, Priority 8).
// thinking.enabled가 true일 때만 표시된다.
type ThinkingSection struct{}

func (s *ThinkingSection) Name() string  { return "thinking" }
func (s *ThinkingSection) Priority() int { return 8 }

func (s *ThinkingSection) Render(ctx *Context) string {
	th := ctx.Input.Thinking
	if th == nil || !th.Enabled {
		return ""
	}
	return ctx.Colors.Magenta("✦") + " " + ctx.Colors.Dim(ctx.Locale.Get("thinking"))
}

func (s *ThinkingSection) Width(ctx *Context) int {
	return render.DisplayWidth(s.Render(ctx))
}
