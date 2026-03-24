# claude-telemetry

Customizable multi-line status line for [Claude Code](https://claude.com/claude-code).

<p align="center">
  <img src="assets/preview.svg" alt="claude-telemetry preview" width="720">
</p>

## Installation

### Via marketplace (recommended)

1. Add marketplace:

```
/plugin marketplace add jeongph/claude-telemetry
```

2. Install:

```
/plugin install claude-telemetry@jeongph-claude-telemetry
```

3. Run interactive setup:

```
/claude-telemetry:setup
```

### Manual setup

1. Clone the repo:

```bash
git clone https://github.com/jeongph/claude-telemetry.git
```

2. Copy the example config:

```bash
mkdir -p ~/.claude/statusline
```

```bash
cp claude-telemetry/config.example.json ~/.claude/statusline/config.json
```

3. Add to `~/.claude/settings.json`:

```json
"statusLine": {
  "type": "command",
  "command": "bash /path/to/claude-telemetry/scripts/run.sh"
}
```

4. Run interactive setup:

```
/claude-telemetry:setup
```

5. Restart Claude Code

## Features

- **Remaining % display** — all bars show remaining capacity (like a battery), not usage
- **Auto user detection** — OAuth users see rate limits, API key users see cost
- **Git integration** — folder:branch, ↑push/↓pull, changes (+/-), untracked (?N), stash (≡N), worktrees (⎇N)
- **Rate limit countdown** — remaining time until reset (2h 12m/5h)
- **200k token warning** — context size label turns bold yellow when exceeded
- **Progress bars** — ▰▱ visualization, color-coded green → yellow → red
- **Adaptive width** — auto-drops lower priority sections on narrow terminals
- **i18n** — English, Korean, Japanese, Chinese (auto-detected)
- **Configurable** — toggle sections, bar width, colors, language, user type

## Sections

| Line | Section | Description |
|------|---------|-------------|
| 1 | Model | Current model name |
| 1 | Elapsed | Session duration (Nh Nm format) |
| 1 | Git | folder:branch ↑push ↓pull +add/-del ?untracked ≡stash ⎇worktrees |
| 2 | Context | Remaining context window % with bar (yellow when >200k) |
| 2 | Rate Limits | Remaining 5h / 7d % with reset countdown (OAuth, auto-detected) |
| 2 | Cost | Session cost in USD (API key, auto-detected) |
| 2 | Lines | Session lines added/removed |
| 2 | API Duration | Time spent waiting for API responses |
| 2 | Tokens | Input/output token details |
| 3 | Agent | Active agent name (shown only when active) |
| 3 | Vim | Vim mode indicator (shown only when active) |

Lines 2 and 3 appear only when there is data to display.

## Setup

Run `/claude-telemetry:setup` in Claude Code for interactive configuration — it detects your language and walks you through section selection and style preferences.

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
  "language": "en",
  "user_type": "auto"
}
```

> **Note:** `user_type` can be `"auto"`, `"oauth"`, or `"api"`. Auto detects by `rate_limits` presence. Set `"oauth"` to prevent cost showing at session start before rate limits data arrives.

## Requirements

- Claude Code
- `jq` 1.6+
- `git` (for branch/changes display)

## License

MIT
