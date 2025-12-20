package debatesvc

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/levyxx/LLM-debate-battle/backend/internal/db"
	"github.com/levyxx/LLM-debate-battle/backend/internal/models"
	"github.com/levyxx/LLM-debate-battle/backend/internal/openai"
)

type Service struct {
	database *db.DB
	client   *openai.Client
}

func NewService(database *db.DB, client *openai.Client) *Service {
	return &Service{
		database: database,
		client:   client,
	}
}

// ランダムなディベートテーマを生成
func (s *Service) GenerateRandomTopic(ctx context.Context) (*models.DebateTopicResponse, error) {
	messages := []openai.Message{
		{
			Role: "system",
			Content: `あなたはディベートのテーマを提案するアシスタントです。
興味深く、議論の余地があり、両方の立場から論じることができるディベートテーマを提案してください。
テーマは具体的で、一般の人でも議論に参加できるものにしてください。
政治、社会、技術、倫理、教育など様々な分野からテーマを選んでください。`,
		},
		{
			Role:    "user",
			Content: "新しいディベートテーマを1つ提案してください。",
		},
	}

	response, err := s.client.ChatCompletionWithSchema(ctx, messages, "debate_topic", openai.DebateTopicSchema)
	if err != nil {
		return nil, err
	}

	var topic models.DebateTopicResponse
	if err := json.Unmarshal([]byte(response), &topic); err != nil {
		return nil, fmt.Errorf("failed to parse topic response: %w", err)
	}

	return &topic, nil
}

// ディベートセッションを作成
func (s *Service) CreateDebateSession(ctx context.Context, userID *int64, req *models.CreateDebateRequest) (*models.DebateSession, *models.DebateTopicResponse, error) {
	var topic string
	var topicInfo *models.DebateTopicResponse

	// テーマの決定
	if req.RandomizeTopic || req.Topic == "" {
		generatedTopic, err := s.GenerateRandomTopic(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to generate topic: %w", err)
		}
		topic = generatedTopic.Topic
		topicInfo = generatedTopic
	} else {
		topic = req.Topic
	}

	// ポジションの決定
	var userPosition string
	if req.Mode == "user_vs_llm" {
		userPosition = req.UserPosition
		if req.RandomizePosition || userPosition == "" || userPosition == "random" {
			rand.Seed(time.Now().UnixNano())
			if rand.Intn(2) == 0 {
				userPosition = "pro"
			} else {
				userPosition = "con"
			}
		}
	}

	// データベースに保存
	session, err := s.database.CreateDebateSession(userID, req.Mode, topic, userPosition)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create session: %w", err)
	}

	// LLMポジションをセッションに設定（データベースには保存しないがレスポンスに含める）
	if req.Mode == "user_vs_llm" {
		session.UserPosition = userPosition
		if userPosition == "pro" {
			session.LLMPosition = "con"
		} else {
			session.LLMPosition = "pro"
		}
	} else if req.Mode == "llm_vs_llm" {
		session.LLM1Position = "pro"
		session.LLM2Position = "con"
	}

	// システムメッセージを保存
	systemContent := fmt.Sprintf("ディベートテーマ: %s\n", topic)
	if topicInfo != nil {
		systemContent += fmt.Sprintf("賛成側の立場: %s\n反対側の立場: %s\n背景: %s",
			topicInfo.ProPosition, topicInfo.ConPosition, topicInfo.Background)
	}

	_, err = s.database.CreateMessage(session.ID, "system", systemContent)
	if err != nil {
		log.Printf("Failed to save system message: %v", err)
	}

	return session, topicInfo, nil
}

// ユーザーのメッセージに対してLLMが応答
func (s *Service) ProcessUserMessage(ctx context.Context, sessionID int64, userContent string) (*models.DebateMessage, *models.DebateMessage, error) {
	session, err := s.database.GetDebateSession(sessionID)
	if err != nil {
		return nil, nil, fmt.Errorf("session not found: %w", err)
	}

	if session.Status != "active" && session.Status != "ongoing" {
		return nil, nil, fmt.Errorf("debate has already ended")
	}

	// ユーザーメッセージを保存
	userMsg, err := s.database.CreateMessage(sessionID, "user", userContent)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to save user message: %w", err)
	}

	// 過去のメッセージを取得
	messages, err := s.database.GetSessionMessages(sessionID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get messages: %w", err)
	}

	// LLM用のメッセージを構築
	llmMessages := s.buildLLMMessages(session, messages, "llm")

	// LLMの応答を生成
	response, err := s.client.ChatCompletion(ctx, llmMessages)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get LLM response: %w", err)
	}

	// LLMメッセージを保存
	llmMsg, err := s.database.CreateMessage(sessionID, "llm", response)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to save LLM message: %w", err)
	}

	return userMsg, llmMsg, nil
}

// LLM同士のディベートを1ステップ進める
func (s *Service) ProcessLLMDebateStep(ctx context.Context, sessionID int64) (*models.DebateMessage, *models.DebateMessage, bool, error) {
	session, err := s.database.GetDebateSession(sessionID)
	if err != nil {
		return nil, nil, false, fmt.Errorf("session not found: %w", err)
	}

	if session.Status != "active" && session.Status != "ongoing" {
		return nil, nil, true, nil
	}

	messages, err := s.database.GetSessionMessages(sessionID)
	if err != nil {
		return nil, nil, false, fmt.Errorf("failed to get messages: %w", err)
	}

	// 議論の回数をチェック（最大5往復）
	llm1Count := 0
	llm2Count := 0
	for _, msg := range messages {
		if msg.Role == "llm1" {
			llm1Count++
		} else if msg.Role == "llm2" {
			llm2Count++
		}
	}

	if llm1Count >= 5 && llm2Count >= 5 {
		return nil, nil, true, nil
	}

	// LLM vs LLMの場合のポジション設定
	session.LLM1Position = "pro"
	session.LLM2Position = "con"

	// 1回の呼び出しで1つのLLMの応答のみを返す
	// LLM1の番（LLM1のカウントがLLM2以下の場合）
	if llm1Count <= llm2Count {
		llm1Messages := s.buildLLMMessages(session, messages, "llm1")
		response, err := s.client.ChatCompletion(ctx, llm1Messages)
		if err != nil {
			return nil, nil, false, fmt.Errorf("failed to get LLM1 response: %w", err)
		}

		llm1Msg, err := s.database.CreateMessage(sessionID, "llm1", response)
		if err != nil {
			return nil, nil, false, fmt.Errorf("failed to save LLM1 message: %w", err)
		}

		// 終了判定（次のステップで終わるかどうか）
		isFinished := llm1Count >= 4 && llm2Count >= 5

		return llm1Msg, nil, isFinished, nil
	}

	// LLM2の番
	llm2Messages := s.buildLLMMessages(session, messages, "llm2")
	response, err := s.client.ChatCompletion(ctx, llm2Messages)
	if err != nil {
		return nil, nil, false, fmt.Errorf("failed to get LLM2 response: %w", err)
	}

	llm2Msg, err := s.database.CreateMessage(sessionID, "llm2", response)
	if err != nil {
		return nil, nil, false, fmt.Errorf("failed to save LLM2 message: %w", err)
	}

	// 終了判定
	isFinished := llm1Count >= 5 && llm2Count >= 4

	return nil, llm2Msg, isFinished, nil
}

// ディベートを終了して審査
func (s *Service) EndDebate(ctx context.Context, sessionID int64) (*models.DebateSession, *models.JudgeResponse, error) {
	session, err := s.database.GetDebateSession(sessionID)
	if err != nil {
		return nil, nil, fmt.Errorf("session not found: %w", err)
	}

	if session.Status != "active" && session.Status != "ongoing" {
		return nil, nil, fmt.Errorf("debate has already ended")
	}

	messages, err := s.database.GetSessionMessages(sessionID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get messages: %w", err)
	}

	// LLM vs LLMの場合のポジション設定
	if session.Mode == "llm_vs_llm" {
		session.LLM1Position = "pro"
		session.LLM2Position = "con"
	} else {
		// ユーザーポジションからLLMポジションを設定
		if session.UserPosition == "pro" {
			session.LLMPosition = "con"
		} else {
			session.LLMPosition = "pro"
		}
	}

	// 審査用のメッセージを構築
	judgeMessages := s.buildJudgeMessages(session, messages)

	// 審査結果を取得
	response, err := s.client.ChatCompletionWithSchema(ctx, judgeMessages, "judge_result", openai.JudgeResultSchema)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get judge response: %w", err)
	}

	var judgeResult models.JudgeResponse
	if err := json.Unmarshal([]byte(response), &judgeResult); err != nil {
		return nil, nil, fmt.Errorf("failed to parse judge response: %w", err)
	}

	// 勝者を決定
	var winner string
	if session.Mode == "user_vs_llm" {
		if judgeResult.Winner == session.UserPosition {
			winner = "user"
		} else if judgeResult.Winner == session.LLMPosition {
			winner = "llm"
		} else {
			winner = "draw"
		}

		// ユーザー統計を更新
		if session.UserID != nil {
			stats, err := s.database.GetUserStats(*session.UserID)
			if err != nil {
				log.Printf("Failed to get user stats: %v", err)
			} else {
				stats.TotalDebates++
				if winner == "user" {
					stats.Wins++
				} else if winner == "llm" {
					stats.Losses++
				} else {
					stats.Draws++
				}
				if err := s.database.UpdateUserStats(stats); err != nil {
					log.Printf("Failed to update user stats: %v", err)
				}
			}
		}
	} else {
		if judgeResult.Winner == "pro" {
			winner = "llm1"
		} else if judgeResult.Winner == "con" {
			winner = "llm2"
		} else {
			winner = "draw"
		}
	}

	// セッションを更新
	now := time.Now()
	session.Status = "finished"
	session.Winner = &winner
	session.JudgeComment = &judgeResult.FinalComment
	session.FinishedAt = &now

	if err := s.database.UpdateDebateSession(session); err != nil {
		return nil, nil, fmt.Errorf("failed to update session: %w", err)
	}

	// 審査結果を保存
	judgeContent, _ := json.Marshal(judgeResult)
	_, err = s.database.CreateMessage(sessionID, "judge", string(judgeContent))
	if err != nil {
		log.Printf("Failed to save judge message: %v", err)
	}

	return session, &judgeResult, nil
}

// LLM用のメッセージを構築
func (s *Service) buildLLMMessages(session *models.DebateSession, messages []models.DebateMessage, role string) []openai.Message {
	var position string
	if role == "llm" {
		position = session.LLMPosition
	} else if role == "llm1" {
		position = session.LLM1Position
	} else {
		position = session.LLM2Position
	}

	positionDesc := "賛成"
	if position == "con" {
		positionDesc = "反対"
	}

	systemPrompt := fmt.Sprintf(`あなたはディベートの参加者です。
テーマ: %s
あなたの立場: %s側

以下のルールに従ってディベートを行ってください：
1. 自分の立場を論理的に主張してください
2. 相手の主張に対して適切に反論してください
3. 具体的な例やデータを用いて説得力のある議論をしてください
4. 礼儀正しく、建設的な議論を心がけてください
5. 回答は300文字程度にまとめてください`, session.Topic, positionDesc)

	llmMessages := []openai.Message{
		{Role: "system", Content: systemPrompt},
	}

	for _, msg := range messages {
		if msg.Role == "system" {
			continue
		}

		msgRole := "user"
		if msg.Role == role {
			msgRole = "assistant"
		}

		llmMessages = append(llmMessages, openai.Message{
			Role:    msgRole,
			Content: msg.Content,
		})
	}

	return llmMessages
}

// 審査用のメッセージを構築
func (s *Service) buildJudgeMessages(session *models.DebateSession, messages []models.DebateMessage) []openai.Message {
	systemPrompt := fmt.Sprintf(`あなたは公平なディベートの審査員です。
以下のディベートを評価し、勝者を決定してください。

テーマ: %s
賛成側(pro): 賛成の立場
反対側(con): 反対の立場

評価基準：
1. 論理性：主張の論理的整合性
2. 説得力：具体的な根拠やデータの使用
3. 反論力：相手の主張への効果的な反論
4. 表現力：わかりやすく説得力のある表現

公平に両者を評価し、結果を出してください。`, session.Topic)

	debateContent := "【ディベートの内容】\n\n"
	for _, msg := range messages {
		if msg.Role == "system" || msg.Role == "judge" {
			continue
		}

		var speaker string
		switch msg.Role {
		case "user":
			if session.UserPosition == "pro" {
				speaker = "賛成側(ユーザー)"
			} else {
				speaker = "反対側(ユーザー)"
			}
		case "llm":
			if session.LLMPosition == "pro" {
				speaker = "賛成側(AI)"
			} else {
				speaker = "反対側(AI)"
			}
		case "llm1":
			speaker = "賛成側(AI-1)"
		case "llm2":
			speaker = "反対側(AI-2)"
		}

		debateContent += fmt.Sprintf("%s:\n%s\n\n", speaker, msg.Content)
	}

	return []openai.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: debateContent + "\n上記のディベートを評価してください。"},
	}
}

// ユーザーの統計を取得
func (s *Service) GetUserStats(userID int64) (*models.UserStats, error) {
	return s.database.GetUserStats(userID)
}

// ユーザーのディベート履歴を取得
func (s *Service) GetUserDebateHistory(userID int64) ([]models.DebateSession, error) {
	return s.database.GetUserDebateHistory(userID)
}

// ディベートの詳細を取得
func (s *Service) GetDebateDetail(sessionID int64) (*models.DebateSession, []models.DebateMessage, error) {
	session, err := s.database.GetDebateSession(sessionID)
	if err != nil {
		return nil, nil, err
	}

	messages, err := s.database.GetSessionMessages(sessionID)
	if err != nil {
		return nil, nil, err
	}

	return session, messages, nil
}
