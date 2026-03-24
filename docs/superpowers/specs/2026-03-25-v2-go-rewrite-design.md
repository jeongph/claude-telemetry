# claude-telemetry v2 디자인 스펙

## 포지셔닝

**"The status line you can trust"** — 정확하고, 가볍고, 안 깨진다.

ccstatusline(5.8k stars)의 핵심 약점(성능, 정확도, 안정성, i18n 부재)을 구조적으로 해결하여 차별화한다.

| 약속 | 공략하는 ccstatusline 약점 |
|------|--------------------------|
| 정확하다 | 컨텍스트 % 오류 반복 (#251), 동적 색상 없음 (#38, #174) |
| 가볍다 | Node.js 프로세스 매번 생성, CPU 90%+ 스파이크 (#22) |
| 안 깨진다 | JSONL 파싱 의존으로 CC 업데이트 시 자주 깨짐 (#65, #93, #140) |
| 누구나 쓴다 | 영어 하드코딩, 프리셋 테마 없음 (#226, #227), 프로젝트별 설정 없음 (#58) |

## 설계 원칙

1. **정확성 우선** — 부정확한 데이터보다 미표시가 낫다
2. **경량성** — 단일 바이너리, 런타임 의존성 제로, sub-10ms 렌더링
3. **안정성** — stdin JSON(공식 API) + git 명령만 사용. 비공식 소스 배제
4. **점진적 공개** — 기본값은 심플, 파워유저를 위한 확장성 확보
5. **다국어 네이티브** — i18n은 1등 시민
6. **침묵보다 소통** — 에러/로딩 상태도 시각적으로 피드백

## 아키텍처

### 데이터 흐름

```
Claude Code → stdin (JSON) → Go 바이너리 → stdout (ANSI 텍스트)
```

단방향, 무상태. 매 호출마다 stdin을 읽고, 렌더링하고, 종료. 프로세스 상주 없음.

### 데이터 소스 원칙

| 소스 | 사용 여부 | 이유 |
|------|----------|------|
| stdin JSON (공식 API) | 사용 | 공식, 안정, 문서화됨 |
| git 명령 | 사용 | stdin에 없는 정보, git 자체는 안정 |
| settings.json | 제한적 사용 | 언어 감지, effort level 읽기만 (사용자 공개 설정 파일) |
| transcript JSONL | 사용 안 함 | 비공식 내부 포맷, CC 업데이트 시 변경 가능 |
| Anthropic API 직접 호출 | 사용 안 함 | rate limit 유발, 네트워크 의존 |

### stdin JSON 필드 (공식 스펙)

**항상 존재:**
- `cwd`, `session_id`, `transcript_path`, `version`
- `model.id`, `model.display_name` (string 또는 object — 둘 다 처리)
- `workspace.current_dir`, `workspace.project_dir`
- `cost.total_cost_usd`, `cost.total_duration_ms`, `cost.total_api_duration_ms`, `cost.total_lines_added`, `cost.total_lines_removed`
- `context_window.total_input_tokens`, `context_window.total_output_tokens`, `context_window.context_window_size`, `context_window.used_percentage`, `context_window.remaining_percentage`
- `exceeds_200k_tokens`, `output_style.name`

**조건부 존재:**
- `context_window.current_usage.*` — 첫 API 호출 전 null
- `rate_limits.five_hour/seven_day` — OAuth 사용자만, 각 window 독립 부재 가능
- `vim.mode` — vim 모드 활성 시만
- `agent.name` — 에이전트 실행 시만
- `worktree.*` — worktree 세션 시만 (`branch`, `original_branch` 부재 가능)

**stdin에 없는 필드 (의도적 미지원):**
- `thinking_effort` — settings.json에서 읽기로 대체 (max 감지 불가 한계 있음)
- `session_name`, `skills`, `fast_mode`, `mcp_servers`, `permissions` — 비공식 소스 필요

### 모듈 구조

```
claude-telemetry/
├── cmd/
│   └── claude-telemetry/
│       └── main.go              # 진입점: stdin 읽기 → 렌더 → stdout
├── internal/
│   ├── config/
│   │   ├── config.go            # 설정 로딩 (글로벌 + 프로젝트)
│   │   └── preset.go            # 프리셋 정의 (compact/normal/detailed)
│   ├── input/
│   │   └── parser.go            # stdin JSON 파싱, 검증, model string/object 처리
│   ├── section/
│   │   ├── section.go           # Section 인터페이스 정의
│   │   ├── model.go             # 모델명 + effort level
│   │   ├── elapsed.go           # 경과시간
│   │   ├── git.go               # Git 정보 (병렬 실행 + 캐싱)
│   │   ├── context.go           # 컨텍스트 잔량
│   │   ├── ratelimit.go         # Rate limit (5h/7d)
│   │   ├── cost.go              # 세션 비용
│   │   ├── agent.go             # 에이전트명
│   │   ├── vim.go               # Vim 모드
│   │   ├── lines.go             # 코드 변경량
│   │   ├── tokens.go            # 토큰 수
│   │   └── apiduration.go       # API 대기시간
│   ├── render/
│   │   ├── renderer.go          # 라인 조립, 적응형 너비, 우선순위 기반 섹션 드롭
│   │   └── ansi.go              # ANSI 색상, 프로그레스바, CJK 너비 계산
│   ├── git/
│   │   └── git.go               # git 명령 병렬 실행, 파일 기반 캐싱
│   └── i18n/
│       └── i18n.go              # 다국어 레이블 맵 (en/ko/ja/zh)
├── .claude-plugin/
│   ├── plugin.json
│   └── marketplace.json
├── commands/
│   ├── setup.md                 # 바이너리 다운로드 + 설정 (개편)
│   └── remove.md                # 바이너리 + 설정 제거 (개편)
├── scripts/
│   └── run.sh                   # thin wrapper: Go 바이너리 호출
├── config.example.json
└── go.mod
```

### Section 인터페이스

```go
type Section interface {
    Name() string
    Priority() int
    Render(input Input, cfg Config, locale Locale) string
    Width(input Input, cfg Config, locale Locale) int
}
```

모든 섹션이 동일한 인터페이스를 구현. 새 섹션 추가 = 파일 하나 추가.

## 프리셋 모드 & 설정 체계

### 3단계 설정 계층

```
프리셋 (기본값) → 글로벌 설정 (사용자 커스텀) → 프로젝트 설정 (오버라이드)
```

### 프리셋

| 모드 | 라인 수 | 표시 항목 | 대상 |
|------|---------|----------|------|
| compact | 1줄 | Model · Context% · Rate limit% | "정보는 최소한으로" |
| normal (기본) | 2줄 | Model, Elapsed, Git, Context, Rate limits/Cost | 대부분의 사용자 |
| detailed | 3줄 | 전체 섹션 활성화 | 파워유저 |

### 프리셋 출력 예시

compact (1줄):
```
Opus 💭high │ ◆ 72% │ 5h 88% · 7d 65%
```

normal (2줄):
```
Opus 💭high │ ◷ 경과 12m 34s │ myproject:main ↑1 +5/-2
◆ 컨텍스트 ▰▰▰▱▱ 72% (200k) │ 2h 12m/5h ▰▰▰▰▱ 88%  4d 3h/7d ▰▰▰▱▱ 65%
```

detailed (3줄):
```
Opus 💭high │ ◷ 경과 12m 34s │ myproject:main ↑1 +5/-2 ?3 ≡1
◆ 컨텍스트 ▰▰▰▱▱ 72% (200k) │ 2h 12m/5h ▰▰▰▰▱ 88%  4d 3h/7d ▰▰▰▱▱ 65% │ +156/-23 │ 입력 15k 출력 4k
▶ security-reviewer │ NORMAL
```

### 설정 파일

글로벌 (`~/.claude/statusline/config.json`):
```json
{
  "preset": "normal",
  "language": "auto",
  "colors": true,
  "bar_width": 5,
  "separator": " │ ",
  "user_type": "auto",
  "sections": {},
  "thresholds": {
    "context_warn": 50,
    "context_danger": 80,
    "cost_warn": 1.0,
    "cost_danger": 5.0
  }
}
```

프로젝트별 (프로젝트 루트 `.claude-statusline.json`):
```json
{
  "preset": "detailed",
  "sections": {
    "tokens": true
  }
}
```

글로벌 설정에 shallow merge. 기존 config.json 형식도 호환 (preset 없으면 sections 기반 폴백).

## 시각 체계

### 동적 색상 — 3단계 신호

색상은 장식이 아니라 상태 신호. 숫자를 읽기 전에 색상만으로 상황을 파악.

잔량 기준 (Context, Rate limits):

| 잔량 | 색상 | 의미 |
|------|------|------|
| > 50% | 초록 | 여유 있음 |
| 21~50% | 노랑 | 주의 필요 |
| ≤ 20% | 빨강 | 위험 |

Cost (증가할수록 위험):

| 금액 | 색상 |
|------|------|
| < warn ($1) | 초록 |
| warn~danger ($1~5) | 노랑 |
| ≥ danger ($5) | 빨강 |

모든 임계값은 `thresholds`에서 커스터마이징 가능. ccstatusline은 하드코딩.

### 고정 색상

| 요소 | 색상 | 이유 |
|------|------|------|
| Model 이름 | cyan bold | 시각적 앵커 |
| Git branch | magenta | git 컨벤션 |
| Agent 이름 | magenta | "현재 컨텍스트" 범주 |
| 단위/레이블 | dim gray | 수치와 시각 분리 |
| 구분자 | dim gray | 배경으로 물러남 |

### NO_COLOR 지원

`NO_COLOR` 환경변수 설정 시 모든 ANSI 코드 제거. progress bar는 `▰▱` 문자 자체로 잔량 표현.

## 성능

### 목표

| 지표 | 목표 |
|------|------|
| 전체 렌더링 (캐시 hit) | < 10ms |
| git 포함 (캐시 miss) | < 50ms |
| 바이너리 크기 | < 5MB |
| 메모리 | < 10MB |

### git 캐싱

```
호출 → 캐시 파일 확인 → TTL(5초) 이내? → 캐시 사용
                                    ↓ 만료
                               git 명령 6개 goroutine 병렬 실행 → 캐시 갱신
```

- 캐시 위치: `~/.claude/statusline/cache/<repo-hash>.json`
- TTL: 5초
- 모든 git 명령에 `--no-optional-locks` 적용

## Graceful Degradation

"침묵보다 소통" — 3단계:

| 단계 | 상황 | 표시 |
|------|------|------|
| 대기 | 데이터 미도착 | dim + `···` (로딩 암시) |
| 부분 실패 | 특정 필드 null/실패 | 가용 정보 표시 + 실패 부분 `—` 자리 유지 |
| 전체 실패 | stdin 파싱 실패 | `⚠ statusline: invalid input` |

예시:
```
# rate_limits 로딩 중
5h ▱▱▱▱▱ ···

# git remote 없어 ahead/behind 불가
myproject:main +5/-2 ?3     (가용 정보만 표시, ahead/behind 생략)

# stdin 파싱 실패
⚠ statusline: invalid input
```

## Effort Level 표시

settings.json에서 `effortLevel` 읽기:

| 상황 | 표시 | 정확성 |
|------|------|--------|
| `/effort low~high` | `💭low`, `💭high` 등 | 정확 |
| `/effort max` (세션 한정) | 이전 값 또는 `💭auto` | max 감지 불가 |
| 설정 없음 | `💭auto` | 모델 기본값 |

한계: `/effort max`는 settings.json에 persist되지 않아 감지 불가. stdin에 추가되면 자동 전환 예정.

## 배포

### GitHub Releases

태그 푸시 → GitHub Actions → 4개 플랫폼 크로스 컴파일 → Release 에셋 업로드

| OS | 아키텍처 | 파일명 |
|------|---------|--------|
| Linux | amd64 | `claude-telemetry-linux-amd64` |
| Linux | arm64 | `claude-telemetry-linux-arm64` |
| macOS | amd64 | `claude-telemetry-darwin-amd64` |
| macOS | arm64 | `claude-telemetry-darwin-arm64` |

### 플러그인 통합

marketplace 설치로 레포 파일(commands, scripts, config) 수령 → `/claude-telemetry:setup`으로 바이너리 다운로드 + 설정.

run.sh (thin wrapper):
```bash
#!/bin/bash
BIN="${HOME}/.claude/statusline/bin/claude-telemetry"
[ -x "$BIN" ] && exec "$BIN" || echo "⚠ Run /claude-telemetry:setup to install"
```

### 기존 사용자 호환

- 기존 config.json 형식 그대로 동작
- jq 의존성 제거됨
- run.sh 경로 변경 없음, settings.json 수정 불필요
- `/claude-telemetry:setup` 재실행으로 업그레이드

## i18n

지원 언어: en, ko, ja, zh

```go
var locales = map[string]map[string]string{
    "en": {"context": "Context", "elapsed": "Elapsed", "cost": "Cost", ...},
    "ko": {"context": "컨텍스트", "elapsed": "경과", "cost": "비용", ...},
    "ja": {"context": "コンテキスト", "elapsed": "経過", "cost": "費用", ...},
    "zh": {"context": "上下文", "elapsed": "已用", "cost": "费用", ...},
}
```

언어 감지 우선순위: config.json `language` → `~/.claude/settings.json` `language` → `en` 폴백

## ccstatusline 대비 차별점 요약

| | ccstatusline | claude-telemetry v2 |
|---|---|---|
| 런타임 | Node.js 프로세스 매번 생성 | Go 단일 바이너리 |
| 데이터 소스 | stdin + JSONL 파싱 + API 호출 | stdin + git + settings.json(제한적) |
| 정확성 | 컨텍스트 % 오류 반복 | 타입 안전 파싱, 검증 로직 |
| 다국어 | 영어만 | 4개 언어 |
| 설정 | TUI 필수 (React/Ink) | 프리셋 + JSON 편집 |
| 의존성 | Node.js 14+ | 없음 |
| 프리셋 모드 | 없음 (#227) | compact/normal/detailed |
| 프로젝트별 설정 | 없음 (#58) | .claude-statusline.json |
| 동적 색상 임계값 | 없음 (#38, #174) | thresholds 설정 |
| NO_COLOR | 미지원 | 지원 |
| 에러 표시 | `[Timeout]` | 3단계 graceful degradation |

## 의도적 미지원

| 기능 | 이유 |
|------|------|
| thinking_effort (stdin) | 공식 제공 안 됨 — settings.json 읽기로 대체 |
| session_name | transcript JSONL 의존 = 안정성 위반 |
| Anthropic API 직접 호출 | rate limit 유발, 네트워크 의존 |
| Powerline 구분자/테마 | Nerd Font 의존성 = 범용성 저하 |
| Custom Command 위젯 | 임의 shell 실행 = 성능/보안 예측 불가 |
| TUI 설정 화면 | React/Ink 의존성 = 경량성 위반 |
