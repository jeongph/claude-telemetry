package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDefaultConfig(t *testing.T) {
	cfg := Load("", "")
	if cfg.Preset != "normal" {
		t.Errorf("default preset: got %q, want %q", cfg.Preset, "normal")
	}
	if cfg.BarWidth != 5 {
		t.Errorf("default bar_width: got %d, want 5", cfg.BarWidth)
	}
}

func TestLoadConfigFile(t *testing.T) {
	dir := t.TempDir()
	data := map[string]interface{}{
		"preset":    "compact",
		"bar_width": 7,
		"language":  "ko",
	}
	writeJSON(t, filepath.Join(dir, "config.json"), data)

	cfg := Load(dir, "")
	if cfg.Preset != "compact" {
		t.Errorf("loaded preset: got %q, want %q", cfg.Preset, "compact")
	}
	if cfg.BarWidth != 7 {
		t.Errorf("loaded bar_width: got %d, want 7", cfg.BarWidth)
	}
	if cfg.Language != "ko" {
		t.Errorf("loaded language: got %q, want %q", cfg.Language, "ko")
	}
}

func TestBarWidthClamp(t *testing.T) {
	dir := t.TempDir()

	writeJSON(t, filepath.Join(dir, "config.json"), map[string]interface{}{"bar_width": 100})
	cfg := Load(dir, "")
	if cfg.BarWidth != 10 {
		t.Errorf("clamp upper: got %d, want 10", cfg.BarWidth)
	}

	writeJSON(t, filepath.Join(dir, "config.json"), map[string]interface{}{"bar_width": 1})
	cfg = Load(dir, "")
	if cfg.BarWidth != 3 {
		t.Errorf("clamp lower: got %d, want 3", cfg.BarWidth)
	}
}

func TestProjectConfigMerge(t *testing.T) {
	globalDir := t.TempDir()
	projectDir := t.TempDir()

	writeJSON(t, filepath.Join(globalDir, "config.json"), map[string]interface{}{
		"preset": "compact",
	})
	writeJSON(t, filepath.Join(projectDir, ".claude-statusline.json"), map[string]interface{}{
		"preset": "detailed",
	})

	cfg := Load(globalDir, projectDir)
	if cfg.Preset != "detailed" {
		t.Errorf("project config should override global: got %q, want %q", cfg.Preset, "detailed")
	}
}

func TestSectionEnabled(t *testing.T) {
	// normal preset: model=true, tokens=false
	cfg := defaultConfig()
	cfg.Preset = "normal"

	if !cfg.IsSectionEnabled("model") {
		t.Error("normal preset: model should be enabled")
	}
	if cfg.IsSectionEnabled("tokens") {
		t.Error("normal preset: tokens should be disabled")
	}

	// sections override should win
	cfg.Sections["tokens"] = true
	if !cfg.IsSectionEnabled("tokens") {
		t.Error("sections override: tokens should be enabled after override")
	}
}

func TestV1ConfigCompat(t *testing.T) {
	cfg := defaultConfig()
	cfg.Preset = "normal"
	cfg.Sections = map[string]bool{
		"rate_limits":  true,
		"duration":     true,
		"vim_mode":     false,
		"api_duration": true,
	}

	// v1 "rate_limits" → v2 "ratelimit"
	if !cfg.IsSectionEnabled("ratelimit") {
		t.Error("v1 alias: rate_limits → ratelimit should be enabled")
	}
	// v1 "duration" → v2 "elapsed"
	if !cfg.IsSectionEnabled("elapsed") {
		t.Error("v1 alias: duration → elapsed should be enabled")
	}
	// v1 "vim_mode" → v2 "vim" (disabled override)
	if cfg.IsSectionEnabled("vim") {
		t.Error("v1 alias: vim_mode → vim should be disabled")
	}
	// v1 "api_duration" → v2 "apiduration"
	if !cfg.IsSectionEnabled("apiduration") {
		t.Error("v1 alias: api_duration → apiduration should be enabled")
	}
}

func TestResolveLanguage(t *testing.T) {
	dir := t.TempDir()

	// auto with 한국어 in settings.json
	writeJSON(t, filepath.Join(dir, "settings.json"), map[string]interface{}{
		"language": "한국어",
	})
	lang := ResolveLanguage("auto", dir)
	if lang != "ko" {
		t.Errorf("ResolveLanguage(auto, 한국어): got %q, want %q", lang, "ko")
	}

	// explicit language overrides settings.json
	lang = ResolveLanguage("ja", dir)
	if lang != "ja" {
		t.Errorf("ResolveLanguage(ja, ...): got %q, want %q", lang, "ja")
	}

	// auto with missing settings.json → en
	emptyDir := t.TempDir()
	lang = ResolveLanguage("auto", emptyDir)
	if lang != "en" {
		t.Errorf("ResolveLanguage(auto, empty): got %q, want %q", lang, "en")
	}
}

func TestReadEffortLevel(t *testing.T) {
	dir := t.TempDir()

	writeJSON(t, filepath.Join(dir, "settings.json"), map[string]interface{}{
		"effortLevel": "high",
	})
	level := ReadEffortLevel(dir)
	if level != "high" {
		t.Errorf("ReadEffortLevel: got %q, want %q", level, "high")
	}

	// missing file → "auto"
	emptyDir := t.TempDir()
	level = ReadEffortLevel(emptyDir)
	if level != "auto" {
		t.Errorf("ReadEffortLevel(missing): got %q, want %q", level, "auto")
	}
}

// helper
func writeJSON(t *testing.T, path string, v interface{}) {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
}
