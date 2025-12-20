package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/levyxx/LLM-debate-battle/backend/internal/api"
	"github.com/levyxx/LLM-debate-battle/backend/internal/auth"
	"github.com/levyxx/LLM-debate-battle/backend/internal/db"
	"github.com/levyxx/LLM-debate-battle/backend/internal/debatesvc"
	"github.com/levyxx/LLM-debate-battle/backend/internal/openai"
)

func main() {
	// 環境変数から設定を読み込み
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	model := os.Getenv("OPENAI_MODEL")
	if model == "" {
		model = "gpt-4o-mini"
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./debate.db"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// データベース初期化
	database, err := db.NewDB(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// OpenAIクライアント初期化
	openaiClient := openai.NewClient(apiKey, model)

	// サービス初期化
	debateService := debatesvc.NewService(database, openaiClient)
	tokenStore := auth.NewTokenStore()

	// ハンドラー初期化
	handlers := api.NewHandlers(database, debateService, tokenStore)

	// ルーター設定
	r := chi.NewRouter()

	// ミドルウェア
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://localhost:3000", "*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// ルート登録
	handlers.Register(r)

	// ヘルスチェック
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
