package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hkdf"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"sync"
	"time"

	"github.com/authnas/authnas/go-server/internal/config"
	"github.com/authnas/authnas/go-server/internal/model"
	"github.com/authnas/authnas/go-server/internal/repository"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

type TOTPService struct {
	cfg        *config.Config
	totpRepo   *repository.TOTPRepository
	userRepo   *repository.UserRepository
	encryptKey []byte
	stopChan   chan struct{}
}

func (s *TOTPService) initEncryptKey() {
	if s.encryptKey != nil {
		return
	}
	key, err := hkdf.Key(sha256.New, []byte(s.cfg.Security.StorageKey), nil, "authnas-totp-key", 32)
	if err != nil {
		hash := sha256.Sum256([]byte(s.cfg.Security.StorageKey))
		s.encryptKey = hash[:]
		return
	}
	s.encryptKey = key
}

func (s *TOTPService) encryptSecret(plaintext string) (string, error) {
	s.initEncryptKey()
	block, err := aes.NewCipher(s.encryptKey)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (s *TOTPService) decryptSecret(encryptedText string) (string, error) {
	s.initEncryptKey()
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedText)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(s.encryptKey)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("ciphertext too short")
	}
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

type totpAttempt struct {
	failedAttempts int
	lastAttempt    time.Time
	lockedUntil    time.Time
}

var (
	totpAttempts   = make(map[string]*totpAttempt)
	totpAttemptsMu sync.RWMutex
)

const (
	MaxTOTPAttempts     = 5
	TOTPLockoutDuration = 5 * time.Minute
	TOTPAttemptWindow   = 15 * time.Minute
)

func NewTOTPService(cfg *config.Config, totpRepo *repository.TOTPRepository, userRepo *repository.UserRepository) *TOTPService {
	s := &TOTPService{
		cfg:      cfg,
		totpRepo: totpRepo,
		userRepo: userRepo,
		stopChan: make(chan struct{}),
	}
	go s.cleanupAttempts()
	return s
}

func (s *TOTPService) cleanupAttempts() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.performCleanup()
		case <-s.stopChan:
			return
		}
	}
}

func (s *TOTPService) performCleanup() {
	totpAttemptsMu.Lock()
	defer totpAttemptsMu.Unlock()
	now := time.Now()
	for userID, attempt := range totpAttempts {
		if now.Sub(attempt.lastAttempt) > TOTPAttemptWindow*2 {
			delete(totpAttempts, userID)
		}
	}
}

func (s *TOTPService) Stop() {
	close(s.stopChan)
}

func (s *TOTPService) isLockedOut(userID string) bool {
	totpAttemptsMu.RLock()
	defer totpAttemptsMu.RUnlock()

	attempt, exists := totpAttempts[userID]
	if !exists {
		return false
	}

	if attempt.lockedUntil.After(time.Now()) {
		return true
	}

	if time.Since(attempt.lastAttempt) > TOTPAttemptWindow && attempt.failedAttempts < MaxTOTPAttempts {
		return false
	}

	return attempt.failedAttempts >= MaxTOTPAttempts
}

func (s *TOTPService) recordFailedAttempt(userID string) {
	totpAttemptsMu.Lock()
	defer totpAttemptsMu.Unlock()

	attempt, exists := totpAttempts[userID]
	now := time.Now()

	if !exists {
		totpAttempts[userID] = &totpAttempt{
			failedAttempts: 1,
			lastAttempt:    now,
		}
		return
	}

	if now.Sub(attempt.lastAttempt) > TOTPAttemptWindow {
		attempt.failedAttempts = 1
		attempt.lastAttempt = now
		attempt.lockedUntil = time.Time{}
		return
	}

	attempt.failedAttempts++
	attempt.lastAttempt = now

	if attempt.failedAttempts >= MaxTOTPAttempts {
		attempt.lockedUntil = now.Add(TOTPLockoutDuration)
	}
}

func (s *TOTPService) clearAttempts(userID string) {
	totpAttemptsMu.Lock()
	defer totpAttemptsMu.Unlock()
	delete(totpAttempts, userID)
}

func (s *TOTPService) IsLockedOut(userID string) bool {
	return s.isLockedOut(userID)
}

type TOTPResult struct {
	Secret    string `json:"secret"`
	QRCodeURI string `json:"qr_code_uri"`
}

func (s *TOTPService) Generate(user *model.User) (*TOTPResult, error) {
	if err := s.totpRepo.DeleteByUserID(user.ID); err != nil {
		return nil, err
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      s.cfg.App.Name,
		AccountName: user.Username,
		Period:      30,
		Algorithm:   otp.AlgorithmSHA1,
		Digits:      otp.DigitsSix,
	})
	if err != nil {
		return nil, err
	}

	encryptedSecret, err := s.encryptSecret(key.Secret())
	if err != nil {
		return nil, err
	}

	totpModel := &model.TOTP{
		ID:        generateID(),
		UserID:    user.ID,
		Secret:    encryptedSecret,
		Issuer:    s.cfg.App.Name,
		CreatedAt: now(),
		UpdatedAt: now(),
	}

	if err := s.totpRepo.Create(totpModel); err != nil {
		return nil, err
	}

	return &TOTPResult{
		Secret:    key.Secret(),
		QRCodeURI: key.URL(),
	}, nil
}

func (s *TOTPService) Validate(userID, token string) bool {
	return s.ValidateWithWindow(userID, token, 1)
}

func (s *TOTPService) ValidateWithWindow(userID, token string, window int) bool {
	if s.isLockedOut(userID) {
		return false
	}

	totpModel, err := s.totpRepo.GetByUserID(userID)
	if err != nil {
		return false
	}

	secret, err := s.decryptSecret(totpModel.Secret)
	if err != nil {
		return false
	}

	if window <= 0 {
		window = 1
	}

	valid, err := totp.ValidateCustom(token, secret, time.Now(), totp.ValidateOpts{
		Period:    30,
		Skew:      uint(window),
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	})

	if err != nil {
		return false
	}

	if !valid {
		s.recordFailedAttempt(userID)
		return false
	}

	s.clearAttempts(userID)
	return true
}

func (s *TOTPService) Delete(userID string) error {
	return s.totpRepo.DeleteByUserID(userID)
}

func (s *TOTPService) GetTOTPByUserID(userID string) (*model.TOTP, error) {
	return s.totpRepo.GetByUserID(userID)
}

func (s *TOTPService) HasTOTP(userID string) bool {
	totpModel, err := s.totpRepo.GetByUserID(userID)
	if err != nil {
		return false
	}
	return totpModel != nil
}

func (s *TOTPService) UpdateTOTP(totpModel *model.TOTP) error {
	totpModel.UpdatedAt = time.Now()
	return s.totpRepo.Update(totpModel)
}

func (s *TOTPService) GenerateTOTPCode(secret string) (string, error) {
	return totp.GenerateCode(secret, time.Now())
}
