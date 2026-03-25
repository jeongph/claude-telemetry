package section

import "github.com/jeongph/claude-telemetry/internal/render"

// VimSection displays the current Vim mode (Line 3, Priority 7).
type VimSection struct{}

func (s *VimSection) Name() string  { return "vim" }
func (s *VimSection) Priority() int { return 7 }

func (s *VimSection) Render(ctx *Context) string {
	if ctx.Input.Vim == nil {
		return ""
	}
	mode := ctx.Input.Vim.Mode
	switch mode {
	case "NORMAL":
		return ctx.Colors.Cyan(mode)
	case "INSERT":
		return ctx.Colors.Green(mode)
	default:
		return ctx.Colors.White(mode)
	}
}

func (s *VimSection) Width(ctx *Context) int {
	return render.DisplayWidth(s.Render(ctx))
}
