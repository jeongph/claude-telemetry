# claude-telemetry

Customizable multi-line status line for [Claude Code](https://claude.com/claude-code).

**"The status line you can trust"** — accurate, lightweight, never breaks.

<p align="center">
  <img width="810" height="616" alt="image" src="https://github.com/user-attachments/assets/3eb1c1a5-a8b0-48ef-8f26-d6b691374a33" />
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

This downloads the Go binary, configures your preset, and sets up the status line.

### Manual setup

1. Download the binary for your platform from [Releases](https://github.com/jeongph/claude-telemetry/releases/latest):

```bash
mkdir -p ~/.claude/statusline/bin
curl -fsSL "https://github.com/jeongph/claude-telemetry/releases/latest/download/claude-telemetry-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m | sed 's/x86_64/amd64/' | sed 's/aarch64/arm64/')" \
  -o ~/.claude/statusline/bin/claude-telemetry
chmod +x ~/.claude/statusline/bin/claude-telemetry
```

2. (Optional) Copy the example config:

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

4. Restart Claude Code

## Features

- **Remaining % display** — all bars show remaining capacity (like a battery), not usage
- **Preset modes** — compact (1 line), normal (2 lines), detailed (3 lines)
- **Auto user detection** — OAuth users see rate limits, API key users see cost
- **Git integration** — folder:branch, ↑push/↓pull, changes (+/-), untracked (?N), stash (≡N), worktrees (⎇N)
- **Effort level** — live reasoning effort (low/medium/high/xhigh/max) shown beside the model name, reflects `/effort` changes (Claude Code ≥ 2.1.141)
- **PR badge** — open PR number and review state for the current branch, no `gh` CLI needed (Claude Code ≥ 2.1.145)
- **Session name** — session title shown as `[name]`, auto-truncated to 20 columns (off by default — Claude Code already shows the title in its UI; enable via `sections` or the detailed preset)
- **User identity** — logged-in email + plan (`Max`/`Pro`/`Team`) on a dedicated line, read from `~/.claude.json` (not in the status line JSON). Off by default for privacy — enable via `sections.user`
- **Rate limit countdown** — remaining time until reset with progress bar
- **Dynamic color thresholds** — green/yellow/red based on remaining %, customizable via config
- **Graceful degradation** — loading (···), partial failure (—), error messages instead of silent blank
- **Progress bars** — ▰▱ visualization, color-coded green → yellow → red
- **Adaptive width** — auto-drops lower priority sections on narrow terminals
- **i18n** — English, Korean, Japanese, Chinese (auto-detected)
- **Auto binary sync** — a SessionStart hook keeps the binary matched to the plugin version (pinned download + sha256 verification)
- **Self-cleanup on uninstall** — if you uninstall the plugin, the status line removes its own settings entry and files within a minute (setup-managed installs only)
- **NO_COLOR support** — respects `NO_COLOR` environment variable
- **Go binary** — single binary, no runtime dependencies, sub-10ms rendering
- **v1 fallback** — existing jq-based users keep working until they upgrade

## Sections

| Line | Section | Description |
|------|---------|-------------|
| 1 | Session | `[name]` session title (max 20 cols, detailed preset or opt-in) |
| 1 | Model | Model name with effort level beside it (`Fable 5 · xhigh`), color-coded low→max (toggle via `effort` key) |
| 1 | Elapsed | Session duration (Nh Nm format) |
| 1 | Git | folder:branch ↑push ↓pull +add/-del ?untracked ≡stash ⎇worktrees |
| 1 | PR | Open PR number + review state ✓/●/✗/◌ (shown only when a PR is open) |
| 2 | Context | ◆ Remaining context window % with progress bar |
| 2 | Remaining | ◆ 5h / 7d remaining % with reset countdown (OAuth, auto-detected) |
| 2 | Cost | Session cost in USD (API key, auto-detected) |
| 2 | Lines | Session lines added/removed |
| 2 | API Duration | Time spent waiting for API responses |
| 2 | Tokens | Tokens currently in the context window (in/out) |
| 3 | Agent | Active agent name (shown only when active) |
| 3 | Vim | Vim mode indicator (shown only when active) |
| 3 | Thinking | ✦ extended thinking indicator (shown only when enabled) |
| 4 | User | ◉ logged-in email + plan on a dedicated line (off by default, opt-in) |

Line 3 appears only when agent, vim mode, or thinking indicator is active. Line 4 appears only when the `user` section is enabled.

> **Note:** Since Claude Code 2.1.132, token counts reflect what is currently in the context window, not cumulative session totals.

### Git status symbols

The Git section renders as `folder:branch` followed by status markers. Each marker appears **only when its count is non-zero**, so a clean repo shows just `folder:branch`.

| Symbol | Meaning | Color |
|--------|---------|-------|
| `folder:branch` | Current directory and current branch | white `:` magenta |
| `↑N` | N commits ahead of upstream (waiting to push) | yellow |
| `↓N` | N commits behind upstream (waiting to pull) | cyan |
| `+N/-N` | Lines added / deleted vs. HEAD (staged + unstaged) | green / red |
| `?N` | N untracked files | yellow |
| `≡N` | N stash entries | magenta |
| `⎇N` | N linked worktrees (excludes the main worktree) | cyan |

> Example: `lighthouse:main ↑1 +12/-3 ?2 ⎇1` means branch `main` is 1 commit ahead of upstream, has 12 added / 3 deleted lines, 2 untracked files, and 1 linked worktree.

## Setup

Run `/claude-telemetry:setup` in Claude Code for interactive configuration — it detects your language, downloads the binary, and walks you through preset selection.

Or edit `~/.claude/statusline/config.json` directly:

```json
{
  "preset": "normal",
  "language": "en",
  "colors": true,
  "bar_width": 5,
  "separator": " │ ",
  "user_type": "auto",
  "sections": {},
  "thresholds": {
    "context_warn": 50,
    "context_danger": 20,
    "cost_warn": 1.0,
    "cost_danger": 5.0
  }
}
```

### Presets

| Preset | Lines | Sections |
|--------|-------|----------|
| `compact` | 1 | Model · Effort, Context, Remaining/Cost |
| `normal` | 2 | Model · Effort, Elapsed, Git, PR, Context, Remaining/Cost, Agent, Vim |
| `detailed` | 3 | All sections enabled |

### Section overrides

Use `sections` to override preset defaults:

```json
{
  "preset": "normal",
  "sections": {
    "tokens": true,
    "lines": true
  }
}
```

### User section (email + plan)

The `user` section shows your logged-in email and plan on a dedicated line (e.g. `◉ you@example.com · Max`). It is **off by default** — enable it explicitly:

```json
{
  "sections": {
    "user": true
  }
}
```

- **Source:** this info is not part of the status line JSON. It is read from `~/.claude.json` (`oauthAccount`), an internal Claude Code file, and parsed defensively — if the file or fields are missing, the section is silently skipped.
- **Privacy:** the status line is visible in screenshots and screen shares. Keep it off unless you want your email on screen at all times.
- **Plan labels:** `claude_max` → `Max`, `claude_pro` → `Pro`, `claude_team` → `Team`, `claude_enterprise` → `Enterprise`. Unknown plans are omitted (email only).

### Thresholds

Color changes at these remaining percentages (customizable):

| Remaining | Color |
|-----------|-------|
| > 50% | Green |
| 21–50% | Yellow |
| ≤ 20% | Red |

### Project-level config

Create `.claude-statusline.json` in your project root to override global settings per project:

```json
{
  "preset": "detailed"
}
```

## Removal

```
/claude-telemetry:remove
```

If you uninstall the plugin without running remove first, the status line detects the missing plugin and cleans itself up automatically within about a minute (settings entry removed from the next session). This applies to installs managed by `/claude-telemetry:setup`; manual installs are never touched.

## Upgrading

- **Plugin users (v2.4.0+):** update the plugin (`/plugin` → Update), then restart Claude Code. The SessionStart hook syncs the binary to the plugin version automatically.
- **Plugin users (older):** run `/claude-telemetry:setup` once after updating the plugin — it downloads the matching binary and migrates your settings to the version-independent launcher path.
- **Manual installs:** re-run the curl command from Manual setup; the binary is all that matters.

## Upgrading from v1

v2 is backward-compatible. Existing v1 config files work as-is. Run `/claude-telemetry:setup` to download the Go binary — your existing settings are preserved.

If you don't run setup, the v1 jq-based rendering continues to work via the built-in fallback.

## Requirements

- Claude Code
- `git` (optional, for branch/changes display)
- Claude Code ≥ 2.1.141 for Effort, ≥ 2.1.145 for PR badge (older versions simply hide these sections)

## License

MIT
