package section

import "github.com/jeongph/claude-telemetry/internal/render"

// APIDurationSection displays the total API wait duration (Line 2, Priority 8).
type APIDurationSection struct{}

func (s *APIDurationSection) Name() string  { return "apiduration" }
func (s *APIDurationSection) Priority() int { return 8 }

func (s *APIDurationSection) Render(ctx *Context) string {
	ms := ctx.Input.Cost.TotalAPIDurationMS
	if ms == 0 {
		return ""
	}
	label := ctx.Locale.Get("api")
	return ctx.Colors.Dim("↻ "+label) + " " + formatDuration(ms, ctx.Colors)
}

func (s *APIDurationSection) Width(ctx *Context) int {
	return render.DisplayWidth(s.Render(ctx))
}
