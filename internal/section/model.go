package section

import "github.com/jeongph/claude-telemetry/internal/render"

// ModelSection displays the model name and effort level (Line 1, Priority 3).
type ModelSection struct{}

func (s *ModelSection) Name() string     { return "model" }
func (s *ModelSection) Priority() int    { return 3 }

func (s *ModelSection) Render(ctx *Context) string {
	name := ctx.Input.Model.DisplayName
	if name == "" {
		return ""
	}
	out := ctx.Colors.Cyan(name)
	if ctx.Effort != "" {
		out += " " + ctx.Colors.Dim("💭"+ctx.Effort)
	}
	return out
}

func (s *ModelSection) Width(ctx *Context) int {
	return render.DisplayWidth(s.Render(ctx))
}
