package repository

import (
	"github.com/authnas/authnas/go-server/internal/model"
	"gorm.io/gorm"
)

type EmailVerificationRepository struct {
	db *gorm.DB
}

func NewEmailVerificationRepository(db *gorm.DB) *EmailVerificationRepository {
	return &EmailVerificationRepository{db: db}
}

func (r *EmailVerificationRepository) Create(ev *model.EmailVerification) error {
	return r.db.Create(ev).Error
}

func (r *EmailVerificationRepository) GetByID(id string) (*model.EmailVerification, error) {
	var ev model.EmailVerification
	err := r.db.First(&ev, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &ev, nil
}

func (r *EmailVerificationRepository) GetByCode(code string) (*model.EmailVerification, error) {
	var ev model.EmailVerification
	err := r.db.First(&ev, "code = ?", code).Error
	if err != nil {
		return nil, err
	}
	return &ev, nil
}

func (r *EmailVerificationRepository) Delete(id string) error {
	return r.db.Delete(&model.EmailVerification{}, "id = ?", id).Error
}

func (r *EmailVerificationRepository) DeleteExpired() error {
	return r.db.Delete(&model.EmailVerification{}, "expires_at < datetime('now')").Error
}

func (r *EmailVerificationRepository) DeleteByUserID(userID string) error {
	return r.db.Delete(&model.EmailVerification{}, "user_id = ?", userID).Error
}

type PasswordResetRepository struct {
	db *gorm.DB
}

func NewPasswordResetRepository(db *gorm.DB) *PasswordResetRepository {
	return &PasswordResetRepository{db: db}
}

func (r *PasswordResetRepository) Create(pr *model.PasswordReset) error {
	return r.db.Create(pr).Error
}

func (r *PasswordResetRepository) GetByID(id string) (*model.PasswordReset, error) {
	var pr model.PasswordReset
	err := r.db.First(&pr, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &pr, nil
}

func (r *PasswordResetRepository) GetByCode(code string) (*model.PasswordReset, error) {
	var pr model.PasswordReset
	err := r.db.First(&pr, "code = ?", code).Error
	if err != nil {
		return nil, err
	}
	return &pr, nil
}

func (r *PasswordResetRepository) Delete(id string) error {
	return r.db.Delete(&model.PasswordReset{}, "id = ?", id).Error
}

func (r *PasswordResetRepository) DeleteExpired() error {
	return r.db.Delete(&model.PasswordReset{}, "expires_at < datetime('now')").Error
}

func (r *PasswordResetRepository) DeleteByUserID(userID string) error {
	return r.db.Delete(&model.PasswordReset{}, "user_id = ?", userID).Error
}
