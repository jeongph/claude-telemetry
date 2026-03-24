package section

import "github.com/jeongph/claude-telemetry/internal/render"

// AgentSection displays the active agent name (Line 3, Priority 7).
type AgentSection struct{}

func (s *AgentSection) Name() string  { return "agent" }
func (s *AgentSection) Priority() int { return 7 }

func (s *AgentSection) Render(ctx *Context) string {
	if ctx.Input.Agent == nil {
		return ""
	}
	return ctx.Colors.Dim("▶") + " " + ctx.Colors.Magenta(ctx.Input.Agent.Name)
}

func (s *AgentSection) Width(ctx *Context) int {
	return render.DisplayWidth(s.Render(ctx))
}
