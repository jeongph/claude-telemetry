package render

import (
	"strings"
	"testing"
)

func TestAssembleLinesNormal(t *testing.T) {
	lines := AssembleLines(
		[]string{"Opus", "12m 34s", "myproject:main"},
		[]ScoredSegment{
			{Text: "Context 72%", Width: 11, Priority: 1, Order: 0},
			{Text: "5h 88%", Width: 6, Priority: 2, Order: 1},
			{Text: "tokens 15k", Width: 10, Priority: 9, Order: 2},
		},
		[]string{"agent-name", "NORMAL"},
		" │ ", 80, false,
	)
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}
	if !strings.Contains(lines[0], "Opus") {
		t.Error("line1 should contain model")
	}
	if !strings.Contains(lines[1], "Context") {
		t.Error("line2 should contain context")
	}
	if !strings.Contains(lines[2], "agent-name") {
		t.Error("line3 should contain agent")
	}
}

func TestAdaptiveWidthDropsLowPriority(t *testing.T) {
	segments := []ScoredSegment{
		{Text: "Context 72%", Width: 11, Priority: 1, Order: 0},
		{Text: "5h 88%", Width: 6, Priority: 2, Order: 1},
		{Text: "long tokens text here", Width: 21, Priority: 9, Order: 2},
	}
	// Width 25: Context(11) + sep(3) + 5h(6) = 20, fits. tokens(21) won't fit.
	result := fitSegments(segments, " │ ", 25)
	if len(result) != 2 {
		t.Errorf("expected 2 segments, got %d", len(result))
	}
	// Verify order preserved
	if result[0].Priority != 1 || result[1].Priority != 2 {
		t.Error("should preserve original order")
	}
}

func TestAdaptiveWidthAllFit(t *testing.T) {
	segments := []ScoredSegment{
		{Text: "A", Width: 1, Priority: 1, Order: 0},
		{Text: "B", Width: 1, Priority: 2, Order: 1},
	}
	result := fitSegments(segments, " │ ", 80)
	if len(result) != 2 {
		t.Errorf("all should fit, got %d", len(result))
	}
}

func TestEmptyLine3(t *testing.T) {
	lines := AssembleLines(
		[]string{"Opus"},
		[]ScoredSegment{{Text: "Context", Width: 7, Priority: 1, Order: 0}},
		[]string{},
		" │ ", 80, false,
	)
	if len(lines) != 2 {
		t.Errorf("empty line3 should not be output, got %d lines", len(lines))
	}
}

func TestCompactMode(t *testing.T) {
	lines := AssembleLines(
		[]string{"Opus"},
		[]ScoredSegment{
			{Text: "72%", Width: 3, Priority: 1, Order: 0},
			{Text: "5h 88%", Width: 6, Priority: 2, Order: 1},
		},
		[]string{"NORMAL"},
		" │ ", 80, true,
	)
	if len(lines) != 1 {
		t.Fatalf("compact should be 1 line, got %d", len(lines))
	}
	if !strings.Contains(lines[0], "Opus") || !strings.Contains(lines[0], "72%") {
		t.Error("compact line should contain both model and context")
	}
}

func TestEmptySegmentsSkipped(t *testing.T) {
	segments := []ScoredSegment{
		{Text: "A", Width: 1, Priority: 1, Order: 0},
		{Text: "", Width: 0, Priority: 2, Order: 1}, // empty
		{Text: "C", Width: 1, Priority: 3, Order: 2},
	}
	result := fitSegments(segments, " │ ", 80)
	if len(result) != 2 {
		t.Errorf("empty segments should be skipped, got %d", len(result))
	}
}
