package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/jeongph/claude-telemetry/internal/config"
	"github.com/jeongph/claude-telemetry/internal/gitinfo"
	"github.com/jeongph/claude-telemetry/internal/i18n"
	"github.com/jeongph/claude-telemetry/internal/input"
	"github.com/jeongph/claude-telemetry/internal/render"
	"github.com/jeongph/claude-telemetry/internal/section"
)

var version = "dev"

func main() {
	var debug bool
	for _, arg := range os.Args[1:] {
		switch arg {
		case "--version":
			fmt.Println("claude-telemetry", version)
			return
		case "--debug":
			debug = true
		}
	}

	// 1. Read stdin
	data, err := io.ReadAll(os.Stdin)
	if err != nil || len(data) == 0 {
		fmt.Println("⚠ statusline: no input")
		return
	}

	// 2. Parse input
	inp, err := input.Parse(data)
	if err != nil {
		fmt.Println("⚠ statusline: invalid input")
		return
	}

	// 3. Load config
	configDir := os.Getenv("CLAUDE_STATUSLINE_CONFIG")
	if configDir == "" {
		home, _ := os.UserHomeDir()
		configDir = filepath.Join(home, ".claude", "statusline")
	}
	projectDir := inp.CWD
	if projectDir == "" && inp.Workspace != nil {
		projectDir = inp.Workspace.ProjectDir
	}
	cfg := config.Load(configDir, projectDir)

	// 4. Resolve language
	home, _ := os.UserHomeDir()
	claudeDir := filepath.Join(home, ".claude")
	lang := config.ResolveLanguage(cfg.Language, claudeDir)
	locale := i18n.New(lang)

	// 5. Setup colors
	colors := render.NewColors(cfg.Colors())

	// 6. Read effort level
	effort := config.ReadEffortLevel(claudeDir)

	// 7. Gather git info
	cwd := inp.CWD
	if cwd == "" {
		cwd = "."
	}
	gitInfo := gitinfo.Gather(cwd, "")

	// 8. Get terminal width
	cols := getTerminalWidth()

	// 9. Build section context
	ctx := &section.Context{
		Input:   inp,
		Config:  cfg,
		Locale:  locale,
		Colors:  colors,
		GitInfo: gitInfo,
		Effort:  effort,
	}

	// 10. Render sections per line
	allSections := section.AllSections()
	var line1Parts []string
	var line2Segments []render.ScoredSegment
	var line3Parts []string

	order := 0
	for _, ls := range allSections {
		if !cfg.IsSectionEnabled(ls.Section.Name()) {
			continue
		}
		text := ls.Section.Render(ctx)
		switch ls.Line {
		case 1:
			if text != "" {
				line1Parts = append(line1Parts, text)
			}
		case 2:
			line2Segments = append(line2Segments, render.ScoredSegment{
				Text:     text,
				Width:    ls.Section.Width(ctx),
				Priority: ls.Section.Priority(),
				Order:    order,
			})
			order++
		case 3:
			if text != "" {
				line3Parts = append(line3Parts, text)
			}
		}
	}

	// 11. Assemble and output
	compact := cfg.Preset == "compact"
	sep := cfg.Separator
	lines := render.AssembleLines(line1Parts, line2Segments, line3Parts, sep, cols, compact)

	if debug {
		fmt.Fprintf(os.Stderr, "[debug] config: preset=%s lang=%s user_type=%s\n", cfg.Preset, lang, cfg.UserType)
		fmt.Fprintf(os.Stderr, "[debug] input: model=%s ctx_used=%v\n", inp.Model.DisplayName, inp.ContextWindow.UsedPercentage)
		fmt.Fprintf(os.Stderr, "[debug] git: branch=%s is_repo=%v\n", gitInfo.Branch, gitInfo.IsRepo)
		fmt.Fprintf(os.Stderr, "[debug] terminal: cols=%d compact=%v\n", cols, compact)
		fmt.Fprintf(os.Stderr, "[debug] line1_parts=%d line2_segments=%d line3_parts=%d\n", len(line1Parts), len(line2Segments), len(line3Parts))
	}

	fmt.Print(strings.Join(lines, "\n"))
}

func getTerminalWidth() int {
	// Try to get terminal width from env or tty
	// Claude Code provides this context, but fallback to 200
	// We can't use syscall.TIOCGWINSZ easily since stdin is piped
	cols := 200
	// Check COLUMNS env var
	if c := os.Getenv("COLUMNS"); c != "" {
		var n int
		fmt.Sscanf(c, "%d", &n)
		if n > 0 {
			cols = n
		}
	}
	return cols
}
