package render

import (
	"sort"
	"strings"
)

type ScoredSegment struct {
	Text     string
	Width    int
	Priority int // lower = more important
	Order    int // original order for restoration
}

// AssembleLines builds the final output lines.
// compact=true merges everything into 1 line.
// line4Parts는 전용 줄(예: user 섹션)로, 비어 있으면 줄을 만들지 않는다.
func AssembleLines(line1Parts []string, line2Segments []ScoredSegment, line3Parts, line4Parts []string, sep string, maxWidth int, compact bool) []string {
	if compact {
		// Compact mode: merge line2 segments into line1
		all := make([]string, 0)
		all = append(all, filterNonEmpty(line1Parts)...)
		for _, seg := range line2Segments {
			if seg.Text != "" {
				all = append(all, seg.Text)
			}
		}
		l := strings.Join(all, sep)
		if l != "" {
			return []string{l}
		}
		return nil
	}

	var lines []string

	// Line 1: fixed layout, join non-empty parts
	l1 := joinNonEmpty(line1Parts, sep)
	if l1 != "" {
		lines = append(lines, l1)
	}

	// Line 2: adaptive width — drop lowest priority segments first
	fitted := fitSegments(line2Segments, sep, maxWidth)
	if len(fitted) > 0 {
		texts := make([]string, len(fitted))
		for i, s := range fitted {
			texts[i] = s.Text
		}
		lines = append(lines, strings.Join(texts, sep))
	}

	// Line 3: only when non-empty
	l3 := joinNonEmpty(line3Parts, sep)
	if l3 != "" {
		lines = append(lines, l3)
	}

	// Line 4: 전용 줄, non-empty일 때만
	l4 := joinNonEmpty(line4Parts, sep)
	if l4 != "" {
		lines = append(lines, l4)
	}

	return lines
}

// fitSegments selects segments that fit within maxWidth,
// preferring higher priority (lower number).
func fitSegments(segments []ScoredSegment, sep string, maxWidth int) []ScoredSegment {
	if len(segments) == 0 {
		return nil
	}

	sepWidth := DisplayWidth(sep)

	// Sort by priority (lower = more important)
	sorted := make([]ScoredSegment, len(segments))
	copy(sorted, segments)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Priority < sorted[j].Priority
	})

	// Greedily add segments by priority until width exceeded
	var selected []ScoredSegment
	totalWidth := 0
	for _, seg := range sorted {
		if seg.Text == "" {
			continue
		}
		extra := 0
		if len(selected) > 0 {
			extra = sepWidth
		}
		if totalWidth+seg.Width+extra <= maxWidth-2 { // 2 char margin
			selected = append(selected, seg)
			totalWidth += seg.Width + extra
		}
	}

	// Restore original order
	sort.Slice(selected, func(i, j int) bool {
		return selected[i].Order < selected[j].Order
	})

	return selected
}

func joinNonEmpty(parts []string, sep string) string {
	return strings.Join(filterNonEmpty(parts), sep)
}

func filterNonEmpty(parts []string) []string {
	var result []string
	for _, p := range parts {
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
