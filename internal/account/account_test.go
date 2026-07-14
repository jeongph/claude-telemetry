package account

import (
	"os"
	"path/filepath"
	"testing"
)

func writeClaudeJSON(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, ".claude.json")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("임시 파일 쓰기 실패: %v", err)
	}
	return path
}

func TestLoadReadsEmailAndPlan(t *testing.T) {
	path := writeClaudeJSON(t, `{
		"oauthAccount": {
			"emailAddress": "user@example.com",
			"organizationType": "claude_max"
		}
	}`)
	acc := Load(path)
	if acc == nil {
		t.Fatal("Load: 계정 정보가 있는데 nil 반환")
	}
	if acc.Email != "user@example.com" {
		t.Errorf("Email: got %q, want %q", acc.Email, "user@example.com")
	}
	if acc.PlanType != "claude_max" {
		t.Errorf("PlanType: got %q, want %q", acc.PlanType, "claude_max")
	}
}

func TestLoadMissingFileReturnsNil(t *testing.T) {
	acc := Load(filepath.Join(t.TempDir(), "does-not-exist.json"))
	if acc != nil {
		t.Errorf("Load: 파일 없음인데 non-nil 반환 — %+v", acc)
	}
}

func TestLoadNoOAuthAccountReturnsNil(t *testing.T) {
	path := writeClaudeJSON(t, `{"userID": "abc"}`)
	if acc := Load(path); acc != nil {
		t.Errorf("Load: oauthAccount 없음인데 non-nil 반환 — %+v", acc)
	}
}

func TestLoadEmptyEmailReturnsNil(t *testing.T) {
	path := writeClaudeJSON(t, `{"oauthAccount": {"emailAddress": "", "organizationType": "claude_max"}}`)
	if acc := Load(path); acc != nil {
		t.Errorf("Load: 이메일 빈값인데 non-nil 반환 — %+v", acc)
	}
}

func TestLoadMalformedJSONReturnsNil(t *testing.T) {
	path := writeClaudeJSON(t, `{ not valid json `)
	if acc := Load(path); acc != nil {
		t.Errorf("Load: 깨진 JSON인데 non-nil 반환 — %+v", acc)
	}
}
