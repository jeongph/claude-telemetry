package section

import "github.com/jeongph/claude-telemetry/internal/render"

// renderEffort는 reasoning effort 레벨을 레벨별 색상으로 렌더링한다.
// Claude Code 2.1.141+ 의 effort.level 입력을 사용한다 (라이브 세션 값).
// ModelSection이 모델명 옆에 가운뎃점으로 이어 표시하며, "effort" 섹션 키가
// 꺼져 있거나 입력에 effort가 없으면(미지원 모델) 빈 문자열을 반환한다.
func renderEffort(ctx *Context) string {
	if !ctx.Config.IsSectionEnabled("effort") {
		return ""
	}
	eff := ctx.Input.Effort
	if eff == nil || eff.Level == "" {
		return ""
	}
	return effortColor(ctx.Colors, eff.Level)(eff.Level)
}

// effortColor는 effort 레벨별 색상 함수를 반환한다 (높을수록 강조).
func effortColor(c render.Colors, level string) func(string) string {
	switch level {
	case "low":
		return c.Dim
	case "medium":
		return c.White
	case "high":
		return c.Green
	case "xhigh":
		return c.Yellow
	case "max":
		return c.Red
	default:
		return c.White
	}
}
