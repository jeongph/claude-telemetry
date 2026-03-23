# claude-telemetry

Customizable multi-line status line for [Claude Code](https://claude.com/claude-code).

```
Opus │ main ↑1 +3/-1
◷ Elapsed 12m34s │ ◆ Context ▰▰▱▱▱ 35% (1M) │ 5h ▰▱▱▱▱ 24%  7d ▰▰▰▰▱ 71%
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

- **Multi-line layout** — identity & git on top, session metrics below
- **Git integration** — branch, ahead/behind (↑↓), uncommitted changes (+/-)
- **Color-coded thresholds** — green/yellow/red for context, rate limits, cost
- **Progress bars** — ▰▱ visualization for usage percentages
- **Adaptive width** — auto-drops lower priority sections on narrow terminals
- **i18n** — English, Korean, Japanese, Chinese (auto-detected)
- **Configurable** — toggle sections, bar width, colors, language

## Sections

| Line | Section | Description |
|------|---------|-------------|
| 1 | Model | Current model name |
| 1 | Git | Branch + ↑push/↓pull + changes (+/-) |
| 1 | Agent | Active agent name |
| 1 | Vim | Vim mode indicator |
| 2 | Elapsed | Session duration |
| 2 | Context | Context window usage with bar |
| 2 | Rate Limits | 5h / 7d rolling window usage |
| 2 | Warning | 200k token exceeded alert |
| 2 | Cost | Session cost in USD (API key users) |
| 2 | Tokens | Input/output token details |

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
    "warn_200k": true,
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

## Requirements

- Claude Code
- `jq` 1.6+
- `git` (for branch/changes display)

## License

MIT
