package models

import "time"

// ユーザーモデル
type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

// ディベートセッション
type DebateSession struct {
	ID           int64      `json:"id"`
	UserID       *int64     `json:"user_id,omitempty"`       // ユーザー vs LLM の場合のみ
	Topic        string     `json:"topic"`                   // ディベートのテーマ
	UserPosition string     `json:"user_position,omitempty"` // ユーザーの立場（pro/con）
	LLMPosition  string     `json:"llm_position,omitempty"`  // LLMの立場（pro/con）
	LLM1Position string     `json:"llm1_position,omitempty"` // LLM vs LLM の場合
	LLM2Position string     `json:"llm2_position,omitempty"` // LLM vs LLM の場合
	Mode         string     `json:"mode"`                    // "user_vs_llm" or "llm_vs_llm"
	Status       string     `json:"status"`                  // "ongoing", "finished"
	Winner       *string    `json:"winner,omitempty"`        // "user", "llm", "llm1", "llm2", "draw"
	JudgeComment *string    `json:"judge_comment,omitempty"` // 審査員のコメント
	CreatedAt    time.Time  `json:"created_at"`
	FinishedAt   *time.Time `json:"finished_at,omitempty"`
}

// ディベートメッセージ
type DebateMessage struct {
	ID        int64     `json:"id"`
	SessionID int64     `json:"session_id"`
	Role      string    `json:"role"` // "user", "llm", "llm1", "llm2", "judge", "system"
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// ユーザー統計
type UserStats struct {
	UserID       int64   `json:"user_id"`
	TotalDebates int     `json:"total_debates"`
	Wins         int     `json:"wins"`
	Losses       int     `json:"losses"`
	Draws        int     `json:"draws"`
	WinRate      float64 `json:"win_rate"`
}

// ディベートテーマ生成のレスポンス（構造化出力用）
type DebateTopicResponse struct {
	Topic       string `json:"topic"`
	ProPosition string `json:"pro_position"`
	ConPosition string `json:"con_position"`
	Background  string `json:"background"`
}

// ディベート引数生成のレスポンス（構造化出力用）
type DebateArgumentResponse struct {
	Argument     string   `json:"argument"`
	KeyPoints    []string `json:"key_points"`
	Counterpoint string   `json:"counterpoint,omitempty"`
}

// 審査結果のレスポンス（構造化出力用）
type JudgeResponse struct {
	Winner        string   `json:"winner"` // "pro", "con", "draw"
	Score         Score    `json:"score"`
	Reasoning     string   `json:"reasoning"`
	ProStrengths  []string `json:"pro_strengths"`
	ProWeaknesses []string `json:"pro_weaknesses"`
	ConStrengths  []string `json:"con_strengths"`
	ConWeaknesses []string `json:"con_weaknesses"`
	FinalComment  string   `json:"final_comment"`
}

type Score struct {
	Pro int `json:"pro"`
	Con int `json:"con"`
}

// APIリクエスト/レスポンス型
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type CreateDebateRequest struct {
	Mode              string `json:"mode"`                    // "user_vs_llm" or "llm_vs_llm"
	Topic             string `json:"topic,omitempty"`         // 空の場合はLLMがランダム生成
	UserPosition      string `json:"user_position,omitempty"` // "pro", "con", "random"
	RandomizeTopic    bool   `json:"randomize_topic"`
	RandomizePosition bool   `json:"randomize_position"`
}

type CreateDebateResponse struct {
	Session   DebateSession        `json:"session"`
	TopicInfo *DebateTopicResponse `json:"topic_info,omitempty"`
}

type SendMessageRequest struct {
	SessionID int64  `json:"session_id"`
	Content   string `json:"content"`
}

type SendMessageResponse struct {
	UserMessage *DebateMessage `json:"user_message,omitempty"`
	LLMMessage  DebateMessage  `json:"llm_message"`
}

type EndDebateRequest struct {
	SessionID int64 `json:"session_id"`
}

type EndDebateResponse struct {
	Session     DebateSession `json:"session"`
	JudgeResult JudgeResponse `json:"judge_result"`
}

type DebateHistoryResponse struct {
	Session  DebateSession   `json:"session"`
	Messages []DebateMessage `json:"messages"`
}

type LLMDebateStepResponse struct {
	LLM1Message *DebateMessage `json:"llm1_message,omitempty"`
	LLM2Message *DebateMessage `json:"llm2_message,omitempty"`
	IsFinished  bool           `json:"is_finished"`
}
