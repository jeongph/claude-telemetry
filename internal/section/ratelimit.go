package section

import (
	"fmt"
	"time"

	"github.com/jeongph/claude-telemetry/internal/render"
)

// RateLimitSection displays the 5h and 7d rate limit windows (Line 2, Priority 2).
type RateLimitSection struct{}

func (s *RateLimitSection) Name() string  { return "ratelimit" }
func (s *RateLimitSection) Priority() int { return 2 }

func (s *RateLimitSection) Render(ctx *Context) string {
	rl := ctx.Input.RateLimits
	if rl == nil {
		return ""
	}

	var parts []string

	if rl.FiveHour != nil {
		parts = append(parts, renderRateWindow(rl.FiveHour.UsedPercentage, rl.FiveHour.ResetsAt, "5h", ctx))
	} else {
		parts = append(parts, ctx.Colors.Dim("5h")+" "+ctx.Colors.Dim("···"))
	}

	if rl.SevenDay != nil {
		parts = append(parts, renderRateWindow(rl.SevenDay.UsedPercentage, rl.SevenDay.ResetsAt, "7d", ctx))
	} else {
		parts = append(parts, ctx.Colors.Dim("7d")+" "+ctx.Colors.Dim("···"))
	}

	result := ""
	for i, p := range parts {
		if i > 0 {
			result += " "
		}
		result += p
	}
	return result
}

func (s *RateLimitSection) Width(ctx *Context) int {
	return render.DisplayWidth(s.Render(ctx))
}

// renderRateWindow renders a single rate limit window.
// remaining = 100 - used_percentage (stdin gives used, we show remaining).
func renderRateWindow(usedPct float64, resetsAt float64, label string, ctx *Context) string {
	// resets_at이 과거 → 리셋됨, 정확한 값 알 수 없음
	if resetsAt > 0 {
		resetTime := time.Unix(int64(resetsAt), 0)
		if time.Now().After(resetTime) {
			emptyBar := render.ProgressBarRemaining(0, ctx.Config.BarWidth, ctx.Colors, 50, 20)
			return ctx.Colors.Dim(label) + " " + emptyBar + " " + ctx.Colors.Dim("···")
		}
	}

	remaining := 100.0 - usedPct

	countdown := ""
	if resetsAt > 0 {
		diff := time.Until(time.Unix(int64(resetsAt), 0))
		if diff > 0 {
			countdown = formatDuration(float64(diff.Milliseconds()), ctx.Colors)
		}
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

	labelStr := ctx.Colors.Dim(label)

	if countdown != "" {
		return countdown + "/" + labelStr + " " + bar + " " + pctStr
	}
	return labelStr + " " + bar + " " + pctStr
}
