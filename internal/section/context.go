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

	if cw.UsedPercentage == nil {
		return prefix + " " + ctx.Colors.Dim("···")
	}

	remaining := 100.0 - *cw.UsedPercentage

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

	size := formatContextSize(cw.ContextWindowSize)
	var sizeStr string
	if ctx.Input.Exceeds200K {
		sizeStr = ctx.Colors.Yellow(size)
	} else {
		sizeStr = ctx.Colors.Dim(size)
	}

	return prefix + " " + bar + " " + pctStr + " (" + sizeStr + ")"
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
