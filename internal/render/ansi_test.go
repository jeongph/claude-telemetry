package render

import (
	"os"
	"testing"
)

func TestColorFunctions(t *testing.T) {
	c := NewColors(true)
	got := c.Cyan("test")
	if got != "\033[1;36mtest\033[0m" {
		t.Errorf("Cyan = %q", got)
	}
}

func TestNoColor(t *testing.T) {
	c := NewColors(false)
	if c.Cyan("test") != "test" {
		t.Errorf("NoColor Cyan should return plain text")
	}
}

func TestNoColorEnv(t *testing.T) {
	os.Setenv("NO_COLOR", "1")
	defer os.Unsetenv("NO_COLOR")
	c := NewColors(true) // enabled=true but NO_COLOR overrides
	if c.Cyan("test") != "test" {
		t.Errorf("NO_COLOR env should disable colors")
	}
}

func TestDisplayWidth(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"hello", 5},
		{"컨텍스트", 8},              // 4 Korean chars x 2
		{"a한b", 4},               // 1 + 2 + 1
		{"\033[1;36mtest\033[0m", 4}, // ANSI stripped
		{"", 0},
	}
	for _, tt := range tests {
		if got := DisplayWidth(tt.input); got != tt.want {
			t.Errorf("DisplayWidth(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestStripANSI(t *testing.T) {
	if got := StripANSI("\033[1;36mhello\033[0m"); got != "hello" {
		t.Errorf("StripANSI = %q, want hello", got)
	}
}

func TestThresholdColorRemaining(t *testing.T) {
	c := NewColors(true)
	tests := []struct {
		pct  float64
		want string
	}{
		{72, "green"},  // > 50
		{30, "yellow"}, // 21-50
		{15, "red"},    // <= 20
		{50, "yellow"}, // exactly 50
		{20, "red"},    // exactly 20
	}
	for _, tt := range tests {
		if got := ThresholdColorRemaining(tt.pct, c, 50, 20); got != tt.want {
			t.Errorf("ThresholdColor(%v) = %q, want %q", tt.pct, got, tt.want)
		}
	}
}

func TestProgressBarWidth(t *testing.T) {
	c := NewColors(false) // no color for easy width check
	bar := ProgressBarRemaining(72, 5, c, 50, 20)
	if DisplayWidth(bar) != 5 {
		t.Errorf("bar width = %d, want 5", DisplayWidth(bar))
	}
}

func TestProgressBarZeroAndFull(t *testing.T) {
	c := NewColors(false)
	bar0 := ProgressBarRemaining(0, 5, c, 50, 20)
	bar100 := ProgressBarRemaining(100, 5, c, 50, 20)
	if DisplayWidth(bar0) != 5 {
		t.Errorf("0%% bar width = %d", DisplayWidth(bar0))
	}
	if DisplayWidth(bar100) != 5 {
		t.Errorf("100%% bar width = %d", DisplayWidth(bar100))
	}
}
