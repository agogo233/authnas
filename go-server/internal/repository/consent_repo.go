package repository

import (
	"github.com/authnas/authnas/go-server/internal/model"
	"gorm.io/gorm"
)

type ConsentRepository struct {
	db *gorm.DB
}

func NewConsentRepository(db *gorm.DB) *ConsentRepository {
	return &ConsentRepository{db: db}
}

func (r *ConsentRepository) Create(consent *model.Consent) error {
	return r.db.Create(consent).Error
}

func (r *ConsentRepository) GetByID(id string) (*model.Consent, error) {
	var consent model.Consent
	err := r.db.First(&consent, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &consent, nil
}

func (r *ConsentRepository) GetByUserAndClient(userID, clientID string) (*model.Consent, error) {
	var consent model.Consent
	err := r.db.Where("user_id = ? AND client_id = ?", userID, clientID).First(&consent).Error
	if err != nil {
		return nil, err
	}
	return &consent, nil
}

func (r *ConsentRepository) Update(consent *model.Consent) error {
	return r.db.Save(consent).Error
}

func (r *ConsentRepository) Delete(id string) error {
	return r.db.Delete(&model.Consent{}, "id = ?", id).Error
}

func (r *ConsentRepository) DeleteByUserAndClient(userID, clientID string) error {
	return r.db.Delete(&model.Consent{}, "user_id = ? AND client_id = ?", userID, clientID).Error
}

func (r *ConsentRepository) DeleteByUserID(userID string) error {
	return r.db.Delete(&model.Consent{}, "user_id = ?", userID).Error
}
