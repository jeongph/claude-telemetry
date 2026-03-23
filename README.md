# claude-telemetry

Customizable multi-line status line for [Claude Code](https://claude.com/claude-code).

```
Opus │ ◷ Elapsed 12m34s │ main ↑1 +3/-1
◆ Context ▰▰▱▱▱ 35% (1M) │ 5h ▰▱▱▱▱ 24%  7d ▰▰▰▰▱ 71%
▶ code-explorer │ NORMAL
```

## Installation

### Via marketplace (recommended)

```bash
# 1. Add marketplace
/plugin marketplace add jeongph/claude-telemetry

# 2. Install
/plugin install claude-telemetry@jeongph-claude-telemetry
```

### Manual setup

1. Clone the repo:

```bash
git clone https://github.com/jeongph/claude-telemetry.git
```

2. Copy the example config:

```bash
mkdir -p ~/.claude/statusline
cp claude-telemetry/config.example.json ~/.claude/statusline/config.json
```

3. Add to `~/.claude/settings.json`:

```json
{
  "statusLine": {
    "type": "command",
    "command": "bash /path/to/claude-telemetry/scripts/run.sh"
  }
}
```

4. Restart Claude Code

## Features

- **Multi-line layout** — identity & git on top, metrics in the middle, agent/vim at the bottom
- **Auto user detection** — OAuth users see rate limits, API key users see cost automatically
- **Git integration** — branch, ahead/behind (↑↓), uncommitted changes (+/-)
- **Color-coded thresholds** — green/yellow/red based on usage percentage
- **200k token warning** — context size label turns bold yellow when exceeded
- **Progress bars** — ▰▱ visualization for usage percentages
- **Adaptive width** — auto-drops lower priority sections on narrow terminals
- **i18n** — English, Korean, Japanese, Chinese (auto-detected)
- **Configurable** — toggle sections, bar width, colors, language

## Sections

| Line | Section | Description |
|------|---------|-------------|
| 1 | Model | Current model name |
| 1 | Elapsed | Session duration |
| 1 | Git | Branch + ↑push/↓pull + changes (+/-) |
| 2 | Context | Context window usage with bar (size label turns yellow when >200k) |
| 2 | Rate Limits | 5h / 7d rolling window usage (OAuth users only, auto-detected) |
| 2 | Cost | Session cost in USD (API key users only, auto-detected) |
| 2 | Lines | Session lines added/removed |
| 2 | API Duration | Time spent waiting for API responses |
| 2 | Tokens | Input/output token details |
| 3 | Agent | Active agent name (shown only when active) |
| 3 | Vim | Vim mode indicator (shown only when active) |

Lines 2 and 3 appear only when there is data to display.

## Setup

Run `/setup` in Claude Code for interactive configuration — it detects your language and walks you through section selection and style preferences.

Or edit `~/.claude/statusline/config.json` directly:

```json
{
  "sections": {
    "git": true,
    "context": true,
    "rate_limits": true,
    "duration": true,
    "lines": false,
    "cost": false,
    "api_duration": false,
    "tokens": false,
    "agent": true,
    "vim_mode": true
  },
  "colors": true,
  "bar_width": 5,
  "separator": " │ ",
  "language": "en"
}
```

> **Note:** `rate_limits` and `cost` are auto-detected by user type. OAuth (Pro/Max) users see rate limits; API key users see cost. The config toggle only applies when the section is relevant to your user type.

## Requirements

- Claude Code
- `jq` 1.6+
- `git` (for branch/changes display)

## License

MIT
