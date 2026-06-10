package section

import (
	"strings"

	"github.com/jeongph/claude-telemetry/internal/render"
)

// sessionNameMaxWidth는 세션명 표시 최대 컬럼 수다.
// CC 2.1.170부터 자동 생성된 긴 세션 제목이 올 수 있어 절단이 필요하다 (실측 확인).
const sessionNameMaxWidth = 20

// SessionSection displays the session name (Line 1, Priority 6).
// /rename, --name 지정 또는 자동 생성된 세션 제목을 표시한다.
type SessionSection struct{}

func (s *SessionSection) Name() string  { return "session" }
func (s *SessionSection) Priority() int { return 6 }

func (s *SessionSection) Render(ctx *Context) string {
	name := ctx.Input.SessionName
	if name == "" {
		return ""
	}
	name = truncateDisplay(name, sessionNameMaxWidth)
	return ctx.Colors.Dim("[") + ctx.Colors.Cyan(name) + ctx.Colors.Dim("]")
}

// truncateDisplay는 표시 폭 기준으로 문자열을 자르고 말줄임표를 붙인다.
// 한글 등 와이드 문자는 2컬럼으로 계산된다.
func truncateDisplay(s string, max int) string {
	if render.DisplayWidth(s) <= max {
		return s
	}
	var b strings.Builder
	w := 0
	for _, r := range s {
		rw := render.DisplayWidth(string(r))
		if w+rw > max-1 { // 말줄임표 1컬럼 예약
			break
		}
		b.WriteRune(r)
		w += rw
	}
	return b.String() + "…"
}

func (s *SessionSection) Width(ctx *Context) int {
	return render.DisplayWidth(s.Render(ctx))
}
