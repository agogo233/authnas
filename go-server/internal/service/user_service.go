package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/authnas/authnas/go-server/internal/config"
	"github.com/authnas/authnas/go-server/internal/model"
	"github.com/authnas/authnas/go-server/internal/repository"
	cryptopkg "github.com/authnas/authnas/go-server/pkg/crypto"
	"github.com/authnas/authnas/go-server/pkg/utils"
	"github.com/google/uuid"
	"github.com/nbutton23/zxcvbn-go"
	"gorm.io/gorm"
)

// InvitationVerifier defines the interface for invitation validation and consumption.
// This breaks the circular dependency between UserService and InvitationService.
type InvitationVerifier interface {
	ValidateInvitation(id, code string) (*InvitationValidation, error)
	ConsumeInvitation(tx *gorm.DB, id string) error
}

type UserService struct {
	cfg                   *config.Config
	userRepo              *repository.UserRepository
	groupRepo             *repository.GroupRepository
	keyRepo               *repository.KeyRepository
	totpRepo              *repository.TOTPRepository
	passkeyRepo           *repository.PasskeyRepository
	emailVerificationRepo *repository.EmailVerificationRepository
	passwordResetRepo     *repository.PasswordResetRepository
	consentRepo           *repository.ConsentRepository
	emailService          *EmailService
	invitationService     InvitationVerifier
	random                *utils.RandomUtil
	time                  *utils.TimeUtil
	db                    *gorm.DB
}

func NewUserService(
	cfg *config.Config,
	userRepo *repository.UserRepository,
	groupRepo *repository.GroupRepository,
	keyRepo *repository.KeyRepository,
	totpRepo *repository.TOTPRepository,
	passkeyRepo *repository.PasskeyRepository,
	emailVerificationRepo *repository.EmailVerificationRepository,
	passwordResetRepo *repository.PasswordResetRepository,
	consentRepo *repository.ConsentRepository,
	emailService *EmailService,
	invitationService InvitationVerifier,
	random *utils.RandomUtil,
	time *utils.TimeUtil,
	db *gorm.DB,
) *UserService {
	return &UserService{
		cfg:                   cfg,
		userRepo:              userRepo,
		groupRepo:             groupRepo,
		keyRepo:               keyRepo,
		totpRepo:              totpRepo,
		passkeyRepo:           passkeyRepo,
		emailVerificationRepo: emailVerificationRepo,
		passwordResetRepo:     passwordResetRepo,
		consentRepo:           consentRepo,
		emailService:          emailService,
		invitationService:     invitationService,
		random:                random,
		time:                  time,
		db:                    db,
	}
}

func (s *UserService) GetConfig() *config.Config {
	return s.cfg
}

func (s *UserService) Create(email, username, password string) (*model.User, error) {
	if username == "" {
		return nil, errors.New("username is required")
	}
	if password == "" {
		password, _ = utils.NewRandom().GenerateToken(16)
	}

	passwordHash, err := s.hashPassword(password)
	if err != nil {
		return nil, err
	}

	now := s.time.Now()
	user := &model.User{
		ID:            uuid.New().String(),
		Email:         stringPtr(email),
		Username:      username,
		PasswordHash:  stringPtr(passwordHash),
		EmailVerified: !s.cfg.Security.EmailVerification,
		Approved:      !s.cfg.Security.SignupRequiresApproval,
		MFARequired:   false,
		TokenVersion:  0,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if s.db != nil {
		err = s.db.Transaction(func(tx *gorm.DB) error {
			if email != "" {
				var existingEmail model.User
				if err := tx.Where("email = ?", email).First(&existingEmail).Error; err == nil {
					return errors.New("email already exists")
				} else if !errors.Is(err, gorm.ErrRecordNotFound) {
					return err
				}
			}

			var existingUsername model.User
			if err := tx.Where("username = ?", username).First(&existingUsername).Error; err == nil {
				return errors.New("username already exists")
			} else if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}

			return tx.Create(user).Error
		})
		if err != nil {
			return nil, err
		}
		return user, nil
	}

	if email != "" {
		existingEmail, _ := s.userRepo.GetByEmail(email)
		if existingEmail != nil {
			return nil, errors.New("email already exists")
		}
	}

	existingUsername, _ := s.userRepo.GetByUsername(username)
	if existingUsername != nil {
		return nil, errors.New("username already exists")
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) CreateWithInvitation(inviteID, code, email, username, password string) (*model.User, *model.Invitation, error) {
	if username == "" || password == "" {
		return nil, nil, errors.New("username and password are required")
	}

	var invitation *model.Invitation
	if inviteID != "" && code != "" {
		if s.invitationService == nil {
			return nil, nil, errors.New("invitation service not available")
		}

		validation, err := s.invitationService.ValidateInvitation(inviteID, code)
		if err != nil {
			return nil, nil, errors.New("failed to validate invitation")
		}
		if !validation.Valid {
			return nil, nil, errors.New(validation.ErrorMessage)
		}
		invitation = validation.Invitation
	}

	passwordHash, err := s.hashPassword(password)
	if err != nil {
		return nil, nil, err
	}

	now := s.time.Now()
	user := &model.User{
		ID:            uuid.New().String(),
		Email:         stringPtr(email),
		Username:      username,
		PasswordHash:  stringPtr(passwordHash),
		EmailVerified: !s.cfg.Security.EmailVerification,
		Approved:      !s.cfg.Security.SignupRequiresApproval,
		MFARequired:   false,
		TokenVersion:  0,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if s.db != nil {
		err = s.db.Transaction(func(tx *gorm.DB) error {
			if email != "" {
				var existingEmail model.User
				if err := tx.Where("email = ?", email).First(&existingEmail).Error; err == nil {
					return errors.New("email already exists")
				} else if !errors.Is(err, gorm.ErrRecordNotFound) {
					return err
				}
			}

			var existingUsername model.User
			if err := tx.Where("username = ?", username).First(&existingUsername).Error; err == nil {
				return errors.New("username already exists")
			} else if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}

			if err := tx.Create(user).Error; err != nil {
				return err
			}

			if invitation != nil && s.invitationService != nil {
				if err := s.invitationService.ConsumeInvitation(tx, invitation.ID); err != nil {
					return err
				}

				if invitation.GroupID != nil && *invitation.GroupID != "" {
					userGroup := &model.UserGroup{
						ID:        uuid.New().String(),
						UserID:    user.ID,
						GroupID:   *invitation.GroupID,
						CreatedAt: now,
					}
					if err := tx.Create(userGroup).Error; err != nil {
						return err
					}
				}
			}

			return nil
		})
		if err != nil {
			return nil, nil, err
		}
		return user, invitation, nil
	}

	if email != "" {
		existingEmail, _ := s.userRepo.GetByEmail(email)
		if existingEmail != nil {
			return nil, nil, errors.New("email already exists")
		}
	}

	existingUsername, _ := s.userRepo.GetByUsername(username)
	if existingUsername != nil {
		return nil, nil, errors.New("username already exists")
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, nil, err
	}

	return user, invitation, nil
}

func (s *UserService) GetByInput(input string) (*model.User, error) {
	return s.userRepo.GetByInput(input)
}

func (s *UserService) GetByUsername(username string) (*model.User, error) {
	return s.userRepo.GetByUsername(username)
}

func (s *UserService) GetByID(id string) (*model.User, error) {
	return s.userRepo.GetByID(id)
}

func (s *UserService) List(offset, limit int) ([]*model.User, int64, error) {
	return s.userRepo.List(offset, limit)
}

func (s *UserService) Search(query string, offset, limit int) ([]*model.User, int64, error) {
	return s.userRepo.Search(query, offset, limit)
}

func (s *UserService) Count() (int64, error) {
	return s.userRepo.Count()
}

func (s *UserService) CountAdmins() (int64, error) {
	return s.userRepo.CountAdmins()
}

func (s *UserService) Delete(id string) error {
	if s.db != nil {
		return s.db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Where("user_id = ?", id).Delete(&model.Key{}).Error; err != nil {
				return err
			}
			if err := tx.Where("user_id = ?", id).Delete(&model.TOTP{}).Error; err != nil {
				return err
			}
			if err := tx.Where("user_id = ?", id).Delete(&model.Passkey{}).Error; err != nil {
				return err
			}
			if err := tx.Where("user_id = ?", id).Delete(&model.EmailVerification{}).Error; err != nil {
				return err
			}
			if err := tx.Where("user_id = ?", id).Delete(&model.PasswordReset{}).Error; err != nil {
				return err
			}
			if err := tx.Where("user_id = ?", id).Delete(&model.Consent{}).Error; err != nil {
				return err
			}
			if err := tx.Exec("DELETE FROM user_group WHERE user_id = ?", id).Error; err != nil {
				return err
			}
			return tx.Where("id = ?", id).Delete(&model.User{}).Error
		})
	}

	s.keyRepo.DeleteByUserID(id)
	s.totpRepo.DeleteByUserID(id)
	s.passkeyRepo.DeleteByUserID(id)
	s.emailVerificationRepo.DeleteByUserID(id)
	s.passwordResetRepo.DeleteByUserID(id)
	s.consentRepo.DeleteByUserID(id)
	s.groupRepo.DeleteByUserID(id)
	return s.userRepo.Delete(id)
}

func (s *UserService) ResetPassword(id, newPassword string) error {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return errors.New("user not found")
	}

	if s.cfg.Security.PasswordStrength > 0 {
		result := zxcvbn.PasswordStrength(newPassword, nil)
		if result.Score < s.cfg.Security.PasswordStrength {
			return errors.New("password is too weak")
		}
	}

	if s.cfg.Security.PasswordMinLength > 0 && len(newPassword) < s.cfg.Security.PasswordMinLength {
		return errors.New("password is too short")
	}

	passwordHash, err := s.hashPassword(newPassword)
	if err != nil {
		return err
	}

	user.PasswordHash = stringPtr(passwordHash)
	user.TokenVersion++
	user.UpdatedAt = s.time.Now()

	return s.userRepo.Update(user)
}

func (s *UserService) ForgotPassword(email string) error {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil || user == nil {
		return nil
	}

	code, err := s.random.GenerateToken(32)
	if err != nil {
		return errors.New("failed to generate reset code")
	}

	pr := &model.PasswordReset{
		ID:        generateID(),
		UserID:    user.ID,
		Code:      code,
		ExpiresAt: s.time.Now().Add(1 * time.Hour),
		CreatedAt: s.time.Now(),
	}

	if err := s.passwordResetRepo.Create(pr); err != nil {
		return errors.New("failed to create reset record")
	}

	if err := s.emailService.SendPasswordResetEmail(user, code); err != nil {
		return errors.New("failed to send reset email")
	}

	return nil
}

func (s *UserService) ResetPasswordByCode(code, newPassword string) error {
	pr, err := s.passwordResetRepo.GetByCode(code)
	if err != nil || pr == nil {
		return errors.New("invalid reset code")
	}

	if pr.ExpiresAt.Before(s.time.Now()) {
		return errors.New("reset code expired")
	}

	user, err := s.userRepo.GetByID(pr.UserID)
	if err != nil || user == nil {
		return errors.New("user not found")
	}

	if s.cfg.Security.PasswordStrength > 0 {
		result := zxcvbn.PasswordStrength(newPassword, nil)
		if result.Score < s.cfg.Security.PasswordStrength {
			return errors.New("password is too weak")
		}
	}

	if s.cfg.Security.PasswordMinLength > 0 && len(newPassword) < s.cfg.Security.PasswordMinLength {
		return errors.New("password is too short")
	}

	passwordHash, err := s.hashPassword(newPassword)
	if err != nil {
		return err
	}

	user.PasswordHash = stringPtr(passwordHash)
	user.TokenVersion++
	user.UpdatedAt = s.time.Now()

	if err := s.userRepo.Update(user); err != nil {
		return err
	}

	s.passwordResetRepo.Delete(pr.ID)

	return nil
}

func (s *UserService) VerifyEmail(userID, challenge string) error {
	ev, err := s.emailVerificationRepo.GetByCode(challenge)
	if err != nil || ev == nil {
		return errors.New("invalid verification code")
	}

	if ev.UserID != userID {
		return errors.New("invalid verification code")
	}

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	user.EmailVerified = true
	user.Email = &ev.Email
	if err := s.userRepo.Update(user); err != nil {
		return err
	}

	s.emailVerificationRepo.Delete(ev.ID)

	return nil
}

func (s *UserService) SendEmailVerification(userID string) (string, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil || user == nil {
		return "", errors.New("user not found")
	}

	code, err := s.random.GenerateToken(32)
	if err != nil {
		return "", errors.New("failed to generate verification code")
	}

	ev := &model.EmailVerification{
		ID:        generateID(),
		UserID:    user.ID,
		Email:     *user.Email,
		Code:      code,
		ExpiresAt: s.time.Now().Add(24 * time.Hour),
		CreatedAt: s.time.Now(),
	}

	if err := s.emailVerificationRepo.Create(ev); err != nil {
		return "", errors.New("failed to create verification record")
	}

	if err := s.emailService.SendVerificationEmail(user, code); err != nil {
		return "", errors.New("failed to send verification email")
	}

	return code, nil
}

func (s *UserService) Update(user *model.User) error {
	user.UpdatedAt = s.time.Now()
	return s.userRepo.Update(user)
}

func (s *UserService) CheckEmailAvailable(email string, excludeUserID string) (bool, error) {
	existingUser, err := s.userRepo.GetByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return true, nil
		}
		return false, nil
	}
	if existingUser == nil {
		return true, nil
	}
	if excludeUserID != "" && existingUser.ID == excludeUserID {
		return true, nil
	}
	return false, nil
}

func (s *UserService) UpdatePassword(userID, oldPassword, newPassword string) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	if user.PasswordHash != nil && *user.PasswordHash != "" {
		if oldPassword == "" {
			return errors.New("old password is required")
		}
		if !s.verifyPassword(*user.PasswordHash, oldPassword) {
			return errors.New("invalid old password")
		}
	}

	if s.cfg.Security.PasswordStrength > 0 {
		result := zxcvbn.PasswordStrength(newPassword, nil)
		if result.Score < s.cfg.Security.PasswordStrength {
			return errors.New("password is too weak")
		}
	}

	if s.cfg.Security.PasswordMinLength > 0 && len(newPassword) < s.cfg.Security.PasswordMinLength {
		return fmt.Errorf("password must be at least %d characters", s.cfg.Security.PasswordMinLength)
	}

	passwordHash, err := s.hashPassword(newPassword)
	if err != nil {
		return err
	}

	user.PasswordHash = stringPtr(passwordHash)
	user.TokenVersion++
	user.UpdatedAt = s.time.Now()

	if err := s.userRepo.Update(user); err != nil {
		return err
	}

	s.keyRepo.DeleteByUserID(userID)

	return nil
}

func (s *UserService) RevokeAllSessions(userID string) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil
	}

	user.TokenVersion++
	if err := s.userRepo.Update(user); err != nil {
		return err
	}

	s.keyRepo.DeleteByUserID(userID)

	return nil
}

func (s *UserService) GetUserSessions(userID string) ([]*model.Key, error) {
	return s.keyRepo.GetByUserID(userID)
}

func (s *UserService) RevokeSession(userID, sessionID string) error {
	key, err := s.keyRepo.GetByID(sessionID)
	if err != nil || key == nil {
		return errors.New("session not found")
	}

	if key.UserID != userID {
		return errors.New("session not found")
	}

	return s.keyRepo.Delete(sessionID)
}

func (s *UserService) EnsureInitialAdmin(username, email, password string) error {
	if username == "" || password == "" {
		return nil
	}

	existingAdmin, _ := s.userRepo.GetByUsername(username)
	if existingAdmin != nil {
		return nil
	}

	passwordHash, err := s.hashPassword(password)
	if err != nil {
		return err
	}

	emailPtr := stringPtr(email)
	if email == "" {
		emailPtr = stringPtr(username + "@localhost")
	}

	now := s.time.Now()
	admin := &model.User{
		ID:                 uuid.New().String(),
		Email:              emailPtr,
		Username:           username,
		PasswordHash:       stringPtr(passwordHash),
		EmailVerified:      true,
		Approved:           true,
		IsAdmin:            true,
		MFARequired:        false,
		TokenVersion:       0,
		MustChangePassword: true,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	if err := s.userRepo.Create(admin); err != nil {
		return err
	}

	return nil
}

func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func (s *UserService) hashPassword(password string) (string, error) {
	salt, err := s.random.GenerateRandomBytes(cryptopkg.Argon2SaltLength)
	if err != nil {
		return "", err
	}
	return cryptopkg.HashPassword(password, salt)
}

func (s *UserService) verifyPassword(hashWithSalt, password string) bool {
	return cryptopkg.VerifyPassword(hashWithSalt, password)
}
