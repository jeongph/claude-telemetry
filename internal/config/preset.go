package config

type PresetDef struct {
	Sections map[string]bool
}

var presets = map[string]PresetDef{
	"compact": {Sections: map[string]bool{
		"model": true, "context": true, "ratelimit": true, "cost": true,
		"elapsed": false, "git": false, "lines": false,
		"tokens": false, "apiduration": false, "agent": false, "vim": false,
	}},
	"normal": {Sections: map[string]bool{
		"model": true, "elapsed": true, "git": true,
		"context": true, "ratelimit": true, "cost": true,
		"lines": false, "tokens": false, "apiduration": false,
		"agent": true, "vim": true,
	}},
	"detailed": {Sections: map[string]bool{
		"model": true, "elapsed": true, "git": true,
		"context": true, "ratelimit": true, "cost": true,
		"lines": true, "tokens": true, "apiduration": true,
		"agent": true, "vim": true,
	}},
}

func GetPreset(name string) PresetDef {
	if p, ok := presets[name]; ok {
		return p
	}
	return presets["normal"]
}
