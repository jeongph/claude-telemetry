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
3. Map: `"한국어"` → ko, `"English"` → en, `"日本語"` → ja, `"中文"` → zh
4. If the `language` field is missing or does not match any mapping → ask the user:

```json
{
  "questions": [{
    "question": "Select your preferred language for the status line",
    "header": "Language",
    "multiSelect": false,
    "options": [
      {"label": "English"},
      {"label": "한국어"},
      {"label": "日本語"},
      {"label": "中文"}
    ]
  }]
}
```

Map the selection: "English" → en, "한국어" → ko, "日本語" → ja, "中文" → zh

5. Use this language for ALL user-facing text in the remaining steps. Do NOT infer language from conversation context.

---

## Step 2: Download Go Binary

Run this EXACT script as a single Bash command. Do NOT split it into multiple commands:

```bash
PLUGIN_VER=$(sed -n 's/.*"version"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' "${CLAUDE_PLUGIN_ROOT}/.claude-plugin/plugin.json" | head -1)
EXISTING_VER=$(~/.claude/statusline/bin/claude-telemetry --version 2>/dev/null | sed 's/^claude-telemetry v\{0,1\}//')
if [ "$EXISTING_VER" = "$PLUGIN_VER" ]; then
  echo "UP_TO_DATE: v$EXISTING_VER"
else
  OS=$(uname -s | tr '[:upper:]' '[:lower:]')
  ARCH=$(uname -m)
  case "$ARCH" in
    x86_64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
  esac
  NAME="claude-telemetry-${OS}-${ARCH}"
  BASE="https://github.com/jeongph/claude-telemetry/releases/download/v${PLUGIN_VER}"
  mkdir -p ~/.claude/statusline/bin
  TMP=$(mktemp -p ~/.claude/statusline/bin)
  echo "Downloading: ${BASE}/${NAME}"
  if curl -fsSL "${BASE}/${NAME}" -o "$TMP"; then
    SUM=$(curl -fsSL "${BASE}/checksums.txt" | awk -v n="$NAME" '$2 == n {print $1}')
    if command -v sha256sum >/dev/null 2>&1; then ACTUAL=$(sha256sum "$TMP" | awk '{print $1}'); else ACTUAL=$(shasum -a 256 "$TMP" | awk '{print $1}'); fi
    if [ -n "$SUM" ] && [ "$ACTUAL" = "$SUM" ]; then
      chmod +x "$TMP" && mv -f "$TMP" ~/.claude/statusline/bin/claude-telemetry && \
      ~/.claude/statusline/bin/claude-telemetry --version && echo "SUCCESS"
    else
      rm -f "$TMP"; echo "CHECKSUM_FAILED"
    fi
  else
    rm -f "$TMP"; echo "DOWNLOAD_FAILED"
  fi
fi
```

Handle the output:
- If `UP_TO_DATE: ...` → inform user the binary is already at the current version, proceed to Step 3
- If `SUCCESS` → proceed to Step 3
- If `CHECKSUM_FAILED` → inform user the checksum verification failed and stop
- If `DOWNLOAD_FAILED` → inform user the download failed and stop

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
        "description": "Model · Effort, ◷ Elapsed, Git, PR, ◆ Context, ◆ Remaining, ▶ Agent, Vim"
      },
      {
        "label": "<translated: Detailed>",
        "description": "<translated: All sections enabled including Tokens, API Duration, Code Changes>"
      },
      {
        "label": "<translated: Compact>",
        "description": "<translated: Single line — Model · Effort, Context, Remaining>"
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
      {"label": "Token Details", "description": "<translated: Input/output token counts>"},
      {"label": "◉ User (email + plan)", "description": "<translated: Logged-in email + plan on a dedicated line — read from ~/.claude.json, hidden by default for privacy>"}
    ]
  }]
}
```

Note: Agent, Vim Mode, Effort, and PR are always ON in Custom mode (they only appear when relevant, no downside). Session Name is OFF by default (Claude Code already shows the title in its UI) — users can enable it via `"sections": {"session": true}`.

**If the user selects "◉ User (email + plan)"**, warn them before enabling (translate to detected language): the email is read from `~/.claude.json` and will be **visible in screenshots and screen shares** whenever the status line is shown. Only enable it if they are comfortable with that. If they decline, drop `user` from the selection.

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
| ◉ User (email + plan) | user |

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
- For Custom: populate `sections` with user selections (only non-default values needed). If ◉ User was selected and confirmed, add `"user": true` — it renders a dedicated line (Line 4) with email + plan.
- Set `language` to detected language code, or `"en"` for Defaults

---

## Step 6: Configure statusLine in settings.json

1. ALWAYS copy the launcher first (run as a single Bash command, even if statusLine is already configured — this keeps the launcher up to date across plugin updates):
   ```bash
   mkdir -p ~/.claude/statusline
   cp "${CLAUDE_PLUGIN_ROOT}/scripts/run.sh" ~/.claude/statusline/run.sh
   touch ~/.claude/statusline/.managed-by-plugin
   ```
   (the marker file enables automatic cleanup when the plugin is uninstalled)
2. Read `~/.claude/settings.json`
3. Check the `statusLine` field:
   - **Not present** → add it
   - **Present and same script** → skip the settings.json edit only (tell user it's already configured; the copy in step 1 still runs)
   - **Present but different** (including old plugin-cache paths) → ask user with AskUserQuestion whether to replace
4. The statusLine entry must be (resolve `<home>` to the absolute home directory):
```json
"statusLine": {
  "type": "command",
  "command": "bash <home>/.claude/statusline/run.sh"
}
```

---

## Step 7: Preview & Done

Output a preview using this EXACT template (substitute values based on user selections, translate to detected language):

```
Setup complete! Preview:

Line 1: Opus · high │ ◷ Elapsed 12m 30s │ myproject:main ↑1 +15/-3 ?2 │ ✓ PR#12
Line 2: ◆ Context ▰▰▰▱▱ 55% │ ◆ Remaining 5h ▰▰▰▰▱ 70% (3h 45m) / 7d ▰▰▰▰▰ 94% (6d 12h)
Line 3: ▶ code-explorer │ NORMAL
Line 4: ◉ you@example.com · Max

Restart Claude Code to apply.
From now on the binary stays in sync with the plugin automatically (checked at session start).
```

- Only show sections the user enabled (show Line 4 only if ◉ User was enabled)
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
