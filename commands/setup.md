---
description: Interactive setup for Claude Code status line
allowed-tools: [Bash, Read, Write, Edit, AskUserQuestion]
---

# Status Line Interactive Setup

You are helping the user configure their Claude Code status line.

## Execution Flow

Execute steps 1 → 2 → 3 → 4 → 5 → 6 → 7 sequentially. Do NOT skip steps. Do NOT combine steps.

---

## Step 1: Detect Language

1. Read `~/.claude/settings.json`
2. Check the `language` field
3. Map: `"한국어"` → ko, `"English"` → en, `"日本語"` → ja, `"中文"` → zh. Default: en
4. Use this language for ALL user-facing text in the remaining steps

---

## Step 2: Download Go Binary

1. Check if Go binary already exists:
   ```bash
   ~/.claude/statusline/bin/claude-telemetry --version 2>/dev/null
   ```
   If it exists, show the current version and ask if user wants to update. If user says no, skip to Step 3.

2. Detect OS:
   ```bash
   uname -s
   ```
   Map: `Linux` → `linux`, `Darwin` → `darwin`

3. Detect architecture:
   ```bash
   uname -m
   ```
   Map: `x86_64` → `amd64`, `aarch64` → `arm64`, `arm64` → `arm64`

4. Create the binary directory:
   ```bash
   mkdir -p ~/.claude/statusline/bin
   ```

5. Download the binary from GitHub Releases:
   ```bash
   curl -fsSL "https://github.com/jeongph/claude-telemetry/releases/latest/download/claude-telemetry-{os}-{arch}" \
     -o ~/.claude/statusline/bin/claude-telemetry
   ```
   (Replace `{os}` and `{arch}` with the detected values.)

6. Make it executable:
   ```bash
   chmod +x ~/.claude/statusline/bin/claude-telemetry
   ```

7. Verify the binary works:
   ```bash
   ~/.claude/statusline/bin/claude-telemetry --version
   ```
   If this fails, inform the user and stop.

---

## Step 3: Check Existing Configuration

1. Check if `~/.claude/statusline/config.json` exists
2. If exists:
   - Inform user: "<translated: Existing configuration found. Your current settings will be preserved.>"
   - **Skip to Step 6** (keep existing config, just ensure statusLine is configured)
3. If not exists: proceed to Step 4

---

## Step 4: Section Selection

Call AskUserQuestion with EXACTLY this structure (translate labels/descriptions to detected language):

```json
{
  "questions": [{
    "question": "<translated: Choose a display preset>",
    "header": "Preset",
    "multiSelect": false,
    "options": [
      {
        "label": "<translated: Normal (Recommended)>",
        "description": "Model, ◷ Elapsed, Git, ◆ Context, ◆ Remaining, ▶ Agent, Vim"
      },
      {
        "label": "<translated: Detailed>",
        "description": "<translated: All sections enabled including Tokens, API Duration, Code Changes>"
      },
      {
        "label": "<translated: Compact>",
        "description": "<translated: Single line — Model, Context, Remaining only>"
      },
      {
        "label": "<translated: Custom>",
        "description": "<translated: Choose each section individually>"
      }
    ]
  }]
}
```

**If user selects "Custom"**, call AskUserQuestion again with multiSelect:

```json
{
  "questions": [{
    "question": "<translated: Select sections to enable (model name is always shown)>",
    "header": "Sections",
    "multiSelect": true,
    "options": [
      {"label": "◆ Context", "description": "<translated: Remaining context window % with progress bar>"},
      {"label": "◆ Remaining", "description": "<translated: 5h/7d remaining % with reset countdown>"},
      {"label": "◷ Elapsed", "description": "<translated: Session elapsed time>"},
      {"label": "Git", "description": "<translated: folder:branch, sync, changes, untracked, stash, worktree>"},
      {"label": "Code Changes", "description": "<translated: Lines added/removed in session>"},
      {"label": "Cost", "description": "<translated: Session cost in USD (API key users only)>"},
      {"label": "↻ API Duration", "description": "<translated: Time spent waiting for API responses>"},
      {"label": "Token Details", "description": "<translated: Input/output token counts>"}
    ]
  }]
}
```

Note: Agent and Vim Mode are always ON in Custom mode (they only appear when active, no downside).

### Preset mappings

| Preset | preset value | Config sections override |
|--------|-------------|------------------------|
| Normal | `"normal"` | (empty — use preset defaults) |
| Detailed | `"detailed"` | (empty — use preset defaults) |
| Compact | `"compact"` | (empty — use preset defaults) |
| Custom | `"normal"` | (user selection as overrides) |

### Custom section key mapping

| Option label | Config key |
|---|---|
| ◆ Context | context |
| ◆ Remaining | ratelimit |
| ◷ Elapsed | elapsed |
| Git | git |
| Code Changes | lines |
| Cost | cost |
| ↻ API Duration | apiduration |
| Token Details | tokens |

---

## Step 5: Write Configuration

1. Run `mkdir -p ~/.claude/statusline`
2. Write `~/.claude/statusline/config.json`:

```json
{
  "preset": "<preset_value>",
  "language": "<lang_code>",
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

- Set `preset` from Step 4 selection
- For Custom: populate `sections` with user selections (only non-default values needed)
- Set `language` to detected language code, or `"en"` for Defaults

---

## Step 6: Configure statusLine in settings.json

1. Resolve the script path: `${CLAUDE_PLUGIN_ROOT}/scripts/run.sh`
2. Read `~/.claude/settings.json`
3. Check the `statusLine` field:
   - **Not present** → add it
   - **Present and same script** → skip (tell user it's already configured)
   - **Present but different** → ask user with AskUserQuestion whether to replace
4. The statusLine entry must be:
```json
"statusLine": {
  "type": "command",
  "command": "bash <resolved-absolute-path>/scripts/run.sh"
}
```

---

## Step 7: Preview & Done

Output a preview using this EXACT template (substitute values based on user selections, translate to detected language):

```
Setup complete! Preview:

Line 1: Opus │ ◷ Elapsed 12m 30s │ myproject:main ↑1 +15/-3 ?2
Line 2: ◆ Context ▰▰▰▱▱ 55% │ ◆ Remaining 5h ▰▰▰▰▱ 70% (3h 45m) / 7d ▰▰▰▰▰ 94% (6d 12h)
Line 3: ▶ code-explorer │ NORMAL

Restart Claude Code to apply.
```

- Only show sections the user enabled
- Adjust bar width to match user's choice
- For compact preset, show single line preview
- Translate the message to detected language

---

## Rules

- Be concise — no explanations unless asked
- Use detected language consistently
- Follow AskUserQuestion structures EXACTLY as specified above
- Do NOT modify `${CLAUDE_PLUGIN_ROOT}/scripts/run.sh`
- Do NOT output text between AskUserQuestion calls except brief transitions
- Preserve existing config.json if it already exists (Step 3)
