package service

import (
	"log"
	"time"

	"github.com/authnas/authnas/go-server/internal/repository"
)

type CleanupService struct {
	keyRepo               *repository.KeyRepository
	passkeyAuthOptsRepo   *repository.PasskeyAuthOptionsRepository
	oidcPayloadRepo       *repository.OIDCPayloadRepository
	emailVerificationRepo *repository.EmailVerificationRepository
	passwordResetRepo     *repository.PasswordResetRepository
	emailLogRepo          *repository.EmailLogRepository
	csrfService           *CSRFService
	stopCh                chan struct{}
}

func NewCleanupService(
	keyRepo *repository.KeyRepository,
	passkeyAuthOptsRepo *repository.PasskeyAuthOptionsRepository,
	oidcPayloadRepo *repository.OIDCPayloadRepository,
	emailVerificationRepo *repository.EmailVerificationRepository,
	passwordResetRepo *repository.PasswordResetRepository,
	emailLogRepo *repository.EmailLogRepository,
	csrfService *CSRFService,
) *CleanupService {
	return &CleanupService{
		keyRepo:               keyRepo,
		passkeyAuthOptsRepo:   passkeyAuthOptsRepo,
		oidcPayloadRepo:       oidcPayloadRepo,
		emailVerificationRepo: emailVerificationRepo,
		passwordResetRepo:     passwordResetRepo,
		emailLogRepo:          emailLogRepo,
		csrfService:           csrfService,
		stopCh:                make(chan struct{}),
	}
}

type CleanupResult struct {
	KeysDeleted               int64
	PasskeyAuthOptionsDeleted int64
	OIDCPayloadsDeleted       int64
	EmailVerificationsDeleted int64
	PasswordResetsDeleted     int64
	EmailLogsDeleted          int64
	CSRFTokensCleaned         int
}

func (s *CleanupService) CleanupExpired() (*CleanupResult, error) {
	result := &CleanupResult{}
	var err error

	if s.keyRepo != nil {
		deleted, err := s.keyRepo.DeleteExpired()
		if err != nil {
			log.Printf("Error cleaning up expired keys: %v", err)
		} else {
			result.KeysDeleted = deleted
		}
	}

	if s.passkeyAuthOptsRepo != nil {
		err = s.passkeyAuthOptsRepo.DeleteExpired()
		if err != nil {
			log.Printf("Error cleaning up expired passkey auth options: %v", err)
		} else {
			result.PasskeyAuthOptionsDeleted = 1
		}
	}

	if s.oidcPayloadRepo != nil {
		result.OIDCPayloadsDeleted, err = s.oidcPayloadRepo.DeleteExpired()
		if err != nil {
			log.Printf("Error cleaning up expired OIDC payloads: %v", err)
		}
	}

	if s.emailVerificationRepo != nil {
		err = s.emailVerificationRepo.DeleteExpired()
		if err != nil {
			log.Printf("Error cleaning up expired email verifications: %v", err)
		} else {
			result.EmailVerificationsDeleted = 1
		}
	}

	if s.passwordResetRepo != nil {
		err = s.passwordResetRepo.DeleteExpired()
		if err != nil {
			log.Printf("Error cleaning up expired password resets: %v", err)
		} else {
			result.PasswordResetsDeleted = 1
		}
	}

	if s.csrfService != nil {
		s.csrfService.CleanupExpired()
		result.CSRFTokensCleaned = 1
	}

	return result, nil
}

func (s *CleanupService) CleanupOldEmailLogs(keepDays int) (int64, error) {
	if s.emailLogRepo == nil {
		return 0, nil
	}
	err := s.emailLogRepo.DeleteOlderThan(keepDays)
	return 0, err
}

func (s *CleanupService) StartCleanupScheduler(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				result, err := s.CleanupExpired()
				if err != nil {
					log.Printf("Cleanup error: %v", err)
				} else {
					log.Printf("Cleanup completed: %+v", result)
				}
			case <-s.stopCh:
				return
			}
		}
	}()
}

func (s *CleanupService) Stop() {
	close(s.stopCh)
	if s.csrfService != nil {
		s.csrfService.Stop()
	}
}
