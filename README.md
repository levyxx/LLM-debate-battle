# 🎯 LLM Debate Battle

AIとディベートバトルを楽しめるWebアプリケーション

## ✨ 機能

### 🤖 ディベートモード
- **ユーザー vs LLM**: AIと1対1の真剣勝負
  - 賛成/反対の立場を選択（またはランダム）
  - リアルタイムでAIの応答を表示
  - 第三者AI審査員による公平な判定

- **LLM vs LLM**: AI同士のバトルを観戦
  - 2つのAIが自動で議論
  - 各発言が即座に表示
  - 最大5往復の熱い攻防

### 🎲 トピック生成
- **AIによる自動生成**: OpenAI APIで興味深いテーマを自動作成
- **手動入力**: 自分で好きなテーマを設定可能
- **多様なジャンル**: 政治、社会、技術、倫理など幅広いトピック

### 📊 審査・統計機能
- **詳細な審査結果**: 
  - 勝者判定（賛成側/反対側/引き分け）
  - スコア表示
  - 判定理由
  - 各陣営の強み・弱みの分析
  - 総評コメント
  
- **戦績管理**:
  - 総ディベート数
  - 勝利/敗北/引き分け数
  - 勝率の自動計算
  - 過去のディベート履歴

## 🏗️ 技術スタック

### バックエンド
- **言語**: Go 1.23
- **ルーター**: chi v5
- **AI API**: OpenAI Go SDK (構造化出力対応)
- **データベース**: SQLite3
- **認証**: トークンベース認証
- **セキュリティ**: bcryptによるパスワードハッシュ化

### フロントエンド
- **フレームワーク**: React 18
- **言語**: TypeScript 5
- **ビルドツール**: Vite 7
- **ルーティング**: React Router DOM
- **HTTP クライアント**: Axios
- **スタイリング**: カスタムCSS（ダークテーマ）

### インフラ
- **コンテナ**: Docker & Docker Compose
- **Webサーバー**: Nginx (リバースプロキシ)
- **データ永続化**: Docker Volume

## 🚀 クイックスタート

### ローカル開発環境

#### バックエンドのセットアップ

1. **ディレクトリに移動**
   ```bash
   cd backend
   ```

2. **環境変数を設定**
   ```bash
   cp .env.example .env
   nano .env
   ```
   
   以下を設定：
   ```env
   OPENAI_API_KEY=sk-your-actual-api-key-here
   OPENAI_MODEL=gpt-4o-mini
   PORT=8080
   DB_PATH=./debate.db
   ```

3. **依存関係をインストール**
   ```bash
   go mod download
   ```

4. **アプリケーションをビルド**
   ```bash
   go build ./cmd/server
   ```

5. **サーバーを起動**
   ```bash
   ./server
   ```
   
   サーバーは http://localhost:8080 で起動します。

#### フロントエンドのセットアップ

1. **新しいターミナルを開き、ディレクトリに移動**
   ```bash
   cd frontend
   ```

2. **依存関係をインストール**
   ```bash
   npm install
   ```

3. **開発サーバーを起動**
   ```bash
   npm run dev
   ```
   
   開発サーバーは http://localhost:5173 で起動します。

4. **本番ビルド（オプション）**
   ```bash
   npm run build
   ```

## 📖 使い方

### 1. アカウント作成
- 「新規登録」をクリック
- ユーザー名とパスワードを入力
- 登録完了後、自動的にログイン

### 2. ディベート作成
1. ホーム画面から「ユーザー vs LLM」または「LLM vs LLM」を選択
2. トピックを選択：
   - 「ランダム生成」でAIにテーマを作ってもらう
   - 自分でテーマを入力
3. 立場を選択（ユーザー vs LLMの場合）：
   - 賛成側
   - 反対側  
   - ランダム

### 3. ディベートを楽しむ
- **ユーザー vs LLM**:
  - テキストエリアに主張を入力
  - 「送信」ボタンでAIに返答
  - 満足したら「ディベートを終了して審査」

- **LLM vs LLM**:
  - 「ディベートを開始」ボタンをクリック
  - AI同士が自動で議論
  - リアルタイムで発言が表示
  - 途中で「終了して審査」も可能

### 4. 審査結果を確認
- 勝者の発表
- 各陣営のスコア
- 判定理由と詳細なフィードバック
- 強み・弱みの分析

### 5. 履歴を確認
- 「履歴」メニューから過去のディベートを閲覧
- 戦績統計を確認
- 過去のディベート詳細を再確認

## 🔧 環境変数

### バックエンド（`backend/.env`）

| 変数名 | 必須 | デフォルト値 | 説明 |
|--------|------|-------------|------|
| `OPENAI_API_KEY` | ✅ | - | OpenAI APIキー |
| `OPENAI_MODEL` | ❌ | `gpt-4o-mini` | 使用するOpenAIモデル |
| `PORT` | ❌ | `8080` | バックエンドサーバーのポート |
| `DB_PATH` | ❌ | `./debate.db` | SQLiteデータベースファイルのパス |

### フロントエンド（`frontend/.env.development`）

| 変数名 | 必須 | デフォルト値 | 説明 |
|--------|------|-------------|------|
| `VITE_API_URL` | ❌ | `http://localhost:8080` | バックエンドAPIのURL |

## 🗄️ データベーススキーマ

### users
- `id`: ユーザーID（主キー）
- `username`: ユーザー名（ユニーク）
- `password_hash`: ハッシュ化されたパスワード
- `created_at`: 作成日時

### debate_sessions
- `id`: セッションID（主キー）
- `user_id`: ユーザーID（外部キー）
- `mode`: ディベートモード（user_vs_llm/llm_vs_llm）
- `topic`: ディベートテーマ
- `user_position`: ユーザーの立場（pro/con）
- `status`: ステータス（active/finished）
- `winner`: 勝者
- `judge_comment`: 審査コメント
- `created_at`: 作成日時
- `ended_at`: 終了日時

### debate_messages
- `id`: メッセージID（主キー）
- `session_id`: セッションID（外部キー）
- `role`: 役割（user/llm/llm1/llm2/judge/system）
- `content`: メッセージ内容
- `created_at`: 作成日時

### user_stats
- `id`: 統計ID（主キー）
- `user_id`: ユーザーID（外部キー、ユニーク）
- `total_debates`: 総ディベート数
- `wins`: 勝利数
- `losses`: 敗北数
- `draws`: 引き分け数

## 🔍 トラブルシューティング

**問題**: ポートが既に使用されている
```bash
# 使用中のポートを確認
sudo lsof -i :3000
sudo lsof -i :8080

# プロセスを終了
sudo kill -9 <PID>
```

### API関連

**問題**: OpenAI APIエラー
- `.env`ファイルでAPIキーが正しく設定されているか確認
- APIキーの有効性を確認
- モデル名が正しいか確認（gpt-4o-mini推奨）

**問題**: CORS エラー
- バックエンドのCORS設定を確認
- フロントエンドのAPI URLが正しいか確認

### データベース関連

**問題**: データベース接続エラー
```bash
# SQLiteファイルのパーミッションを確認
ls -la backend/debate.db

# 必要に応じてパーミッション変更
chmod 666 backend/debate.db
```
