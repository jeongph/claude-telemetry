package input

import (
	"encoding/json"
	"errors"
)

// Input은 Claude Code 상태 라인에서 stdin으로 전달되는 JSON 데이터를 나타낸다.
type Input struct {
	CWD            string          `json:"cwd"`
	SessionID      string          `json:"session_id"`
	TranscriptPath string          `json:"transcript_path"`
	Version        string          `json:"version"`
	Model          ModelInfo       `json:"-"`        // RawModel에서 파싱
	RawModel       json.RawMessage `json:"model"`
	Workspace      *Workspace      `json:"workspace"`
	OutputStyle    *OutputStyle    `json:"output_style"`
	Cost           Cost            `json:"cost"`
	ContextWindow  ContextWindow   `json:"context_window"`
	Exceeds200K    bool            `json:"exceeds_200k_tokens"`
	RateLimits     *RateLimits     `json:"rate_limits"`
	Vim            *Vim            `json:"vim"`
	Agent          *Agent          `json:"agent"`
	Worktree       *Worktree       `json:"worktree"`
}

// ModelInfo는 모델 식별 정보를 나타낸다.
type ModelInfo struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}

// Workspace는 현재 작업 중인 워크스페이스 정보를 나타낸다.
type Workspace struct {
	CurrentDir string `json:"current_dir"`
	ProjectDir string `json:"project_dir"`
}

// OutputStyle은 출력 스타일 설정을 나타낸다.
type OutputStyle struct {
	Name string `json:"name"`
}

// Cost는 API 사용 비용 및 통계를 나타낸다.
type Cost struct {
	TotalCostUSD       float64 `json:"total_cost_usd"`
	TotalDurationMS    float64 `json:"total_duration_ms"`
	TotalAPIDurationMS float64 `json:"total_api_duration_ms"`
	TotalLinesAdded    int     `json:"total_lines_added"`
	TotalLinesRemoved  int     `json:"total_lines_removed"`
}

// ContextWindow는 컨텍스트 윈도우 사용 현황을 나타낸다.
type ContextWindow struct {
	TotalInputTokens  int           `json:"total_input_tokens"`
	TotalOutputTokens int           `json:"total_output_tokens"`
	ContextWindowSize int           `json:"context_window_size"`
	UsedPercentage    *float64      `json:"used_percentage"`
	RemainingPct      *float64      `json:"remaining_percentage"`
	CurrentUsage      *CurrentUsage `json:"current_usage"`
}

// CurrentUsage는 현재 요청의 토큰 사용량 세부 정보를 나타낸다.
type CurrentUsage struct {
	InputTokens              int `json:"input_tokens"`
	OutputTokens             int `json:"output_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens"`
}

// RateLimits는 API 사용 제한 현황을 나타낸다.
type RateLimits struct {
	FiveHour *RateWindow `json:"five_hour"`
	SevenDay *RateWindow `json:"seven_day"`
}

// RateWindow는 특정 기간의 사용량 제한 정보를 나타낸다.
type RateWindow struct {
	UsedPercentage float64 `json:"used_percentage"`
	ResetsAt       float64 `json:"resets_at"`
}

// Vim은 Vim 모드 정보를 나타낸다.
type Vim struct {
	Mode string `json:"mode"`
}

// Agent는 현재 활성화된 에이전트 정보를 나타낸다.
type Agent struct {
	Name string `json:"name"`
}

// Worktree는 Git worktree 정보를 나타낸다.
type Worktree struct {
	Name           string `json:"name"`
	Path           string `json:"path"`
	Branch         string `json:"branch"`
	OriginalCWD    string `json:"original_cwd"`
	OriginalBranch string `json:"original_branch"`
}

// Parse는 바이트 슬라이스를 Input 구조체로 파싱한다.
// model 필드는 오브젝트("id", "display_name") 또는 문자열 형태 모두 허용한다.
// 문자열인 경우 ID와 DisplayName 모두 해당 문자열로 설정된다.
func Parse(data []byte) (*Input, error) {
	if len(data) == 0 {
		return nil, errors.New("input: 빈 데이터")
	}

	var inp Input
	if err := json.Unmarshal(data, &inp); err != nil {
		return nil, err
	}

	// model 필드를 오브젝트 또는 문자열로 파싱
	if len(inp.RawModel) > 0 {
		var modelObj ModelInfo
		if err := json.Unmarshal(inp.RawModel, &modelObj); err == nil && modelObj.ID != "" {
			// 오브젝트 형태 파싱 성공
			inp.Model = modelObj
		} else {
			// 문자열 형태로 파싱 시도
			var modelStr string
			if err := json.Unmarshal(inp.RawModel, &modelStr); err == nil {
				inp.Model = ModelInfo{
					ID:          modelStr,
					DisplayName: modelStr,
				}
			}
		}
	}

	return &inp, nil
}
