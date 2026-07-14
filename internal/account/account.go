// Package account는 Claude Code의 로그인 계정 정보를 읽는다.
//
// 이메일·플랜 정보는 상태 라인 stdin JSON에는 담기지 않으므로,
// Claude Code 내부 상태 파일(~/.claude.json)의 oauthAccount에서 읽는다.
// 이 파일은 공식 스키마가 아니라 내부 구현 세부이므로, 형식이 바뀌거나
// 없을 수 있음을 전제로 방어적으로 파싱한다 — 실패 시 조용히 nil을 반환한다.
package account

import (
	"encoding/json"
	"os"
)

// Account는 표시에 필요한 최소한의 계정 정보다.
type Account struct {
	Email    string // oauthAccount.emailAddress
	PlanType string // oauthAccount.organizationType (예: "claude_max")
}

// Load는 주어진 ~/.claude.json 경로에서 계정 정보를 읽는다.
// 파일이 없거나, 파싱에 실패하거나, 이메일이 비어 있으면 nil을 반환한다.
func Load(claudeJSONPath string) *Account {
	data, err := os.ReadFile(claudeJSONPath)
	if err != nil {
		return nil
	}

	var parsed struct {
		OAuthAccount struct {
			EmailAddress     string `json:"emailAddress"`
			OrganizationType string `json:"organizationType"`
		} `json:"oauthAccount"`
	}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil
	}

	if parsed.OAuthAccount.EmailAddress == "" {
		return nil
	}

	return &Account{
		Email:    parsed.OAuthAccount.EmailAddress,
		PlanType: parsed.OAuthAccount.OrganizationType,
	}
}
