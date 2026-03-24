package section

import (
	"fmt"

	"github.com/jeongph/claude-telemetry/internal/render"
)

// CostSection displays the total cost in USD (Line 2, Priority 2).
// Only shown for API key users (when rate_limits is nil).
type CostSection struct{}

func (s *CostSection) Name() string  { return "cost" }
func (s *CostSection) Priority() int { return 2 }

func (s *CostSection) Render(ctx *Context) string {
	cost := ctx.Input.Cost.TotalCostUSD
	if cost == 0 {
		return ""
	}

	label := ctx.Locale.Get("cost")
	costStr := fmt.Sprintf("$%.2f", cost)

	var colorName string
	if cost < ctx.Config.Thresholds.CostWarn {
		colorName = "green"
	} else if cost < ctx.Config.Thresholds.CostDanger {
		colorName = "yellow"
	} else {
		colorName = "red"
	}

	colored := render.ApplyColor(colorName, costStr, ctx.Colors)
	return ctx.Colors.Dim(label) + " " + colored
}

func (s *CostSection) Width(ctx *Context) int {
	return render.DisplayWidth(s.Render(ctx))
}
