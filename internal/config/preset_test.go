package config

import "testing"

func TestPresetNormal(t *testing.T) {
	p := GetPreset("normal")
	if !p.Sections["context"] {
		t.Error("normal preset: context should be enabled")
	}
	if p.Sections["tokens"] {
		t.Error("normal preset: tokens should be disabled")
	}
}

func TestPresetCompact(t *testing.T) {
	p := GetPreset("compact")
	if !p.Sections["context"] {
		t.Error("compact preset: context should be enabled")
	}
	if p.Sections["git"] {
		t.Error("compact preset: git should be disabled")
	}
}

func TestPresetDetailed(t *testing.T) {
	p := GetPreset("detailed")
	if !p.Sections["tokens"] {
		t.Error("detailed preset: tokens should be enabled")
	}
}

func TestPresetUnknownFallback(t *testing.T) {
	p := GetPreset("nonexistent")
	// should fall back to normal
	normal := GetPreset("normal")
	for k, v := range normal.Sections {
		if p.Sections[k] != v {
			t.Errorf("unknown preset fallback: section %q mismatch: got %v, want %v", k, p.Sections[k], v)
		}
	}
}
