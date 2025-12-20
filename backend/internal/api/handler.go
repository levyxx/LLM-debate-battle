package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/levyxx/LLM-debate-battle/backend/internal/auth"
	"github.com/levyxx/LLM-debate-battle/backend/internal/db"
	"github.com/levyxx/LLM-debate-battle/backend/internal/debatesvc"
	"github.com/levyxx/LLM-debate-battle/backend/internal/models"
)

type Handlers struct {
	database      *db.DB
	debateService *debatesvc.Service
	tokenStore    *auth.TokenStore
}

func NewHandlers(database *db.DB, debateService *debatesvc.Service, tokenStore *auth.TokenStore) *Handlers {
	return &Handlers{
		database:      database,
		debateService: debateService,
		tokenStore:    tokenStore,
	}
}

func (h *Handlers) Register(r chi.Router) {
	// 認証不要のエンドポイント
	r.Post("/api/auth/register", h.RegisterUser)
	r.Post("/api/auth/login", h.Login)

	// 認証が必要なエンドポイント
	r.Group(func(r chi.Router) {
		r.Use(h.AuthMiddleware)

		r.Post("/api/auth/logout", h.Logout)
		r.Get("/api/auth/me", h.GetCurrentUser)

		r.Post("/api/debate/create", h.CreateDebate)
		r.Post("/api/debate/message", h.SendMessage)
		r.Post("/api/debate/end", h.EndDebate)
		r.Post("/api/debate/llm-step", h.LLMDebateStep)
		r.Get("/api/debate/{id}", h.GetDebate)
		r.Get("/api/debate/{id}/messages", h.GetDebateMessages)

		r.Get("/api/user/stats", h.GetUserStats)
		r.Get("/api/user/history", h.GetUserHistory)
	})

	// トピック生成は認証なしでも可能
	r.Post("/api/debate/generate-topic", h.GenerateTopic)
}

// 認証ミドルウェア
func (h *Handlers) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// "Bearer " プレフィックスを削除
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		userID, err := h.tokenStore.ValidateToken(token)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// ユーザーIDをコンテキストに追加
		ctx := r.Context()
		ctx = setUserID(ctx, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ユーザー登録
func (h *Handlers) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		http.Error(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	if len(req.Password) < 6 {
		http.Error(w, "Password must be at least 6 characters", http.StatusBadRequest)
		return
	}

	// パスワードをハッシュ化
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		log.Printf("Failed to hash password: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// ユーザーを作成
	user, err := h.database.CreateUser(req.Username, hashedPassword)
	if err != nil {
		log.Printf("Failed to create user: %v", err)
		http.Error(w, "Username already exists", http.StatusConflict)
		return
	}

	respondJSON(w, http.StatusCreated, user)
}

// ログイン
func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.database.GetUserByUsername(req.Username)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if !auth.CheckPassword(req.Password, user.PasswordHash) {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := h.tokenStore.CreateToken(user.ID)
	if err != nil {
		log.Printf("Failed to create token: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, models.LoginResponse{
		Token: token,
		User:  *user,
	})
}

// ログアウト
func (h *Handlers) Logout(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}
	h.tokenStore.DeleteToken(token)
	w.WriteHeader(http.StatusOK)
}

// 現在のユーザー情報を取得
func (h *Handlers) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r.Context())
	user, err := h.database.GetUserByID(userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	respondJSON(w, http.StatusOK, user)
}

// トピック生成
func (h *Handlers) GenerateTopic(w http.ResponseWriter, r *http.Request) {
	topic, err := h.debateService.GenerateRandomTopic(r.Context())
	if err != nil {
		log.Printf("Failed to generate topic: %v", err)
		http.Error(w, "Failed to generate topic", http.StatusInternalServerError)
		return
	}
	respondJSON(w, http.StatusOK, topic)
}

// ディベート作成
func (h *Handlers) CreateDebate(w http.ResponseWriter, r *http.Request) {
	var req models.CreateDebateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userID := getUserID(r.Context())
	var userIDPtr *int64
	if req.Mode == "user_vs_llm" {
		userIDPtr = &userID
	}

	session, topicInfo, err := h.debateService.CreateDebateSession(r.Context(), userIDPtr, &req)
	if err != nil {
		log.Printf("Failed to create debate: %v", err)
		http.Error(w, "Failed to create debate", http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusCreated, models.CreateDebateResponse{
		Session:   *session,
		TopicInfo: topicInfo,
	})
}

// メッセージ送信
func (h *Handlers) SendMessage(w http.ResponseWriter, r *http.Request) {
	var req models.SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userMsg, llmMsg, err := h.debateService.ProcessUserMessage(r.Context(), req.SessionID, req.Content)
	if err != nil {
		log.Printf("Failed to process message: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, models.SendMessageResponse{
		UserMessage: userMsg,
		LLMMessage:  *llmMsg,
	})
}

// LLM同士のディベートを1ステップ進める
func (h *Handlers) LLMDebateStep(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SessionID int64 `json:"session_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	llm1Msg, llm2Msg, isFinished, err := h.debateService.ProcessLLMDebateStep(r.Context(), req.SessionID)
	if err != nil {
		log.Printf("Failed to process LLM debate step: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, models.LLMDebateStepResponse{
		LLM1Message: llm1Msg,
		LLM2Message: llm2Msg,
		IsFinished:  isFinished,
	})
}

// ディベート終了
func (h *Handlers) EndDebate(w http.ResponseWriter, r *http.Request) {
	var req models.EndDebateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	session, judgeResult, err := h.debateService.EndDebate(r.Context(), req.SessionID)
	if err != nil {
		log.Printf("Failed to end debate: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, models.EndDebateResponse{
		Session:     *session,
		JudgeResult: *judgeResult,
	})
}

// ディベート詳細取得
func (h *Handlers) GetDebate(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid debate ID", http.StatusBadRequest)
		return
	}

	session, messages, err := h.debateService.GetDebateDetail(id)
	if err != nil {
		http.Error(w, "Debate not found", http.StatusNotFound)
		return
	}

	respondJSON(w, http.StatusOK, models.DebateHistoryResponse{
		Session:  *session,
		Messages: messages,
	})
}

// ディベートメッセージ取得
func (h *Handlers) GetDebateMessages(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid debate ID", http.StatusBadRequest)
		return
	}

	_, messages, err := h.debateService.GetDebateDetail(id)
	if err != nil {
		http.Error(w, "Debate not found", http.StatusNotFound)
		return
	}

	respondJSON(w, http.StatusOK, messages)
}

// ユーザー統計取得
func (h *Handlers) GetUserStats(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r.Context())
	stats, err := h.debateService.GetUserStats(userID)
	if err != nil {
		log.Printf("Failed to get user stats: %v", err)
		http.Error(w, "Failed to get stats", http.StatusInternalServerError)
		return
	}
	respondJSON(w, http.StatusOK, stats)
}

// ユーザー履歴取得
func (h *Handlers) GetUserHistory(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r.Context())
	history, err := h.debateService.GetUserDebateHistory(userID)
	if err != nil {
		log.Printf("Failed to get user history: %v", err)
		http.Error(w, "Failed to get history", http.StatusInternalServerError)
		return
	}
	respondJSON(w, http.StatusOK, history)
}

// ヘルパー関数
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
