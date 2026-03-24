---
description: Remove Claude Code status line configuration
allowed-tools: [Bash, Read, Edit, AskUserQuestion]
---

# Status Line Removal

You are removing the claude-telemetry status line configuration.

## Execution Flow

Execute steps 1 → 2 → 3 → 4 sequentially. Do NOT skip steps.

---

## Step 1: Detect Language

1. Read `~/.claude/settings.json`
2. Check the `language` field
3. Map: `"한국어"` → ko, `"English"` → en, `"日本語"` → ja, `"中文"` → zh. Default: en
4. Use this language for ALL user-facing text

---

## Step 2: Confirm Removal

Call AskUserQuestion with EXACTLY this structure (translate to detected language):

```json
{
  "questions": [{
    "question": "<translated: Remove the status line configuration? This will delete config, binary, cache, and disable the status line.>",
    "header": "Remove",
    "multiSelect": false,
    "options": [
      {
        "label": "<translated: Remove all>",
        "description": "<translated: Delete config, binary, cache directory, and remove statusLine from settings.json>"
      },
      {
        "label": "<translated: Config only>",
        "description": "<translated: Delete config file only, keep binary and statusLine entry in settings.json>"
      },
      {
        "label": "<translated: Cancel>",
        "description": "<translated: Do nothing>"
      }
    ]
  }]
}
```

**If user selects "Cancel"**, output a cancellation message and stop.

---

## Step 3: Remove Configuration

Based on user choice:

### "Remove all"

1. Remove the Go binary:
   ```bash
   rm -f ~/.claude/statusline/bin/claude-telemetry
   ```
2. Remove the git cache directory:
   ```bash
   rm -rf ~/.claude/statusline/cache/
   ```
3. Delete the full config directory:
   ```bash
   rm -rf ~/.claude/statusline/
   ```
4. Read `~/.claude/settings.json`
5. Remove the `"statusLine"` entry from the JSON using Edit tool
6. Verify `settings.json` is still valid JSON:
   ```bash
   jq . ~/.claude/settings.json > /dev/null
   ```

### "Config only"

1. Delete config file only:
   ```bash
   rm -f ~/.claude/statusline/config.json
   ```

---

## Step 4: Done

Output a completion message using this EXACT template (translate to detected language):

```
Status line 설정이 제거되었습니다. Claude Code를 재시작하면 적용됩니다.

다시 설정하려면: /claude-telemetry:setup
```

---

## Rules

- Be concise — no explanations unless asked
- Use detected language consistently
- Follow AskUserQuestion structure EXACTLY as specified
- Do NOT modify `${CLAUDE_PLUGIN_ROOT}/scripts/run.sh`
- Do NOT uninstall the plugin itself — only remove configuration
