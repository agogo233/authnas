package service

import (
	"crypto/rsa"
	"crypto/sha256"
	"crypto/subtle"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/authnas/authnas/go-server/internal/config"
	"github.com/authnas/authnas/go-server/internal/model"
	"github.com/authnas/authnas/go-server/internal/repository"
	"github.com/authnas/authnas/go-server/pkg/utils"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/hkdf"
	"os"
	"strings"
	"sync"
	"time"
)

type AuthService struct {
	cfg                   *config.Config
	userRepo              *repository.UserRepository
	keyRepo               *repository.KeyRepository
	totpRepo              *repository.TOTPRepository
	passkeyRepo           *repository.PasskeyRepository
	emailVerificationRepo *repository.EmailVerificationRepository
	random                *utils.RandomUtil
	privateKey            *rsa.PrivateKey
	stopCleanupChan       chan struct{}
	cleanupDoneChan       chan struct{}
}

func NewAuthService(
	cfg *config.Config,
	userRepo *repository.UserRepository,
	keyRepo *repository.KeyRepository,
	totpRepo *repository.TOTPRepository,
	passkeyRepo *repository.PasskeyRepository,
	emailVerificationRepo *repository.EmailVerificationRepository,
	random *utils.RandomUtil,
) *AuthService {
	svc := &AuthService{
		cfg:                   cfg,
		userRepo:              userRepo,
		keyRepo:               keyRepo,
		totpRepo:              totpRepo,
		passkeyRepo:           passkeyRepo,
		emailVerificationRepo: emailVerificationRepo,
		random:                random,
		stopCleanupChan:       make(chan struct{}),
		cleanupDoneChan:       make(chan struct{}),
	}

	if cfg.JWT.PrivateKey != "" {
		if pk, err := loadJWTPrivateKey(cfg.JWT.PrivateKey); err == nil {
			svc.privateKey = pk
		}
	}

	go cleanupLoginAttempts(svc.stopCleanupChan, svc.cleanupDoneChan)
	return svc
}

func (s *AuthService) StopCleanup() {
	close(s.stopCleanupChan)
	<-s.cleanupDoneChan
}

func loadJWTPrivateKey(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.New("failed to read private key file: " + err.Error())
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	if block.Type != "RSA PRIVATE KEY" && block.Type != "PRIVATE KEY" {
		return nil, errors.New("unsupported PEM block type: " + block.Type)
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		keyInterface, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, errors.New("failed to parse private key: " + err.Error())
		}
		var ok bool
		key, ok = keyInterface.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("private key is not RSA key")
		}
	}

	return key, nil
}

const (
	argon2Memory      uint32 = 64 * 1024
	argon2Iterations  uint32 = 3
	argon2Parallelism uint8  = 4
	argon2SaltLength         = 16
	argon2KeyLength          = 32
)

func (s *AuthService) HashPassword(password string) (string, error) {
	salt, err := s.random.GenerateRandomBytes(argon2SaltLength)
	if err != nil {
		return "", err
	}
	hash := argon2.IDKey([]byte(password), salt, argon2Iterations, argon2Memory, argon2Parallelism, argon2KeyLength)
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, argon2Memory, argon2Iterations, argon2Parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash[:])), nil
}

func (s *AuthService) VerifyPassword(hashWithSalt, password string) bool {
	parts := strings.Split(hashWithSalt, "$")
	if len(parts) != 6 {
		return false
	}

	if parts[1] != "argon2id" || parts[2] != fmt.Sprintf("v=%d", argon2.Version) {
		return false
	}

	var memory, iterations int
	var parallelism uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism); err != nil {
		return false
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil || len(salt) != argon2SaltLength {
		return false
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil || len(expectedHash) != argon2KeyLength {
		return false
	}

	actualHash := argon2.IDKey([]byte(password), salt, uint32(iterations), uint32(memory), parallelism, argon2KeyLength)
	return subtle.ConstantTimeCompare(expectedHash, actualHash) == 1
}

type loginAttempt struct {
	failedAttempts int
	lastAttempt    time.Time
	lockedUntil    time.Time
}

var (
	loginAttempts               = make(map[string]*loginAttempt)
	loginAttemptsMu             sync.RWMutex
	maxLoginAttempts            = 5
	loginLockoutDuration        = 15 * time.Minute
	loginAttemptWindow          = 15 * time.Minute
	loginAttemptCleanupInterval = 5 * time.Minute
	maxLoginAttemptsMapSize     = 100000
)

func loginAttemptKey(userID, ip string) string {
	return userID + ":" + ip
}

func cleanupLoginAttempts(stop <-chan struct{}, done chan<- struct{}) {
	ticker := time.NewTicker(loginAttemptCleanupInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			loginAttemptsMu.Lock()
			now := time.Now()
			cutoff := now.Add(-loginAttemptWindow * 2)
			for key, attempt := range loginAttempts {
				if attempt.lastAttempt.Before(cutoff) && attempt.lockedUntil.Before(now) {
					delete(loginAttempts, key)
				}
			}
			if len(loginAttempts) > maxLoginAttemptsMapSize {
				oldestCount := len(loginAttempts) / 2
				for key, attempt := range loginAttempts {
					if oldestCount <= 0 {
						break
					}
					if attempt.lockedUntil.Before(now) && attempt.failedAttempts < maxLoginAttempts {
						delete(loginAttempts, key)
						oldestCount--
					}
				}
			}
			loginAttemptsMu.Unlock()
		case <-stop:
			close(done)
			return
		}
	}
}

func ResetLoginAttempts() {
	loginAttemptsMu.Lock()
	defer loginAttemptsMu.Unlock()
	loginAttempts = make(map[string]*loginAttempt)
}

func (s *AuthService) ResetLoginAttemptsMap() {
	ResetLoginAttempts()
}

func (s *AuthService) IsLoginLockedOut(userID string, ip string) bool {
	loginAttemptsMu.RLock()
	defer loginAttemptsMu.RUnlock()

	key := loginAttemptKey(userID, ip)
	attempt, exists := loginAttempts[key]
	if !exists {
		return false
	}

	if attempt.lockedUntil.After(time.Now()) {
		return true
	}

	if time.Since(attempt.lastAttempt) > loginAttemptWindow {
		return false
	}

	return attempt.failedAttempts >= maxLoginAttempts
}

func (s *AuthService) RecordFailedLogin(userID string, ip string) {
	loginAttemptsMu.Lock()
	defer loginAttemptsMu.Unlock()

	key := loginAttemptKey(userID, ip)
	attempt, exists := loginAttempts[key]
	if !exists {
		attempt = &loginAttempt{}
		loginAttempts[key] = attempt
	}

	attempt.failedAttempts++
	attempt.lastAttempt = time.Now()

	if attempt.failedAttempts >= maxLoginAttempts {
		attempt.lockedUntil = time.Now().Add(loginLockoutDuration)
	}
}

func (s *AuthService) RecordSuccessfulLogin(userID string, ip string) {
	loginAttemptsMu.Lock()
	defer loginAttemptsMu.Unlock()

	key := loginAttemptKey(userID, ip)
	delete(loginAttempts, key)
}

func (s *AuthService) GetLoginLockoutRemainingTime(userID string, ip string) time.Duration {
	loginAttemptsMu.RLock()
	defer loginAttemptsMu.RUnlock()

	key := loginAttemptKey(userID, ip)
	attempt, exists := loginAttempts[key]
	if !exists || attempt.lockedUntil.IsZero() {
		return 0
	}

	remaining := attempt.lockedUntil.Sub(time.Now())
	if remaining < 0 {
		return 0
	}
	return remaining
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

type Claims struct {
	UserID       string
	Username     string
	Email        string
	TokenVersion int
	IsAdmin      bool
}

func (s *AuthService) GenerateTokenPair(user *model.User, userAgent string) (*TokenPair, error) {
	accessExpiry := s.cfg.JWT.GetAccessTokenExpiry()
	refreshExpiry := s.cfg.JWT.GetRefreshTokenExpiry()

	accessClaims := jwt.MapClaims{
		"sub":           user.ID,
		"iss":           s.cfg.OIDC.Issuer,
		"user_id":       user.ID,
		"username":      user.Username,
		"email":         user.Email,
		"token_version": user.TokenVersion,
		"is_admin":      user.IsAdmin,
		"exp":           time.Now().Add(accessExpiry).Unix(),
		"iat":           time.Now().Unix(),
	}

	var accessTokenString string
	var err error
	if s.privateKey != nil {
		accessToken := jwt.NewWithClaims(jwt.SigningMethodRS256, accessClaims)
		accessTokenString, err = accessToken.SignedString(s.privateKey)
		if err != nil {
			return nil, err
		}
	} else {
		accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
		accessTokenString, err = accessToken.SignedString([]byte(s.cfg.Security.StorageKey))
		if err != nil {
			return nil, err
		}
	}

	refreshToken, err := s.random.GenerateToken(32)
	if err != nil {
		return nil, err
	}

	hashedRefreshToken, err := bcrypt.GenerateFromPassword([]byte(refreshToken), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	h := sha256.Sum256([]byte(refreshToken))
	lookupHash := base64.RawStdEncoding.EncodeToString(h[:])

	key := &model.Key{
		ID:                     generateID(),
		UserID:                 user.ID,
		TokenVersion:           user.TokenVersion,
		RefreshTokenHash:       string(hashedRefreshToken),
		RefreshTokenLookupHash: lookupHash,
		ExpiresAt:              time.Now().Add(refreshExpiry),
		CreatedAt:              time.Now(),
		UserAgent:              userAgent,
	}
	if err := s.keyRepo.Create(key); err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) GenerateMFAToken(userID string) (string, error) {
	mfaExpiry := 5 * time.Minute

	user, err := s.userRepo.GetByID(userID)
	if err != nil || user == nil {
		return "", errors.New("user not found")
	}

	mfaClaims := jwt.MapClaims{
		"sub":           userID,
		"iss":           s.cfg.OIDC.Issuer,
		"aud":           "authnas-mfa",
		"type":          "mfa",
		"jti":           generateID(),
		"token_version": user.TokenVersion,
		"exp":           time.Now().Add(mfaExpiry).Unix(),
		"iat":           time.Now().Unix(),
	}

	var mfaTokenString string
	if s.privateKey != nil {
		mfaToken := jwt.NewWithClaims(jwt.SigningMethodRS256, mfaClaims)
		mfaTokenString, err = mfaToken.SignedString(s.privateKey)
	} else {
		derivedKey := deriveUserKey(s.cfg.Security.StorageKey, userID, user.TokenVersion)
		mfaToken := jwt.NewWithClaims(jwt.SigningMethodHS256, mfaClaims)
		mfaTokenString, err = mfaToken.SignedString(derivedKey)
	}
	if err != nil {
		return "", err
	}

	return mfaTokenString, nil
}

func deriveUserKey(masterKey, userID string, tokenVersion int) []byte {
	info := []byte("authnas-mfa-" + userID)
	hkdfReader := hkdf.New(sha256.New, []byte(masterKey), []byte{}, info)
	derived := make([]byte, 32)
	hkdfReader.Read(derived)
	return derived
}

func (s *AuthService) ValidateMFAToken(mfaTokenString string) (string, error) {
	token, err := jwt.Parse(mfaTokenString, func(token *jwt.Token) (interface{}, error) {
		if s.privateKey != nil {
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return &s.privateKey.PublicKey, nil
		}

		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return nil, jwt.ErrSignatureInvalid
		}

		userID, ok := claims["sub"].(string)
		if !ok || userID == "" {
			return nil, jwt.ErrSignatureInvalid
		}

		user, err := s.userRepo.GetByID(userID)
		if err != nil || user == nil {
			return nil, jwt.ErrSignatureInvalid
		}

		return deriveUserKey(s.cfg.Security.StorageKey, userID, user.TokenVersion), nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		tokenType, _ := claims["type"].(string)
		if tokenType != "mfa" {
			return "", jwt.ErrSignatureInvalid
		}

		userID, _ := claims["sub"].(string)
		if userID == "" {
			return "", jwt.ErrSignatureInvalid
		}

		if s.privateKey == nil {
			tokenVersion, _ := claims["token_version"].(float64)
			user, err := s.userRepo.GetByID(userID)
			if err != nil || user == nil {
				return "", jwt.ErrSignatureInvalid
			}
			if int(tokenVersion) != user.TokenVersion {
				return "", jwt.ErrSignatureInvalid
			}
		}

		return userID, nil
	}

	return "", jwt.ErrSignatureInvalid
}

func (s *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); ok {
			return []byte(s.cfg.Security.StorageKey), nil
		}
		if _, ok := token.Method.(*jwt.SigningMethodRSA); ok {
			if s.privateKey != nil {
				return &s.privateKey.PublicKey, nil
			}
			return nil, errors.New("RSA public key not available")
		}
		return nil, jwt.ErrSignatureInvalid
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID, _ := claims["user_id"].(string)
		username, _ := claims["username"].(string)
		email, _ := claims["email"].(string)
		tokenVersion, _ := claims["token_version"].(float64)
		isAdmin, _ := claims["is_admin"].(bool)

		user, err := s.userRepo.GetByID(userID)
		if err != nil || user == nil {
			return nil, jwt.ErrSignatureInvalid
		}

		if int(tokenVersion) != user.TokenVersion {
			return nil, jwt.ErrSignatureInvalid
		}

		return &Claims{
			UserID:       userID,
			Username:     username,
			Email:        email,
			TokenVersion: int(tokenVersion),
			IsAdmin:      isAdmin,
		}, nil
	}

	return nil, jwt.ErrSignatureInvalid
}

func (s *AuthService) RefreshToken(refreshToken string) (*TokenPair, error) {
	key, err := s.keyRepo.FindByRefreshToken(refreshToken)
	if err != nil || key == nil {
		return nil, jwt.ErrSignatureInvalid
	}

	if key.ExpiresAt.Before(time.Now()) {
		s.keyRepo.Delete(key.ID)
		return nil, jwt.ErrSignatureInvalid
	}

	user, err := s.userRepo.GetByID(key.UserID)
	if err != nil || user == nil {
		return nil, jwt.ErrSignatureInvalid
	}

	if key.TokenVersion != user.TokenVersion {
		s.keyRepo.Delete(key.ID)
		return nil, jwt.ErrSignatureInvalid
	}

	s.keyRepo.Delete(key.ID)

	return s.GenerateTokenPair(user, key.UserAgent)
}

func (s *AuthService) RevokeAllSessions(userID string) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil || user == nil {
		return nil
	}

	user.TokenVersion++
	if err := s.userRepo.Update(user); err != nil {
		return err
	}

	s.keyRepo.DeleteByUserID(userID)

	return nil
}

func (s *AuthService) GetPublicKey() *rsa.PublicKey {
	if s.privateKey != nil {
		return &s.privateKey.PublicKey
	}
	return nil
}
