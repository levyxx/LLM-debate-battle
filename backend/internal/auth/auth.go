package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
	ErrTokenExpired       = errors.New("token expired")
	ErrInvalidToken       = errors.New("invalid token")
)

type TokenInfo struct {
	UserID    int64
	ExpiresAt time.Time
}

type TokenStore struct {
	tokens map[string]TokenInfo
	mu     sync.RWMutex
}

func NewTokenStore() *TokenStore {
	return &TokenStore{
		tokens: make(map[string]TokenInfo),
	}
}

func (s *TokenStore) CreateToken(userID int64) (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	token := hex.EncodeToString(bytes)

	s.mu.Lock()
	defer s.mu.Unlock()

	s.tokens[token] = TokenInfo{
		UserID:    userID,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	return token, nil
}

func (s *TokenStore) ValidateToken(token string) (int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	info, exists := s.tokens[token]
	if !exists {
		return 0, ErrInvalidToken
	}

	if time.Now().After(info.ExpiresAt) {
		return 0, ErrTokenExpired
	}

	return info.UserID, nil
}

func (s *TokenStore) DeleteToken(token string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.tokens, token)
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
