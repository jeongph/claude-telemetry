# claude-telemetry

Customizable 2-line status line for [Claude Code](https://claude.com/claude-code).

```
Opus │ main +3/-1
◷ Elapsed 12m34s │ ◆ Context ▰▰▱▱▱ 35% (1M) │ 5h ▰▱▱▱▱ 24%  7d ▰▰▰▰▱ 71%
```

## Features

- **2-line layout** — Line 1: model + git, Line 2: session metrics
- **Git integration** — branch name + uncommitted changes (+/-)
- **Color-coded thresholds** — green/yellow/red for context, rate limits, cost
- **Progress bars** — ▰▱ visualization for usage percentages
- **Adaptive width** — auto-drops lower priority sections on narrow terminals
- **i18n** — English, Korean, Japanese, Chinese (auto-detected)
- **Configurable** — toggle sections, bar width, colors, language

## Sections

| Line | Section | Description |
|------|---------|-------------|
| 1 | Model | Current model name |
| 1 | Git | Branch + changes (+/-) |
| 1 | Agent | Active agent name |
| 1 | Vim | Vim mode indicator |
| 2 | Elapsed | Session duration |
| 2 | Context | Context window usage with bar |
| 2 | Rate Limits | 5h / 7d rolling window usage |
| 2 | Warning | 200k token exceeded alert |
| 2 | Cost | Session cost in USD (API key users) |
| 2 | Tokens | Input/output token details |

## Installation

### As Claude Code plugin (recommended)

```bash
# Clone the repo
git clone https://github.com/jeongph/claude-telemetry.git

# Register as plugin in Claude Code
# (plugin installation via marketplace coming soon)
```

### Manual setup

1. Copy `scripts/run.sh` to your preferred location
2. Copy `config.example.json` to `~/.claude/statusline/config.json`
3. Add to `~/.claude/settings.json`:

```json
{
  "statusLine": {
    "type": "command",
    "command": "bash /path/to/scripts/run.sh"
  }
}
```

4. Restart Claude Code

## Configuration

Run `/setup` in Claude Code for interactive configuration, or edit `~/.claude/statusline/config.json` directly:

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
