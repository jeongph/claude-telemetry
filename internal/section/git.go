package section

import (
	"fmt"
	"path/filepath"

	"github.com/jeongph/claude-telemetry/internal/render"
)

// GitSection displays git repository info (Line 1, Priority 4).
type GitSection struct{}

func (s *GitSection) Name() string  { return "git" }
func (s *GitSection) Priority() int { return 4 }

func (s *GitSection) Render(ctx *Context) string {
	folder := filepath.Base(ctx.Input.CWD)
	if folder == "" || folder == "." {
		folder = ctx.Input.CWD
	}

	gi := ctx.GitInfo
	if gi == nil || !gi.IsRepo {
		return ctx.Colors.White(folder)
	}

	out := ctx.Colors.White(folder) + ctx.Colors.Dim(":") + ctx.Colors.Magenta(gi.Branch)

	if gi.Ahead > 0 {
		out += " " + ctx.Colors.Yellow(fmt.Sprintf("↑%d", gi.Ahead))
	}
	if gi.Behind > 0 {
		out += ctx.Colors.Cyan(fmt.Sprintf("↓%d", gi.Behind))
	}

	if gi.Added > 0 || gi.Deleted > 0 {
		out += " " + ctx.Colors.Green(fmt.Sprintf("+%d", gi.Added)) +
			ctx.Colors.Dim("/") +
			ctx.Colors.Red(fmt.Sprintf("-%d", gi.Deleted))
	}

	if gi.Untracked > 0 {
		out += " " + ctx.Colors.Dim("?") + ctx.Colors.Yellow(fmt.Sprintf("%d", gi.Untracked))
	}

	if gi.Stash > 0 {
		out += " " + ctx.Colors.Dim("≡") + ctx.Colors.Magenta(fmt.Sprintf("%d", gi.Stash))
	}

	if gi.Worktrees > 0 {
		out += " " + ctx.Colors.Dim("⎇") + ctx.Colors.Cyan(fmt.Sprintf("%d", gi.Worktrees))
	}

	return out
}

func (s *GitSection) Width(ctx *Context) int {
	return render.DisplayWidth(s.Render(ctx))
}
