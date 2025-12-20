package db

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/levyxx/LLM-debate-battle/backend/internal/models"
)

type DB struct {
	conn *sql.DB
}

func NewDB(dbPath string) (*DB, error) {
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	db := &DB{conn: conn}
	if err := db.migrate(); err != nil {
		return nil, err
	}

	return db, nil
}

func (d *DB) Close() error {
	return d.conn.Close()
}

func (d *DB) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS debate_sessions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
		mode TEXT NOT NULL,
		topic TEXT NOT NULL,
		user_position TEXT,
		status TEXT DEFAULT 'active',
		winner TEXT,
		judge_comment TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		ended_at DATETIME,
		FOREIGN KEY (user_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS debate_messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		session_id INTEGER NOT NULL,
		role TEXT NOT NULL,
		content TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (session_id) REFERENCES debate_sessions(id)
	);

	CREATE TABLE IF NOT EXISTS user_stats (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER UNIQUE NOT NULL,
		total_debates INTEGER DEFAULT 0,
		wins INTEGER DEFAULT 0,
		losses INTEGER DEFAULT 0,
		draws INTEGER DEFAULT 0,
		FOREIGN KEY (user_id) REFERENCES users(id)
	);
	`

	_, err := d.conn.Exec(schema)
	return err
}

// ユーザー作成
func (d *DB) CreateUser(username, passwordHash string) (*models.User, error) {
	result, err := d.conn.Exec(
		"INSERT INTO users (username, password_hash) VALUES (?, ?)",
		username, passwordHash,
	)
	if err != nil {
		return nil, err
	}

	id, _ := result.LastInsertId()

	// ユーザー統計も初期化
	_, err = d.conn.Exec(
		"INSERT INTO user_stats (user_id) VALUES (?)",
		id,
	)
	if err != nil {
		log.Printf("Warning: failed to create user stats: %v", err)
	}

	return &models.User{
		ID:        id,
		Username:  username,
		CreatedAt: time.Now(),
	}, nil
}

// ユーザー名でユーザー取得
func (d *DB) GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	err := d.conn.QueryRow(
		"SELECT id, username, password_hash, created_at FROM users WHERE username = ?",
		username,
	).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// IDでユーザー取得
func (d *DB) GetUserByID(id int64) (*models.User, error) {
	var user models.User
	err := d.conn.QueryRow(
		"SELECT id, username, password_hash, created_at FROM users WHERE id = ?",
		id,
	).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// ディベートセッション作成
func (d *DB) CreateDebateSession(userID *int64, mode, topic, userPosition string) (*models.DebateSession, error) {
	result, err := d.conn.Exec(
		"INSERT INTO debate_sessions (user_id, mode, topic, user_position) VALUES (?, ?, ?, ?)",
		userID, mode, topic, userPosition,
	)
	if err != nil {
		return nil, err
	}

	id, _ := result.LastInsertId()
	return d.GetDebateSession(id)
}

// ディベートセッション取得
func (d *DB) GetDebateSession(id int64) (*models.DebateSession, error) {
	var session models.DebateSession
	var userID sql.NullInt64
	var userPosition sql.NullString
	var winner sql.NullString
	var judgeComment sql.NullString
	var finishedAt sql.NullTime

	err := d.conn.QueryRow(
		`SELECT id, user_id, mode, topic, user_position, status, winner, judge_comment, created_at, ended_at 
		FROM debate_sessions WHERE id = ?`,
		id,
	).Scan(&session.ID, &userID, &session.Mode, &session.Topic, &userPosition,
		&session.Status, &winner, &judgeComment, &session.CreatedAt, &finishedAt)
	if err != nil {
		return nil, err
	}

	if userID.Valid {
		session.UserID = &userID.Int64
	}
	if userPosition.Valid {
		session.UserPosition = userPosition.String
	}
	if winner.Valid {
		session.Winner = &winner.String
	}
	if judgeComment.Valid {
		session.JudgeComment = &judgeComment.String
	}
	if finishedAt.Valid {
		session.FinishedAt = &finishedAt.Time
	}

	return &session, nil
}

// ディベートセッション更新
func (d *DB) UpdateDebateSession(session *models.DebateSession) error {
	var finishedAt interface{}
	if session.FinishedAt != nil {
		finishedAt = *session.FinishedAt
	}

	_, err := d.conn.Exec(
		`UPDATE debate_sessions SET status = ?, winner = ?, judge_comment = ?, ended_at = ? WHERE id = ?`,
		session.Status, session.Winner, session.JudgeComment, finishedAt, session.ID,
	)
	return err
}

// メッセージ作成
func (d *DB) CreateMessage(sessionID int64, role, content string) (*models.DebateMessage, error) {
	result, err := d.conn.Exec(
		"INSERT INTO debate_messages (session_id, role, content) VALUES (?, ?, ?)",
		sessionID, role, content,
	)
	if err != nil {
		return nil, err
	}

	id, _ := result.LastInsertId()
	return &models.DebateMessage{
		ID:        id,
		SessionID: sessionID,
		Role:      role,
		Content:   content,
		CreatedAt: time.Now(),
	}, nil
}

// セッションのメッセージ取得
func (d *DB) GetSessionMessages(sessionID int64) ([]models.DebateMessage, error) {
	rows, err := d.conn.Query(
		"SELECT id, session_id, role, content, created_at FROM debate_messages WHERE session_id = ? ORDER BY created_at ASC",
		sessionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []models.DebateMessage
	for rows.Next() {
		var msg models.DebateMessage
		if err := rows.Scan(&msg.ID, &msg.SessionID, &msg.Role, &msg.Content, &msg.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}
	return messages, nil
}

// ユーザー統計取得
func (d *DB) GetUserStats(userID int64) (*models.UserStats, error) {
	var stats models.UserStats
	var id int64
	err := d.conn.QueryRow(
		"SELECT id, user_id, total_debates, wins, losses, draws FROM user_stats WHERE user_id = ?",
		userID,
	).Scan(&id, &stats.UserID, &stats.TotalDebates, &stats.Wins, &stats.Losses, &stats.Draws)
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

// ユーザー統計更新
func (d *DB) UpdateUserStats(stats *models.UserStats) error {
	_, err := d.conn.Exec(
		`UPDATE user_stats SET total_debates = ?, wins = ?, losses = ?, draws = ? WHERE user_id = ?`,
		stats.TotalDebates, stats.Wins, stats.Losses, stats.Draws, stats.UserID,
	)
	return err
}

// ユーザーのディベート履歴取得
func (d *DB) GetUserDebateHistory(userID int64) ([]models.DebateSession, error) {
	rows, err := d.conn.Query(
		`SELECT id, user_id, mode, topic, user_position, status, winner, judge_comment, created_at, ended_at 
		FROM debate_sessions WHERE user_id = ? ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []models.DebateSession
	for rows.Next() {
		var session models.DebateSession
		var userIDVal sql.NullInt64
		var userPosition sql.NullString
		var winner sql.NullString
		var judgeComment sql.NullString
		var finishedAt sql.NullTime

		if err := rows.Scan(&session.ID, &userIDVal, &session.Mode, &session.Topic, &userPosition,
			&session.Status, &winner, &judgeComment, &session.CreatedAt, &finishedAt); err != nil {
			return nil, err
		}

		if userIDVal.Valid {
			session.UserID = &userIDVal.Int64
		}
		if userPosition.Valid {
			session.UserPosition = userPosition.String
		}
		if winner.Valid {
			session.Winner = &winner.String
		}
		if judgeComment.Valid {
			session.JudgeComment = &judgeComment.String
		}
		if finishedAt.Valid {
			session.FinishedAt = &finishedAt.Time
		}

		sessions = append(sessions, session)
	}
	return sessions, nil
}
