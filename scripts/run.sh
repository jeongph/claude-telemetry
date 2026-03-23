#!/bin/bash
# Claude Code custom status line (multi-line layout)
# Line 1: Model │ Elapsed │ Git
# Line 2: Context │ Rate limits (OAuth) or Cost (API key)
# Line 3: Agent │ Vim (only when active)
# Config: ~/.claude/statusline/config.json

CFG_DIR="${CLAUDE_STATUSLINE_CONFIG:-$HOME/.claude/statusline}"
CFG=$(cat "$CFG_DIR/config.json" 2>/dev/null || echo '{"sections":{}}')

# ── Language detection ──
resolve_lang() {
  local cfg_lang
  cfg_lang=$(echo "$CFG" | jq -r '.language // "auto"')
  if [ "$cfg_lang" != "auto" ]; then echo "$cfg_lang"; return; fi
  local cc_lang
  cc_lang=$(jq -r '.language // empty' "$HOME/.claude/settings.json" 2>/dev/null)
  case "$cc_lang" in
    한국어|ko*) echo "ko"; return ;; 日本語|ja*) echo "ja"; return ;;
    中文|zh*) echo "zh"; return ;; *) echo "en" ;;
  esac
}

LANG_CODE=$(resolve_lang)
COLS=$(tput cols < /dev/tty 2>/dev/null) 2>/dev/null
[ -z "$COLS" ] && COLS=$(stty size < /dev/tty 2>/dev/null | awk '{print $2}') 2>/dev/null
[ -z "$COLS" ] && COLS=200

# ── Read stdin, then gather git info ──
INPUT=$(cat)

CWD=$(echo "$INPUT" | jq -r '.cwd // .workspace.current_dir // "."' 2>/dev/null)
[ -z "$CWD" ] && CWD="."

GIT_BRANCH=""
GIT_ADD=0
GIT_DEL=0
GIT_AHEAD=0
GIT_BEHIND=0
if git -C "$CWD" rev-parse --is-inside-work-tree &>/dev/null; then
  GIT_BRANCH=$(git -C "$CWD" rev-parse --abbrev-ref HEAD 2>/dev/null)
  read -r GIT_ADD GIT_DEL <<< "$({ git -C "$CWD" diff --numstat 2>/dev/null; git -C "$CWD" diff --cached --numstat 2>/dev/null; } | awk '{a+=$1; d+=$2} END {print a+0, d+0}')"
  read -r GIT_AHEAD GIT_BEHIND <<< "$(git -C "$CWD" rev-list --left-right --count HEAD...@{u} 2>/dev/null || echo "0 0")"
fi

echo "$INPUT" | jq -r \
  --argjson cfg "$CFG" \
  --arg lang "$LANG_CODE" \
  --argjson cols "$COLS" \
  --arg git_branch "$GIT_BRANCH" \
  --argjson git_add "${GIT_ADD:-0}" \
  --argjson git_del "${GIT_DEL:-0}" \
  --argjson git_ahead "${GIT_AHEAD:-0}" \
  --argjson git_behind "${GIT_BEHIND:-0}" \
'

# ── Config ──
def on(key): ($cfg.sections // {}) | if has(key) then .[key] else true end;
def colors: $cfg.colors // true;
def bw: $cfg.bar_width // 5;

# ── ANSI ──
def c(code): if colors then "\u001b[" + code + "m" else "" end;
def R: c("0");
def D: c("2;37");
def grn: c("32");
def ylw: c("33");
def red: c("31");
def cyn: c("1;36");
def mag: c("35");
def wht: c("37");

def sep: D + ($cfg.separator // " \u2502 ") + R;

# ── i18n ──
def L:
  if $lang == "ko" then
    { "ctx":"\ucee8\ud14d\uc2a4\ud2b8", "dur":"\uacbd\uacfc",
      "api":"API \ub300\uae30", "in":"\uc785\ub825", "out":"\ucd9c\ub825",
      "cost":"\ube44\uc6a9" }
  elif $lang == "ja" then
    { "ctx":"\u30b3\u30f3\u30c6\u30ad\u30b9\u30c8", "dur":"\u7d4c\u904e",
      "api":"API\u5f85\u6a5f", "in":"\u5165\u529b", "out":"\u51fa\u529b",
      "cost":"\u8cbb\u7528" }
  elif $lang == "zh" then
    { "ctx":"\u4e0a\u4e0b\u6587", "dur":"\u5df2\u7528",
      "api":"API\u7b49\u5f85", "in":"\u8f93\u5165", "out":"\u8f93\u51fa",
      "cost":"\u8d39\u7528" }
  else
    { "ctx":"Context", "dur":"Elapsed",
      "api":"API", "in":"In", "out":"Out",
      "cost":"Cost" }
  end;
def l(k): L[k] // k;

# ── Helpers ──
def tc: if . >= 80 then red elif . >= 50 then ylw else grn end;

def bar:
  . as $pct | bw as $w |
  ($pct / 100 * $w + 0.5 | floor | [., 0] | max | [., $w] | min) as $f |
  ($w - $f) as $e |
  ($pct | tc) +
  ([range($f)] | map("\u25b0") | join("")) +
  D + ([range($e)] | map("\u25b1") | join("")) + R;

def fmt_dur:
  (. / 1000 | floor) |
  if . >= 3600 then "\(. / 3600 | floor)h\(. % 3600 / 60 | floor)m"
  elif . >= 60 then "\(. / 60 | floor)m\(. % 60)s"
  else "\(.)s" end;

def fmt_remaining:
  (. - now | floor) |
  if . <= 0 then null
  elif . >= 86400 then "\(. / 86400 | floor)d\(. % 86400 / 3600 | floor)h"
  elif . >= 3600 then "\(. / 3600 | floor)h\(. % 3600 / 60 | floor)m"
  elif . >= 60 then "\(. / 60 | floor)m"
  else "\(.)s" end;

def fmt_cost:
  . * 100 | round | . / 100 | tostring |
  if test("[.]") then split(".") | "\(.[0]).\(.[1] + "00" | .[:2])"
  else . + ".00" end;

def fmt_k:
  if . >= 1000000 then "\(. / 1000000 * 10 | round / 10)M"
  elif . >= 1000 then "\(. / 1000 | floor)k"
  else "\(.)" end;

def dw:
  gsub("\u001b\\[[0-9;]*m"; "") |
  explode | map(
    if . >= 44032 and . <= 55203 then 2
    elif . >= 19968 and . <= 40959 then 2
    elif . >= 12288 and . <= 12351 then 2
    else 1 end
  ) | add // 0;

# ── Detect user type: OAuth (has rate_limits) vs API key ──
. as $d |
(($cfg.user_type // "auto") |
  if . == "oauth" then true
  elif . == "api" then false
  else $d.rate_limits != null
  end
) as $is_oauth |

# ══════════════════════════════════════════
# LINE 1: Model │ Elapsed │ Git
# ══════════════════════════════════════════

([ # Model
   ($d.model.display_name // null) |
   if . then cyn + . + R else empty end,

   # Elapsed
   (if on("duration") then
     ($d.cost.total_duration_ms // null) |
     if . and . > 0 then
       D + "\u25f7 " + l("dur") + " " + R + wht + fmt_dur + R
     else empty end
   else empty end),

   # Folder:branch + sync + diff
   (if on("git") and $git_branch != "" then
     (($d.workspace.project_dir // $d.cwd // null) |
       if . then split("/") | last else null end
     ) as $folder |
     (if $folder then wht + $folder + D + ":" + R else "" end) +
     mag + $git_branch + R +
     (if $git_ahead > 0 or $git_behind > 0 then
       " " +
       (if $git_ahead > 0 then ylw + "\u2191\($git_ahead)" + R else "" end) +
       (if $git_behind > 0 then cyn + "\u2193\($git_behind)" + R else "" end)
     else "" end) +
     (if $git_add > 0 or $git_del > 0 then
       " " + grn + "+\($git_add)" + R + D + "/" + R + red + "-\($git_del)" + R
     else "" end)
   else
     # No git — show folder name only
     (($d.workspace.project_dir // $d.cwd // null) |
       if . then split("/") | last else null end
     ) | if . then wht + . + R else empty end
   end)

 ] | join(sep)
) as $line1 |

# ══════════════════════════════════════════
# LINE 2: Context │ Rate limits / Cost
# ══════════════════════════════════════════

([
  (if on("context") then
    ($d.context_window.used_percentage // 0) as $pct |
    ($d.context_window.context_window_size // null) as $sz |
    ($d.exceeds_200k_tokens == true) as $over |
    {ord:1, pri:1, txt: (
      cyn + "\u25c6 " + D + l("ctx") + " " + R +
      ($pct | bar) + " " +
      ($pct | tc) + "\($pct | round)%" + R +
      (if $sz then
        (if $over then c("1;33") else D end) + " (\($sz | fmt_k))" + R
      else "" end))}
  else empty end),

  # OAuth: show rate limits with reset countdown
  (if $is_oauth and on("rate_limits") then
    [
      ($d.rate_limits.five_hour // null |
        if . then
          (.used_percentage // 0) as $pct |
          (.resets_at // null | if . then fmt_remaining else null end) as $reset |
          D + (if $reset then $reset + "/" else "" end) + "5h " + R +
          ($pct | bar) + " " + ($pct | tc) + "\($pct | round)%" + R
        else empty end),
      ($d.rate_limits.seven_day // null |
        if . then
          (.used_percentage // 0) as $pct |
          (.resets_at // null | if . then fmt_remaining else null end) as $reset |
          D + (if $reset then $reset + "/" else "" end) + "7d " + R +
          ($pct | bar) + " " + ($pct | tc) + "\($pct | round)%" + R
        else empty end)
    ] | if length > 0 then {ord:2, pri:2, txt: (join("  "))} else empty end
  else empty end),

  # API key: auto-show cost (replaces rate limits)
  (if ($is_oauth | not) then
    ($d.cost.total_cost_usd // null) |
    if . and . > 0 then
      {ord:2, pri:2, txt: (
        D + l("cost") + " " + R +
        (if . >= 5 then red elif . >= 1 then ylw else grn end) +
        "$" + fmt_cost + R)}
    else empty end
  else empty end),

  (if on("lines") then
    ($d.cost.total_lines_added // null) as $a |
    ($d.cost.total_lines_removed // null) as $r |
    if $a or $r then
      {ord:3, pri:4, txt: (grn + "+\($a // 0)" + R + D + "/" + R + red + "-\($r // 0)" + R)}
    else empty end
  else empty end),

  (if on("api_duration") then
    ($d.cost.total_api_duration_ms // null) |
    if . and . > 0 then
      {ord:4, pri:8, txt: (D + "\u21bb " + l("api") + " " + R + wht + fmt_dur + R)}
    else empty end
  else empty end),

  (if on("tokens") then
    ($d.context_window.total_input_tokens // null) as $ti |
    ($d.context_window.total_output_tokens // null) as $to |
    if $ti then
      {ord:5, pri:9, txt: (
        D + l("in") + " " + R + wht + ($ti | fmt_k) + R +
        D + " " + l("out") + " " + R + wht + ($to // 0 | fmt_k) + R)}
    else empty end
  else empty end)

] |
  sort_by(.pri) |
  (sep | dw) as $sw |
  reduce .[] as $s (
    {sel: [], w: 0};
    ($s.txt | dw) as $tw |
    (if .sel | length > 0 then $sw else 0 end) as $extra |
    if (.w + $tw + $extra) <= ($cols - 2) then
      .sel += [$s] | .w += ($tw + $extra)
    else . end
  ) |
  .sel | sort_by(.ord) | map(.txt) | join(sep)
) as $line2 |

# ══════════════════════════════════════════
# LINE 3: Agent │ Vim (only when active)
# ══════════════════════════════════════════

([
  (if on("agent") then
    ($d.agent.name // null) |
    if . then D + "\u25b6 " + R + mag + . + R else empty end
  else empty end),

  (if on("vim_mode") then
    ($d.vim.mode // null) |
    if . then
      (if . == "NORMAL" then cyn elif . == "INSERT" then grn else wht end) + . + R
    else empty end
  else empty end)

] | map(select(. != null and . != "")) | join(sep)
) as $line3 |

# ── Output ──
[ $line1,
  (if $line2 != "" then $line2 else empty end),
  (if $line3 != "" then $line3 else empty end)
] | join("\n")
'
