package service

import (
	"sync"
	"time"

	"github.com/authnas/authnas/go-server/pkg/utils"
)

type CSRFToken struct {
	Token     string
	UserID    string
	CreatedAt time.Time
	ExpiresAt time.Time
}

type CSRFService struct {
	tokens      map[string]*CSRFToken
	tokensMu    sync.RWMutex
	expiry      time.Duration
	random      *utils.RandomUtil
	stopChan    chan struct{}
	cleanupDone chan struct{}
}

func NewCSRFService(random *utils.RandomUtil, expiryMinutes int) *CSRFService {
	if expiryMinutes <= 0 {
		expiryMinutes = 60
	}
	s := &CSRFService{
		tokens:      make(map[string]*CSRFToken),
		expiry:      time.Duration(expiryMinutes) * time.Minute,
		random:      random,
		stopChan:    make(chan struct{}),
		cleanupDone: make(chan struct{}),
	}
	go s.startCleanupTask()
	return s
}

func (s *CSRFService) startCleanupTask() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.CleanupExpired()
		case <-s.stopChan:
			close(s.cleanupDone)
			return
		}
	}
}

func (s *CSRFService) Stop() {
	close(s.stopChan)
	<-s.cleanupDone
}

func (s *CSRFService) GenerateToken(userID string) (*CSRFToken, error) {
	tokenStr, err := s.random.GenerateToken(32)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	csrfToken := &CSRFToken{
		Token:     tokenStr,
		UserID:    userID,
		CreatedAt: now,
		ExpiresAt: now.Add(s.expiry),
	}

	s.tokensMu.Lock()
	defer s.tokensMu.Unlock()
	s.tokens[tokenStr] = csrfToken

	return csrfToken, nil
}

func (s *CSRFService) ValidateToken(tokenStr string, userID string) bool {
	s.tokensMu.Lock()
	defer s.tokensMu.Unlock()

	csrfToken, exists := s.tokens[tokenStr]
	if !exists {
		return false
	}

	if time.Now().After(csrfToken.ExpiresAt) {
		delete(s.tokens, tokenStr)
		return false
	}

	if userID != "" && csrfToken.UserID != "" && csrfToken.UserID != userID {
		delete(s.tokens, tokenStr)
		return false
	}

	delete(s.tokens, tokenStr)
	return true
}

func (s *CSRFService) ValidateTokenAndRefresh(tokenStr string, userID string) (bool, string) {
	s.tokensMu.Lock()
	defer s.tokensMu.Unlock()

	csrfToken, exists := s.tokens[tokenStr]
	if !exists {
		return false, ""
	}

	if time.Now().After(csrfToken.ExpiresAt) {
		delete(s.tokens, tokenStr)
		return false, ""
	}

	if userID != "" && csrfToken.UserID != "" && csrfToken.UserID != userID {
		delete(s.tokens, tokenStr)
		return false, ""
	}

	delete(s.tokens, tokenStr)

	newTokenStr, err := s.random.GenerateToken(32)
	if err != nil {
		return true, ""
	}

	now := time.Now()
	newCsrfToken := &CSRFToken{
		Token:     newTokenStr,
		UserID:    userID,
		CreatedAt: now,
		ExpiresAt: now.Add(s.expiry),
	}
	s.tokens[newTokenStr] = newCsrfToken

	return true, newTokenStr
}

func (s *CSRFService) InvalidateToken(tokenStr string) {
	s.tokensMu.Lock()
	defer s.tokensMu.Unlock()
	delete(s.tokens, tokenStr)
}

func (s *CSRFService) CleanupExpired() {
	s.tokensMu.Lock()
	defer s.tokensMu.Unlock()

	now := time.Now()
	for token, csrfToken := range s.tokens {
		if now.After(csrfToken.ExpiresAt) {
			delete(s.tokens, token)
		}
	}
}

func (s *CSRFService) GetExpiry() time.Duration {
	return s.expiry
}
