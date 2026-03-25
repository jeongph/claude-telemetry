package section

import (
	"fmt"

	"github.com/jeongph/claude-telemetry/internal/render"
)

// ContextSection displays the context window usage (Line 2, Priority 1).
type ContextSection struct{}

func (s *ContextSection) Name() string  { return "context" }
func (s *ContextSection) Priority() int { return 1 }

func (s *ContextSection) Render(ctx *Context) string {
	cw := ctx.Input.ContextWindow
	label := ctx.Locale.Get("context")
	prefix := ctx.Colors.Cyan("◆") + " " + ctx.Colors.Dim(label)

	var remaining float64
	if cw.UsedPercentage == nil {
		remaining = 100.0 // 새 세션: 아직 미사용
	} else {
		remaining = 100.0 - *cw.UsedPercentage
	}

	bar := render.ProgressBarRemaining(
		remaining,
		ctx.Config.BarWidth,
		ctx.Colors,
		ctx.Config.Thresholds.ContextWarn,
		ctx.Config.Thresholds.ContextDanger,
	)

	colorName := render.ThresholdColorRemaining(
		remaining,
		ctx.Colors,
		ctx.Config.Thresholds.ContextWarn,
		ctx.Config.Thresholds.ContextDanger,
	)
	pctStr := render.ApplyColor(colorName, fmt.Sprintf("%.0f%%", remaining), ctx.Colors)

	return prefix + " " + bar + " " + pctStr
}

func (s *ContextSection) Width(ctx *Context) int {
	return render.DisplayWidth(s.Render(ctx))
}

// formatContextSize converts a raw token count to a human-readable size string.
// 200000 → "200k", 1000000 → "1M"
func formatContextSize(size int) string {
	if size >= 1_000_000 {
		return fmt.Sprintf("%dM", size/1_000_000)
	}
	return fmt.Sprintf("%dk", size/1000)
}
