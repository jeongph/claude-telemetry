package section

import (
	"fmt"

	"github.com/jeongph/claude-telemetry/internal/render"
)

// LinesSection displays lines added/removed (Line 2, Priority 6).
type LinesSection struct{}

func (s *LinesSection) Name() string  { return "lines" }
func (s *LinesSection) Priority() int { return 6 }

func (s *LinesSection) Render(ctx *Context) string {
	added := ctx.Input.Cost.TotalLinesAdded
	removed := ctx.Input.Cost.TotalLinesRemoved
	if added == 0 && removed == 0 {
		return ""
	}
	return ctx.Colors.Green(fmt.Sprintf("+%d", added)) +
		ctx.Colors.Dim("/") +
		ctx.Colors.Red(fmt.Sprintf("-%d", removed))
}

func (s *LinesSection) Width(ctx *Context) int {
	return render.DisplayWidth(s.Render(ctx))
}
