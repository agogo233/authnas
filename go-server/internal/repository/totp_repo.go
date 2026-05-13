package repository

import (
	"github.com/authnas/authnas/go-server/internal/model"
	"gorm.io/gorm"
)

type TOTPRepository struct {
	db *gorm.DB
}

func NewTOTPRepository(db *gorm.DB) *TOTPRepository {
	return &TOTPRepository{db: db}
}

func (r *TOTPRepository) Create(totp *model.TOTP) error {
	return r.db.Create(totp).Error
}

func (r *TOTPRepository) GetByID(id string) (*model.TOTP, error) {
	var totp model.TOTP
	err := r.db.First(&totp, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &totp, nil
}

func (r *TOTPRepository) GetByUserID(userID string) (*model.TOTP, error) {
	var totp model.TOTP
	err := r.db.First(&totp, "user_id = ?", userID).Error
	if err != nil {
		return nil, err
	}
	return &totp, nil
}

func (r *TOTPRepository) Update(totp *model.TOTP) error {
	return r.db.Save(totp).Error
}

func (r *TOTPRepository) Delete(id string) error {
	return r.db.Delete(&model.TOTP{}, "id = ?", id).Error
}

func (r *TOTPRepository) DeleteByUserID(userID string) error {
	return r.db.Delete(&model.TOTP{}, "user_id = ?", userID).Error
}
