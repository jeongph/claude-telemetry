package section

import (
	"github.com/jeongph/claude-telemetry/internal/config"
	"github.com/jeongph/claude-telemetry/internal/gitinfo"
	"github.com/jeongph/claude-telemetry/internal/i18n"
	"github.com/jeongph/claude-telemetry/internal/input"
	"github.com/jeongph/claude-telemetry/internal/render"
)

// Context holds all data needed by sections when rendering.
type Context struct {
	Input   *input.Input
	Config  config.Config
	Locale  i18n.Locale
	Colors  render.Colors
	GitInfo *gitinfo.GitInfo
	Effort  string // from settings.json
}

// Section is the interface that all status-line sections implement.
type Section interface {
	Name() string     // config key: "model", "context", etc.
	Priority() int    // lower = more important (1=highest, 9=lowest)
	Render(ctx *Context) string
	Width(ctx *Context) int
}

// LineSection associates a Section with the line number it belongs to.
type LineSection struct {
	Section Section
	Line    int // 1, 2, or 3
}

// AllSections returns all sections in their canonical line order.
func AllSections() []LineSection {
	return []LineSection{
		// Line 1
		{&ModelSection{}, 1},
		{&ElapsedSection{}, 1},
		{&GitSection{}, 1},
		// Line 2
		{&ContextSection{}, 2},
		{&RateLimitSection{}, 2},
		{&CostSection{}, 2},
		{&LinesSection{}, 2},
		{&APIDurationSection{}, 2},
		{&TokensSection{}, 2},
		// Line 3
		{&AgentSection{}, 3},
		{&VimSection{}, 3},
	}
}
