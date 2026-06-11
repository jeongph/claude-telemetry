package selfuninstall

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("파일 쓰기 실패: %v", err)
	}
}

func TestRemoveStatusLine(t *testing.T) {
	dir := t.TempDir()
	settings := filepath.Join(dir, "settings.json")
	writeFile(t, settings, `{
  "model": "opus",
  "statusLine": {"type": "command", "command": "bash /x/run.sh"},
  "permissions": {"allow": ["Bash(ls:*)"]}
}`)

	if err := RemoveStatusLine(settings); err != nil {
		t.Fatalf("RemoveStatusLine 실패: %v", err)
	}

	data, _ := os.ReadFile(settings)
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("결과가 유효한 JSON이 아님: %v", err)
	}
	if _, ok := m["statusLine"]; ok {
		t.Error("statusLine 키가 제거되지 않음")
	}
	if _, ok := m["model"]; !ok {
		t.Error("model 키가 보존되지 않음")
	}
	if _, ok := m["permissions"]; !ok {
		t.Error("permissions 키가 보존되지 않음")
	}

	// 백업 존재 + 원본 내용 보존
	bak, err := os.ReadFile(settings + ".claude-telemetry.bak")
	if err != nil {
		t.Fatalf("백업 파일 없음: %v", err)
	}
	if !strings.Contains(string(bak), "statusLine") {
		t.Error("백업에 원본 statusLine이 없음")
	}
}

func TestRemoveStatusLineIdempotent(t *testing.T) {
	dir := t.TempDir()
	settings := filepath.Join(dir, "settings.json")
	original := `{"model": "opus"}`
	writeFile(t, settings, original)

	if err := RemoveStatusLine(settings); err != nil {
		t.Fatalf("statusLine 없는 파일에서 에러: %v", err)
	}
	data, _ := os.ReadFile(settings)
	if string(data) != original {
		t.Error("statusLine 없는 파일이 변경됨 — no-op이어야 함")
	}
	if _, err := os.Stat(settings + ".claude-telemetry.bak"); err == nil {
		t.Error("no-op인데 백업이 생성됨")
	}
}

func TestRemoveStatusLineMissingFile(t *testing.T) {
	dir := t.TempDir()
	if err := RemoveStatusLine(filepath.Join(dir, "settings.json")); err != nil {
		t.Errorf("파일 없음은 에러가 아니어야 함: %v", err)
	}
}

func TestRemoveStatusLineInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	settings := filepath.Join(dir, "settings.json")
	broken := `{invalid json`
	writeFile(t, settings, broken)

	if err := RemoveStatusLine(settings); err == nil {
		t.Error("깨진 JSON에서 에러가 반환되어야 함")
	}
	data, _ := os.ReadFile(settings)
	if string(data) != broken {
		t.Error("깨진 JSON 파일을 건드리면 안 됨")
	}
}

func TestCleanupFiles(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "bin"), 0o755)
	os.MkdirAll(filepath.Join(dir, "cache"), 0o755)
	writeFile(t, filepath.Join(dir, "bin", "claude-telemetry"), "binary")
	writeFile(t, filepath.Join(dir, "config.json"), "{}")
	writeFile(t, filepath.Join(dir, ".managed-by-plugin"), "")
	writeFile(t, filepath.Join(dir, ".removal-detected"), "123")
	writeFile(t, filepath.Join(dir, "run.sh"), "#!/bin/bash\nexec something")

	if err := CleanupFiles(dir); err != nil {
		t.Fatalf("CleanupFiles 실패: %v", err)
	}

	for _, gone := range []string{"bin", "cache", "config.json", ".managed-by-plugin", ".removal-detected"} {
		if _, err := os.Stat(filepath.Join(dir, gone)); err == nil {
			t.Errorf("%s가 제거되지 않음", gone)
		}
	}
	// run.sh는 삭제가 아니라 무해화 스텁으로 교체 (현재 세션이 계속 호출하므로)
	stub, err := os.ReadFile(filepath.Join(dir, "run.sh"))
	if err != nil {
		t.Fatal("run.sh 스텁이 없음")
	}
	if !strings.Contains(string(stub), "exit 0") {
		t.Errorf("run.sh가 무해화 스텁이 아님: %q", string(stub))
	}
}

func TestRunWithLock(t *testing.T) {
	claudeDir := t.TempDir()
	slDir := filepath.Join(claudeDir, "statusline")
	os.MkdirAll(slDir, 0o755)
	settings := filepath.Join(claudeDir, "settings.json")
	writeFile(t, settings, `{"statusLine": {"type": "command"}}`)
	// 락 선점 → Run은 조용히 no-op
	writeFile(t, filepath.Join(slDir, ".uninstall.lock"), "")

	if err := Run(claudeDir, slDir); err != nil {
		t.Fatalf("락 존재 시 에러 없이 종료해야 함: %v", err)
	}
	data, _ := os.ReadFile(settings)
	if !strings.Contains(string(data), "statusLine") {
		t.Error("락 존재 시 settings를 건드리면 안 됨")
	}
}

func TestRunFull(t *testing.T) {
	claudeDir := t.TempDir()
	slDir := filepath.Join(claudeDir, "statusline")
	os.MkdirAll(filepath.Join(slDir, "bin"), 0o755)
	settings := filepath.Join(claudeDir, "settings.json")
	writeFile(t, settings, `{"model": "opus", "statusLine": {"type": "command"}}`)
	writeFile(t, filepath.Join(slDir, "run.sh"), "#!/bin/bash\nexec x")

	if err := Run(claudeDir, slDir); err != nil {
		t.Fatalf("Run 실패: %v", err)
	}
	data, _ := os.ReadFile(settings)
	if strings.Contains(string(data), "statusLine") {
		t.Error("statusLine이 제거되지 않음")
	}
	if _, err := os.Stat(filepath.Join(slDir, "bin")); err == nil {
		t.Error("bin 디렉토리가 제거되지 않음")
	}
}
