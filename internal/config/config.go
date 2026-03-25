package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

type Thresholds struct {
	ContextWarn   float64 `json:"context_warn"`
	ContextDanger float64 `json:"context_danger"`
	CostWarn      float64 `json:"cost_warn"`
	CostDanger    float64 `json:"cost_danger"`
}

type Config struct {
	Preset     string          `json:"preset"`
	Language   string          `json:"language"`
	ColorsOn   *bool           `json:"colors"`
	BarWidth   int             `json:"bar_width"`
	Separator  string          `json:"separator"`
	UserType   string          `json:"user_type"`
	Sections   map[string]bool `json:"sections"`
	Thresholds Thresholds      `json:"thresholds"`
}

// v1 → v2 section name aliases
var sectionAliases = map[string]string{
	"rate_limits":  "ratelimit",
	"duration":     "elapsed",
	"vim_mode":     "vim",
	"api_duration": "apiduration",
}

func defaultConfig() Config {
	return Config{
		Preset:    "normal",
		Language:  "auto",
		ColorsOn:  boolPtr(true),
		BarWidth:  5,
		Separator: " │ ",
		UserType:  "auto",
		Sections:  map[string]bool{},
		Thresholds: Thresholds{
			ContextWarn:   50,
			ContextDanger: 20,
			CostWarn:      1.0,
			CostDanger:    5.0,
		},
	}
}

func boolPtr(b bool) *bool { return &b }

func Load(configDir, projectDir string) Config {
	cfg := defaultConfig()

	// global config
	if configDir != "" {
		loadFile(filepath.Join(configDir, "config.json"), &cfg)
	}

	// project config (shallow merge)
	if projectDir != "" {
		var proj Config
		if loadFile(filepath.Join(projectDir, ".claude-statusline.json"), &proj) {
			if proj.Preset != "" {
				cfg.Preset = proj.Preset
			}
			if proj.Language != "" {
				cfg.Language = proj.Language
			}
			for k, v := range proj.Sections {
				if cfg.Sections == nil {
					cfg.Sections = map[string]bool{}
				}
				cfg.Sections[k] = v
			}
		}
	}

	// clamp bar_width
	if cfg.BarWidth < 3 {
		cfg.BarWidth = 3
	}
	if cfg.BarWidth > 10 {
		cfg.BarWidth = 10
	}

	return cfg
}

func loadFile(path string, cfg *Config) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	json.Unmarshal(data, cfg) //nolint:errcheck
	return true
}

// Colors returns whether colors are enabled.
func (c Config) Colors() bool {
	if c.ColorsOn == nil {
		return true
	}
	return *c.ColorsOn
}

// IsSectionEnabled checks if a section is enabled using the 3-layer priority:
// 1. sections override (direct match)
// 2. sections override (v1 alias reverse lookup)
// 3. preset default
func (c Config) IsSectionEnabled(name string) bool {
	// direct match in sections override
	if v, ok := c.Sections[name]; ok {
		return v
	}
	// check v1 aliases (reverse lookup: if name is v2 name, check if v1 name exists in sections)
	for old, newName := range sectionAliases {
		if newName == name {
			if v, ok := c.Sections[old]; ok {
				return v
			}
		}
	}
	// preset default
	p := GetPreset(c.Preset)
	if v, ok := p.Sections[name]; ok {
		return v
	}
	return true
}

// ResolveLanguage resolves "auto" to an actual language code by reading ~/.claude/settings.json.
func ResolveLanguage(cfgLang, claudeDir string) string {
	if cfgLang != "auto" && cfgLang != "" {
		return cfgLang
	}
	data, _ := os.ReadFile(filepath.Join(claudeDir, "settings.json"))
	var s struct {
		Language string `json:"language"`
	}
	json.Unmarshal(data, &s) //nolint:errcheck
	switch {
	case strings.HasPrefix(s.Language, "ko"), s.Language == "한국어":
		return "ko"
	case strings.HasPrefix(s.Language, "ja"), s.Language == "日本語":
		return "ja"
	case strings.HasPrefix(s.Language, "zh"), s.Language == "中文":
		return "zh"
	default:
		return "en"
	}
}

// ReadEffortLevel reads the effortLevel from ~/.claude/settings.json.
func ReadEffortLevel(claudeDir string) string {
	data, err := os.ReadFile(filepath.Join(claudeDir, "settings.json"))
	if err != nil {
		return "auto"
	}
	var s struct {
		EffortLevel string `json:"effortLevel"`
	}
	json.Unmarshal(data, &s) //nolint:errcheck
	if s.EffortLevel == "" {
		return "auto"
	}
	return s.EffortLevel
}
