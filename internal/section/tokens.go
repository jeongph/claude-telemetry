package section

import (
	"fmt"

	"github.com/jeongph/claude-telemetry/internal/render"
)

// TokensSection displays input/output token counts (Line 2, Priority 9).
type TokensSection struct{}

func (s *TokensSection) Name() string  { return "tokens" }
func (s *TokensSection) Priority() int { return 9 }

func (s *TokensSection) Render(ctx *Context) string {
	cw := ctx.Input.ContextWindow
	if cw.TotalInputTokens == 0 {
		return ""
	}
	inLabel := ctx.Locale.Get("in")
	outLabel := ctx.Locale.Get("out")
	return ctx.Colors.Dim(inLabel) + " " +
		ctx.Colors.White(formatTokens(cw.TotalInputTokens)) + " " +
		ctx.Colors.Dim(outLabel) + " " +
		ctx.Colors.White(formatTokens(cw.TotalOutputTokens))
}

func (s *TokensSection) Width(ctx *Context) int {
	return render.DisplayWidth(s.Render(ctx))
}

// formatTokens formats a token count as "1.5M", "15k", or "123".
func formatTokens(n int) string {
	if n >= 1_000_000 {
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	}
	if n >= 1000 {
		return fmt.Sprintf("%dk", n/1000)
	}
	return fmt.Sprintf("%d", n)
}
