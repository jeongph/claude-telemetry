package section

import (
	"fmt"

	"github.com/jeongph/claude-telemetry/internal/render"
)

// ElapsedSection displays the total elapsed duration (Line 1, Priority 5).
type ElapsedSection struct{}

func (s *ElapsedSection) Name() string  { return "elapsed" }
func (s *ElapsedSection) Priority() int { return 5 }

func (s *ElapsedSection) Render(ctx *Context) string {
	ms := ctx.Input.Cost.TotalDurationMS
	if ms == 0 {
		return ""
	}
	label := ctx.Locale.Get("elapsed")
	return ctx.Colors.Dim("◷ "+label) + " " + formatDuration(ms, ctx.Colors)
}

func (s *ElapsedSection) Width(ctx *Context) int {
	return render.DisplayWidth(s.Render(ctx))
}

// formatDuration converts milliseconds to a human-readable string with
// ANSI coloring: numbers in white, units in dim.
func formatDuration(ms float64, c render.Colors) string {
	totalSeconds := int(ms / 1000)
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60

	white := func(n int) string { return c.White(fmt.Sprintf("%d", n)) }
	dim := func(u string) string { return c.Dim(u) }

	if ms >= 3600000 {
		return white(hours) + dim("h") + " " + white(minutes) + dim("m")
	}
	if ms >= 60000 {
		return white(minutes) + dim("m") + " " + white(seconds) + dim("s")
	}
	return white(seconds) + dim("s")
}
