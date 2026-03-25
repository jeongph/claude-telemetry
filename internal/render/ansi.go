package render

import (
	"math"
	"os"
	"regexp"
	"unicode"
)

var ansiRegex = regexp.MustCompile(`\033\[[^m]*m`)

// Colors holds ANSI color state.
type Colors struct {
	enabled bool
}

// NewColors creates a Colors instance. If the NO_COLOR env var is set,
// colors are disabled regardless of the enabled param.
func NewColors(enabled bool) Colors {
	if os.Getenv("NO_COLOR") != "" {
		enabled = false
	}
	return Colors{enabled: enabled}
}

// Enabled reports whether colors are enabled.
func (c Colors) Enabled() bool {
	return c.enabled
}

// wrap wraps s with ANSI escape code.
func (c Colors) wrap(code, s string) string {
	return "\033[" + code + "m" + s + "\033[0m"
}

// Cyan returns s in cyan (1;36).
func (c Colors) Cyan(s string) string {
	if !c.enabled {
		return s
	}
	return c.wrap("1;36", s)
}

// Green returns s in green (32).
func (c Colors) Green(s string) string {
	if !c.enabled {
		return s
	}
	return c.wrap("32", s)
}

// Yellow returns s in yellow (33).
func (c Colors) Yellow(s string) string {
	if !c.enabled {
		return s
	}
	return c.wrap("33", s)
}

// Red returns s in red (31).
func (c Colors) Red(s string) string {
	if !c.enabled {
		return s
	}
	return c.wrap("31", s)
}

// Magenta returns s in magenta (35).
func (c Colors) Magenta(s string) string {
	if !c.enabled {
		return s
	}
	return c.wrap("35", s)
}

// White returns s in white (37).
func (c Colors) White(s string) string {
	if !c.enabled {
		return s
	}
	return c.wrap("37", s)
}

// Dim returns s in dim white (2;37).
func (c Colors) Dim(s string) string {
	if !c.enabled {
		return s
	}
	return c.wrap("2;37", s)
}

// Reset returns the ANSI reset sequence.
func (c Colors) Reset() string {
	if !c.enabled {
		return ""
	}
	return "\033[0m"
}

// StripANSI removes all ANSI escape sequences from s.
func StripANSI(s string) string {
	return ansiRegex.ReplaceAllString(s, "")
}

// DisplayWidth returns the terminal display width of s, accounting for
// CJK wide characters (width 2) and stripping ANSI codes first.
func DisplayWidth(s string) int {
	s = StripANSI(s)
	width := 0
	for _, r := range s {
		if isWide(r) {
			width += 2
		} else {
			width += 1
		}
	}
	return width
}

// isWide reports whether r is a wide (CJK) character.
func isWide(r rune) bool {
	return unicode.Is(unicode.Han, r) ||
		(r >= 0xAC00 && r <= 0xD7A3) || // Hangul syllables
		(r >= 0x3000 && r <= 0x303F) || // CJK Symbols and Punctuation
		unicode.Is(unicode.Katakana, r) ||
		unicode.Is(unicode.Hiragana, r)
}

// ThresholdColorRemaining returns a color name based on the remaining
// percentage pct. warn and danger are the thresholds.
// pct <= danger → "red", pct <= warn → "yellow", otherwise → "green".
func ThresholdColorRemaining(pct float64, c Colors, warn, danger float64) string {
	if pct <= danger {
		return "red"
	}
	if pct <= warn {
		return "yellow"
	}
	return "green"
}

// ApplyColor applies the named color to s using c.
func ApplyColor(colorName string, s string, c Colors) string {
	switch colorName {
	case "red":
		return c.Red(s)
	case "yellow":
		return c.Yellow(s)
	case "green":
		return c.Green(s)
	default:
		return s
	}
}

// ProgressBarRemaining renders a horizontal progress bar based on remaining
// percentage. Filled blocks are colored by threshold; empty blocks are dimmed.
func ProgressBarRemaining(pct float64, width int, c Colors, warn, danger float64) string {
	filled := int(math.Round(pct / 100.0 * float64(width)))
	if filled < 0 {
		filled = 0
	}
	if filled > width {
		filled = width
	}
	empty := width - filled

	colorName := ThresholdColorRemaining(pct, c, warn, danger)

	filledBlock := "▰"
	emptyBlock := "▱"

	bar := ""
	for i := 0; i < filled; i++ {
		bar += ApplyColor(colorName, filledBlock, c)
	}
	for i := 0; i < empty; i++ {
		bar += c.Dim(emptyBlock)
	}
	return bar
}
