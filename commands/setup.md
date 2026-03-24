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

1. Detect OS:
   ```bash
   uname -s
   ```
   Map: `Linux` → `linux`, `Darwin` → `darwin`

2. Detect architecture:
   ```bash
   uname -m
   ```
   Map: `x86_64` → `amd64`, `aarch64` → `arm64`, `arm64` → `arm64`

3. Create the binary directory:
   ```bash
   mkdir -p ~/.claude/statusline/bin
   ```

4. Download the binary from GitHub Releases:
   ```bash
   curl -fsSL "https://github.com/jeongph/claude-telemetry/releases/latest/download/claude-telemetry-{os}-{arch}" \
     -o ~/.claude/statusline/bin/claude-telemetry
   ```
   (Replace `{os}` and `{arch}` with the detected values.)

5. Make it executable:
   ```bash
   chmod +x ~/.claude/statusline/bin/claude-telemetry
   ```

6. Verify the binary works:
   ```bash
   ~/.claude/statusline/bin/claude-telemetry --version
   ```
   If this fails, inform the user and stop.

---

## Step 3: Section Selection

Call AskUserQuestion with EXACTLY this structure (translate labels/descriptions to detected language):

```json
{
  "questions": [{
    "question": "<translated: Which sections would you like to enable?>",
    "header": "Sections",
    "multiSelect": false,
    "options": [
      {
        "label": "<translated: Recommended defaults>",
        "description": "◆ Context, Rate Limits, ◷ Elapsed, Git, ▶ Agent, Vim Mode"
      },
      {
        "label": "<translated: All ON>",
        "description": "<translated: Enable all 10 sections>"
      },
      {
        "label": "<translated: Minimal>",
        "description": "<translated: Context + Rate Limits only>"
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
      {"label": "Rate Limits", "description": "<translated: 5h/7d remaining % with reset countdown>"},
      {"label": "◷ Elapsed", "description": "<translated: Session elapsed time>"},
      {"label": "Git", "description": "<translated: folder:branch, sync, changes, untracked, stash, worktree>"}
    ]
  }]
}
```

Then a second AskUserQuestion for the remaining sections:

```json
{
  "questions": [{
    "question": "<translated: Select additional sections to enable>",
    "header": "Additional",
    "multiSelect": true,
    "options": [
      {"label": "Code Changes", "description": "<translated: Lines added/removed in session>"},
      {"label": "Cost", "description": "<translated: Session cost in USD (API key users only)>"},
      {"label": "↻ API Duration", "description": "<translated: Time spent waiting for API responses>"},
      {"label": "Token Details", "description": "<translated: Input/output token counts>"}
    ]
  }]
}
```

Note: Agent and Vim Mode are always ON in Custom mode (they only appear when active, no downside).

### Section key mapping

| Option label | Config key |
|---|---|
| ◆ Context | context |
| Rate Limits | rate_limits |
| ◷ Elapsed | duration |
| Git | git |
| Code Changes | lines |
| Cost | cost |
| ↻ API Duration | api_duration |
| Token Details | tokens |

### Preset mappings

| Preset | preset value | ON | OFF |
|--------|-------------|----|-----|
| Recommended | `"recommended"` | context, rate_limits, duration, git, agent, vim_mode | lines, cost, api_duration, tokens |
| All ON | `"all"` | ALL sections | (none) |
| Minimal | `"minimal"` | context, rate_limits, agent, vim_mode | duration, git, lines, cost, api_duration, tokens |
| Custom | `"custom"` | (user selection) | (user selection) |

---

## Step 4: Style Preferences

Call AskUserQuestion with EXACTLY this structure:

```json
{
  "questions": [{
    "question": "<translated: Choose your style preferences>",
    "header": "Style",
    "multiSelect": false,
    "options": [
      {
        "label": "<translated: Defaults>",
        "description": "<translated: Bar width: 5, Colors: ON, Labels: en>"
      },
      {
        "label": "<translated: Localized labels>",
        "description": "<translated: Bar width: 5, Colors: ON, Labels: (detected language)>"
      },
      {
        "label": "<translated: Compact>",
        "description": "<translated: Bar width: 3, Colors: ON, Labels: en>"
      }
    ]
  }]
}
```

### Style preset mappings

| Preset | bar_width | colors | language |
|--------|-----------|--------|----------|
| Defaults | 5 | true | "en" |
| Localized | 5 | true | (detected language code) |
| Compact | 3 | true | "en" |
| (Custom) | user input (3-10) | user input | user input |

---

## Step 5: Write Configuration

1. Run `mkdir -p ~/.claude/statusline`
2. Write `~/.claude/statusline/config.json` with the following structure:

```json
{
  "preset": "<preset_value>",
  "sections": {
    "git": <bool>,
    "context": <bool>,
    "rate_limits": <bool>,
    "duration": <bool>,
    "lines": <bool>,
    "cost": <bool>,
    "api_duration": <bool>,
    "tokens": <bool>,
    "agent": <bool>,
    "vim_mode": <bool>
  },
  "thresholds": {
    "context_warn": 20,
    "context_critical": 10,
    "rate_warn": 20,
    "rate_critical": 10
  },
  "colors": <bool>,
  "bar_width": <int>,
  "separator": " │ ",
  "language": "<lang_code>",
  "user_type": "auto"
}
```

- Set `preset` to the preset value from Step 3 (`"recommended"`, `"all"`, `"minimal"`, or `"custom"`).
- For non-custom presets, populate `sections` according to the preset mapping table in Step 3.
- For the `"custom"` preset, populate `sections` from the user's selections.

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

Output a preview using this EXACT template (substitute values based on user selections):

```
📋 설정 완료! 미리보기:

Line 1: Opus 4.6 │ ◷ Elapsed 12m 30s │ myproject:main ↑1 +15/-3 ?2
Line 2: ◆ Context ▰▰▰▱▱ 55% (200k) │ 3h 45m/5h ▰▰▰▰▱ 70%  6d 12h/7d ▰▰▰▰▰ 94%
Line 3: ▶ code-explorer

Claude Code를 재시작하면 적용됩니다.
```

- Only show sections the user enabled
- Adjust bar width to match user's choice
- Translate the message to detected language

---

## Rules

- Be concise — no explanations unless asked
- Use detected language consistently
- Follow AskUserQuestion structures EXACTLY as specified above
- Do NOT modify `${CLAUDE_PLUGIN_ROOT}/scripts/run.sh`
- Do NOT output text between AskUserQuestion calls except brief transitions
