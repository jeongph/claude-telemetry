package section

import "github.com/jeongph/claude-telemetry/internal/render"

// ModelSection displays the model name with the reasoning effort level (Line 1, Priority 3).
// effort는 모델의 추론 파라미터이므로 별도 세그먼트가 아니라 모델명 옆에 붙여 표시한다
// (예: "Fable 5 ↯xhigh"). effort 표기는 "effort" 섹션 키로 켜고 끌 수 있다.
type ModelSection struct{}

func (s *ModelSection) Name() string  { return "model" }
func (s *ModelSection) Priority() int { return 3 }

func (s *ModelSection) Render(ctx *Context) string {
	name := ctx.Input.Model.DisplayName
	if name == "" {
		return ""
	}
	out := ctx.Colors.Cyan(name)
	if eff := renderEffort(ctx); eff != "" {
		out += " " + eff
	}
	return out
}

func (s *ModelSection) Width(ctx *Context) int {
	return render.DisplayWidth(s.Render(ctx))
}
