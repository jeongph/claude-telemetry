---
description: Adjust Claude Code status line display (preset, sections, bar width)
allowed-tools: [Bash, Read, Write, AskUserQuestion]
---

# Status Line Display Configuration

You are helping the user adjust their existing status line display — preset, which
sections are shown, and how the progress bars look. This command ONLY edits
`~/.claude/statusline/config.json`; it does not install or remove anything.

## Execution Flow

Execute steps 1 → 2 → 3 → 4 → 5 → 6 → 7 sequentially. Do NOT skip steps. Do NOT combine steps.

---

## Step 1: Load Config & Detect Language

1. Read `~/.claude/statusline/config.json`.
2. **If it does not exist**, tell the user (translate to the language from settings.json, default en):
   > "No status line configuration found. Run `/claude-telemetry:setup` first to install and configure the status line."
   Then STOP. Do NOT create a config here.
3. If it exists, determine the language for user-facing text:
   - Use the config's `language` field if it is `ko`/`en`/`ja`/`zh`.
   - If it is `auto` or missing, read `~/.claude/settings.json` `language` and map `"한국어"` → ko, `"English"` → en, `"日本語"` → ja, `"中文"` → zh. Default: en.
4. Use this language for ALL user-facing text. Do NOT infer language from conversation context.

---

## Step 2: Show Current State

Compute and display the current effective settings, then continue to Step 3. Keep it to a few lines (translate labels):

- **Preset**: value of `preset`
- **Bar width**: value of `bar_width` (`0` = numbers only, no bar)
- **Enabled sections**: resolve each section using the 3-layer priority — `sections` override wins, otherwise the preset default (see Preset defaults table below). List the enabled ones.

Example (translate):
```
Current: preset=detailed, bar_width=5
Sections on: Context, Remaining, Cost, Git, PR, Elapsed, Agent, Vim, Tokens, ...
```

---

## Step 3: Choose Preset

Call AskUserQuestion (translate labels/descriptions):

```json
{
  "questions": [{
    "question": "<translated: Choose a preset (or keep the current one)>",
    "header": "Preset",
    "multiSelect": false,
    "options": [
      {"label": "<translated: Keep current>", "description": "<translated: Do not change the preset>"},
      {"label": "<translated: Normal>", "description": "<translated: Model, Elapsed, Git, PR, Context, Remaining, Agent, Vim>"},
      {"label": "<translated: Detailed>", "description": "<translated: All sections including Tokens, API Duration, Code Changes, Session>"},
      {"label": "<translated: Compact>", "description": "<translated: Single line — Model, Context, Remaining>"}
    ]
  }]
}
```

Map: "Keep current" → keep existing `preset`; "Normal" → `normal`; "Detailed" → `detailed`; "Compact" → `compact`.

---

## Step 4: Choose Bar Width

Call AskUserQuestion (translate labels/descriptions):

```json
{
  "questions": [{
    "question": "<translated: How detailed should the progress bars be?>",
    "header": "Bar",
    "multiSelect": false,
    "options": [
      {"label": "<translated: Keep current>", "description": "<translated: Do not change the bar width>"},
      {"label": "<translated: Numbers only>", "description": "◆ Context 54% <translated: (no bar)>"},
      {"label": "<translated: Normal bar (5)>", "description": "◆ Context ▰▰▰▱▱ 54%"},
      {"label": "<translated: Detailed bar (10)>", "description": "◆ Context ▰▰▰▰▰▱▱▱▱▱ 54%"}
    ]
  }]
}
```

Map: "Keep current" → keep existing `bar_width`; "Numbers only" → `0`; "Normal bar (5)" → `5`; "Detailed bar (10)" → `10`. If the user picks "Other" and types a number, use it as an integer (`0`, or `3`–`10`; values `1`–`2` become `3`).

---

## Step 5: Adjust Sections

Call AskUserQuestion (translate labels/descriptions):

```json
{
  "questions": [{
    "question": "<translated: Adjust individual sections?>",
    "header": "Sections",
    "multiSelect": false,
    "options": [
      {"label": "<translated: Keep current sections>", "description": "<translated: Leave section overrides unchanged>"},
      {"label": "<translated: Reset to preset defaults>", "description": "<translated: Clear all overrides — use the chosen preset's sections as-is>"},
      {"label": "<translated: Choose individually>", "description": "<translated: Turn each section on/off (4 quick questions)>"}
    ]
  }]
}
```

- **Keep current sections** → do NOT change the `sections` object.
- **Reset to preset defaults** → set `sections` to `{}` (empty).
- **Choose individually** → ask the four grouped multiSelect questions below (Step 5a). The section list can't fit in one question (max 4 options each), so it is split into 4 groups.

### Step 5a: Grouped section selection (only if "Choose individually")

Before asking, tell the user which sections are currently on (from Step 2) so they know the baseline. Selecting a section turns it ON; leaving it unselected turns it OFF. Ask all four (translate labels/descriptions):

```json
{
  "questions": [{
    "question": "<translated: Group 1/4 — Top line info (select the ones to show)>",
    "header": "Top line",
    "multiSelect": true,
    "options": [
      {"label": "Effort", "description": "<translated: Reasoning effort next to the model name>"},
      {"label": "◷ Elapsed", "description": "<translated: Session elapsed time>"},
      {"label": "Git", "description": "<translated: folder:branch, sync, changes, untracked, stash, worktree>"},
      {"label": "PR", "description": "<translated: Pull request number + review state>"}
    ]
  }]
}
```

```json
{
  "questions": [{
    "question": "<translated: Group 2/4 — Key metrics (select the ones to show)>",
    "header": "Metrics",
    "multiSelect": true,
    "options": [
      {"label": "◆ Context", "description": "<translated: Remaining context window %>"},
      {"label": "◆ Remaining", "description": "<translated: 5h/7d rate limit remaining>"},
      {"label": "Cost", "description": "<translated: Session cost in USD (API key users only)>"},
      {"label": "Token Details", "description": "<translated: Input/output token counts>"}
    ]
  }]
}
```

```json
{
  "questions": [{
    "question": "<translated: Group 3/4 — Detailed metrics (select the ones to show)>",
    "header": "Details",
    "multiSelect": true,
    "options": [
      {"label": "Code Changes", "description": "<translated: Lines added/removed in session>"},
      {"label": "↻ API Duration", "description": "<translated: Time spent waiting for API responses>"},
      {"label": "Session Name", "description": "<translated: Session title (Claude Code already shows it in its UI)>"}
    ]
  }]
}
```

```json
{
  "questions": [{
    "question": "<translated: Group 4/4 — Modes & account (select the ones to show)>",
    "header": "Modes",
    "multiSelect": true,
    "options": [
      {"label": "▶ Agent", "description": "<translated: Active subagent name>"},
      {"label": "Vim Mode", "description": "<translated: NORMAL/INSERT indicator>"},
      {"label": "✦ Thinking", "description": "<translated: Extended thinking indicator>"},
      {"label": "◉ User (email + plan)", "description": "<translated: Logged-in email + plan on a dedicated line — read from ~/.claude.json, hidden by default for privacy>"}
    ]
  }]
}
```

**If the user selects "◉ User (email + plan)"**, warn them before enabling (translate): the email is read from `~/.claude.json` and will be **visible in screenshots and screen shares** whenever the status line is shown. Only keep it if they are comfortable with that; otherwise treat it as unselected.

After all four questions, build the `sections` object with an **explicit** `true`/`false` for every one of the 15 keys below (selected → `true`, not selected → `false`). Do NOT include `model` (it is always shown).

### Section key mapping

| Option label | Config key |
|---|---|
| Effort | effort |
| ◷ Elapsed | elapsed |
| Git | git |
| PR | pr |
| ◆ Context | context |
| ◆ Remaining | ratelimit |
| Cost | cost |
| Token Details | tokens |
| Code Changes | lines |
| ↻ API Duration | apiduration |
| Session Name | session |
| ▶ Agent | agent |
| Vim Mode | vim |
| ✦ Thinking | thinking |
| ◉ User (email + plan) | user |

### Preset defaults (for resolving "currently on" and after "Reset to preset defaults")

| Section | compact | normal | detailed |
|---|---|---|---|
| model | on | on | on |
| effort | on | on | on |
| context | on | on | on |
| ratelimit | on | on | on |
| cost | on | on | on |
| elapsed | off | on | on |
| git | off | on | on |
| pr | off | on | on |
| agent | off | on | on |
| vim | off | on | on |
| lines | off | off | on |
| tokens | off | off | on |
| apiduration | off | off | on |
| session | off | off | on |
| thinking | off | off | on |
| user | off | off | off |

---

## Step 6: Write Configuration

1. Take the FULL existing `config.json` object and change ONLY these fields:
   - `preset` — from Step 3 (unchanged if "Keep current")
   - `bar_width` — from Step 4 (unchanged if "Keep current")
   - `sections` — from Step 5 ("Keep current" → unchanged; "Reset" → `{}`; "Choose individually" → explicit 15-key object)
2. Preserve every other field exactly (`language`, `colors`, `separator`, `user_type`, `thresholds`).
3. Write the result back to `~/.claude/statusline/config.json` (pretty-printed, 2-space indent).

---

## Step 7: Preview & Done

Render a live preview with the actual binary (this reflects the new settings exactly):

```bash
BIN=~/.claude/statusline/bin/claude-telemetry
SAMPLE="${CLAUDE_PLUGIN_ROOT}/testdata/normal.json"
if [ -x "$BIN" ] && [ -f "$SAMPLE" ]; then
  "$BIN" < "$SAMPLE"
else
  echo "PREVIEW_UNAVAILABLE"
fi
```

- If it prints a status line, show it to the user as the preview.
- If it prints `PREVIEW_UNAVAILABLE`, skip the live preview.

Then output (translate to detected language):

```
Display updated. Restart Claude Code to apply.
```

---

## Rules

- Be concise — no explanations unless asked
- Use detected language consistently
- Follow AskUserQuestion structures EXACTLY as specified above
- Only edit `~/.claude/statusline/config.json` — never touch settings.json, run.sh, or the binary
- Preserve all config fields not being changed
- `model` is always shown and is never toggleable
