// Package selfuninstall은 플러그인 제거 시 statusline 구성을 안전하게 정리한다.
// settings.json 조작을 bash가 아닌 테스트 가능한 Go 코드로 수행하기 위한 패키지다.
package selfuninstall

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// runShStub은 정리 후 run.sh에 남기는 무해화 스텁이다.
// 현재 떠 있는 세션이 statusline 명령을 계속 호출하므로 빈 출력으로 응답해야 한다.
// 다음 세션부터는 settings.json에 등록이 없어 호출되지 않는다.
const runShStub = "#!/bin/bash\n# claude-telemetry removed — this stub is safe to delete\nexit 0\n"

// RemoveStatusLine은 settings.json에서 statusLine 키만 제거한다.
// 다른 키는 보존하며, 원본은 settings.json.claude-telemetry.bak으로 백업한다.
// statusLine 키가 없거나 파일이 없으면 아무것도 하지 않는다 (멱등).
// 깨진 JSON이면 파일을 건드리지 않고 에러를 반환한다.
func RemoveStatusLine(settingsPath string) error {
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return err
	}
	var settings map[string]json.RawMessage
	if err := json.Unmarshal(data, &settings); err != nil {
		return fmt.Errorf("settings.json 파싱 실패 (수정하지 않음): %w", err)
	}
	if _, ok := settings["statusLine"]; !ok {
		return nil
	}
	delete(settings, "statusLine")

	out, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	out = append(out, '\n')

	mode := fs.FileMode(0o600)
	if fi, err := os.Stat(settingsPath); err == nil {
		mode = fi.Mode()
	}
	if err := os.WriteFile(settingsPath+".claude-telemetry.bak", data, mode); err != nil {
		return err
	}
	tmp := settingsPath + ".claude-telemetry.tmp"
	if err := os.WriteFile(tmp, out, mode); err != nil {
		return err
	}
	return os.Rename(tmp, settingsPath)
}

// CleanupFiles는 statusline 디렉토리의 구성 요소를 정리하고 run.sh를 스텁으로 교체한다.
// run.sh 자체는 삭제하지 않는다 — 현재 세션이 계속 호출하기 때문.
func CleanupFiles(statuslineDir string) error {
	for _, f := range []string{"config.json", ".managed-by-plugin", ".removal-detected"} {
		if err := os.Remove(filepath.Join(statuslineDir, f)); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return err
		}
	}
	for _, d := range []string{"bin", "cache"} {
		if err := os.RemoveAll(filepath.Join(statuslineDir, d)); err != nil {
			return err
		}
	}
	runSh := filepath.Join(statuslineDir, "run.sh")
	if _, err := os.Stat(runSh); err == nil {
		if err := os.WriteFile(runSh, []byte(runShStub), 0o755); err != nil {
			return err
		}
	}
	return nil
}

// Run은 self-uninstall 전체 절차를 수행한다.
// 락 파일(O_EXCL)로 동시 실행을 방지하며, 락 선점 실패는 에러가 아니다 (이미 진행 중).
func Run(claudeDir, statuslineDir string) error {
	lock := filepath.Join(statuslineDir, ".uninstall.lock")
	f, err := os.OpenFile(lock, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return nil
	}
	f.Close()

	if err := RemoveStatusLine(filepath.Join(claudeDir, "settings.json")); err != nil {
		os.Remove(lock)
		return err
	}
	if err := CleanupFiles(statuslineDir); err != nil {
		os.Remove(lock)
		return err
	}
	return os.Remove(lock)
}
