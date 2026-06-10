package config

type PresetDef struct {
	Sections map[string]bool
}

var presets = map[string]PresetDef{
	"compact": {Sections: map[string]bool{
		"session": false,
		"model":   true, "effort": true, "context": true, "ratelimit": true, "cost": true,
		"elapsed": false, "git": false, "pr": false, "lines": false,
		"tokens": false, "apiduration": false, "agent": false, "vim": false,
		"thinking": false,
	}},
	"normal": {Sections: map[string]bool{
		"session": true,
		"model":   true, "effort": true, "elapsed": true, "git": true, "pr": true,
		"context": true, "ratelimit": true, "cost": true,
		"lines": false, "tokens": false, "apiduration": false,
		"agent": true, "vim": true,
		"thinking": false,
	}},
	"detailed": {Sections: map[string]bool{
		"session": true,
		"model":   true, "effort": true, "elapsed": true, "git": true, "pr": true,
		"context": true, "ratelimit": true, "cost": true,
		"lines": true, "tokens": true, "apiduration": true,
		"agent": true, "vim": true,
		"thinking": true,
	}},
}

func GetPreset(name string) PresetDef {
	if p, ok := presets[name]; ok {
		return p
	}
	return presets["normal"]
}
