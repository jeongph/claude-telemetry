package section

import (
	"fmt"

	"github.com/jeongph/claude-telemetry/internal/render"
)

// PRSection displays the open pull request for the current branch (Line 1, Priority 5).
// Claude Code 2.1.145+ 의 pr 입력을 사용한다. 열린 PR이 없으면 표시하지 않는다.
type PRSection struct{}

func (s *PRSection) Name() string  { return "pr" }
func (s *PRSection) Priority() int { return 5 }

func (s *PRSection) Render(ctx *Context) string {
	pr := ctx.Input.PR
	if pr == nil || pr.Number == 0 {
		return ""
	}
	out := ctx.Colors.Cyan(fmt.Sprintf("PR#%d", pr.Number))
	switch pr.ReviewState {
	case "approved":
		out += ctx.Colors.Green("✓")
	case "pending":
		out += ctx.Colors.Yellow("●")
	case "changes_requested":
		out += ctx.Colors.Red("✗")
	case "draft":
		out += ctx.Colors.Dim("◌")
	}
	return out
}

func (s *PRSection) Width(ctx *Context) int {
	return render.DisplayWidth(s.Render(ctx))
}
