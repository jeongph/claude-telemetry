---
description: Interactive setup for Claude Code status line
---

# Status Line Interactive Setup

You are helping the user configure their Claude Code status line at `~/.claude/statusline/`.

## Step 1: Detect Language

Read `~/.claude/settings.json` and check the `language` field. Use that language for ALL communication in this setup. If "한국어" → Korean, if "English" → English, etc. Default to English.

## Step 2: Present Sections

Present the following sections as a numbered list. Mark recommended items. Include a short description of what each section shows. The model name section is always shown and not configurable.

| Key | Icon | Name (en) | Name (ko) | Description (en) | Description (ko) | Default |
|-----|------|-----------|-----------|-------------------|-------------------|---------|
| context | ◆ | Context | 컨텍스트 | Context window usage with progress bar | 컨텍스트 윈도우 사용률 (프로그레스 바) | ON |
| rate_limits | | Rate Limits | Rate Limits | 5h / 7d rolling window usage | 5시간/7일 사용량 제한 | ON |
| duration | ◷ | Elapsed | 경과 시간 | Session elapsed time | 세션 경과 시간 | ON |
| lines | | Code Changes | 코드 변경 | Lines added/removed in session | 세션 중 추가/삭제된 코드 라인 | ON |
| warn_200k | ▲ | 200k Warning | 200k 경고 | Alert when tokens exceed 200k | 토큰 200k 초과 시 경고 | ON |
| cost | | Cost | 비용 | Session cost in USD (API key users only) | 세션 비용 - USD (API 키 사용자 전용) | OFF |
| api_duration | ↻ | API Duration | API 대기 | Time spent waiting for API responses | API 응답 대기 시간 합계 | OFF |
| tokens | | Token Details | 토큰 상세 | Input/output token counts | 입출력 토큰 수 상세 | OFF |
| worktree | ⎇ | Worktree | 워크트리 | Git worktree name (shown only when active) | Git worktree 이름 (활성 시에만 표시) | ON |
| agent | ▶ | Agent | 에이전트 | Agent name (shown only when active) | 에이전트 이름 (활성 시에만 표시) | ON |
| vim_mode | | Vim Mode | Vim 모드 | Current vim mode (shown only when enabled) | 현재 Vim 모드 (활성 시에만 표시) | ON |

Ask the user to select which sections to enable. Suggest they can just press Enter to accept the recommended defaults, or type numbers to customize.

## Step 3: Style Preferences

After section selection, ask about style in a single question:

1. **Progress bar width** (3-10, default: 5)
2. **Colors** (on/off, default: on)
3. **Display language** for labels (auto/en/ko/ja/zh, default: en)

Again, let them press Enter to accept defaults or customize.

## Step 4: Apply Configuration

Based on user choices, write `~/.claude/statusline/config.json`:

```json
{
  "sections": {
    "context": true/false,
    "rate_limits": true/false,
    "duration": true/false,
    "lines": true/false,
    "warn_200k": true/false,
    "cost": true/false,
    "api_duration": true/false,
    "tokens": true/false,
    "worktree": true/false,
    "agent": true/false,
    "vim_mode": true/false
  },
  "colors": true/false,
  "bar_width": N,
  "separator": " \u2502 ",
  "language": "xx"
}
```

## Step 5: Configure statusLine in settings.json

Check if `~/.claude/settings.json` already has a `statusLine` entry.
- If not present: add it pointing to `bash ~/.claude/statusline/run.sh`
- If already present and pointing to the same script: leave it
- If pointing to something else: ask the user if they want to replace it

The statusLine config should be:
```json
"statusLine": {
  "type": "command",
  "command": "bash /home/jeonguk/.claude/statusline/run.sh"
}
```

## Step 6: Preview & Confirm

After applying, show a text preview of what the status line will look like with their selections. Use a sample scenario with realistic values. Tell them to restart Claude Code to see it live.

## Rules

- Be concise and friendly
- Use the detected language consistently throughout
- Use AskUserQuestion for each interactive step
- Do NOT explain implementation details unless asked
- The script at `~/.claude/statusline/run.sh` already exists - do not modify it, only configure it
