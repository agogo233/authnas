package repository

import (
	"github.com/authnas/authnas/go-server/internal/model"
	"gorm.io/gorm"
)

type PasskeyRepository struct {
	db *gorm.DB
}

func NewPasskeyRepository(db *gorm.DB) *PasskeyRepository {
	return &PasskeyRepository{db: db}
}

func (r *PasskeyRepository) Create(passkey *model.Passkey) error {
	return r.db.Create(passkey).Error
}

func (r *PasskeyRepository) GetByID(id string) (*model.Passkey, error) {
	var passkey model.Passkey
	err := r.db.First(&passkey, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &passkey, nil
}

func (r *PasskeyRepository) GetByCredentialID(credentialID string) (*model.Passkey, error) {
	var passkey model.Passkey
	err := r.db.Preload("User").First(&passkey, "credential_id = ?", credentialID).Error
	if err != nil {
		return nil, err
	}
	return &passkey, nil
}

func (r *PasskeyRepository) GetByUserID(userID string) ([]*model.Passkey, error) {
	var passkeys []*model.Passkey
	err := r.db.Where("user_id = ?", userID).Find(&passkeys).Error
	if err != nil {
		return nil, err
	}
	return passkeys, nil
}

func (r *PasskeyRepository) Update(passkey *model.Passkey) error {
	return r.db.Save(passkey).Error
}

func (r *PasskeyRepository) Delete(id string) error {
	return r.db.Delete(&model.Passkey{}, "id = ?", id).Error
}

func (r *PasskeyRepository) DeleteByUserID(userID string) error {
	return r.db.Delete(&model.Passkey{}, "user_id = ?", userID).Error
}

type PasskeyAuthOptionsRepository struct {
	db *gorm.DB
}

func NewPasskeyAuthOptionsRepository(db *gorm.DB) *PasskeyAuthOptionsRepository {
	return &PasskeyAuthOptionsRepository{db: db}
}

func (r *PasskeyAuthOptionsRepository) Create(opts *model.PasskeyAuthOptions) error {
	return r.db.Create(opts).Error
}

func (r *PasskeyAuthOptionsRepository) GetByID(id string) (*model.PasskeyAuthOptions, error) {
	var opts model.PasskeyAuthOptions
	err := r.db.First(&opts, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &opts, nil
}

func (r *PasskeyAuthOptionsRepository) GetByUserID(userID string) (*model.PasskeyAuthOptions, error) {
	var opts model.PasskeyAuthOptions
	err := r.db.Where("user_id = ?", userID).Order("created_at DESC").First(&opts).Error
	if err != nil {
		return nil, err
	}
	return &opts, nil
}

func (r *PasskeyAuthOptionsRepository) GetByChallenge(challenge string) (*model.PasskeyAuthOptions, error) {
	var opts model.PasskeyAuthOptions
	err := r.db.Where("challenge = ?", challenge).First(&opts).Error
	if err != nil {
		return nil, err
	}
	return &opts, nil
}

func (r *PasskeyAuthOptionsRepository) Delete(id string) error {
	return r.db.Delete(&model.PasskeyAuthOptions{}, "id = ?", id).Error
}

func (r *PasskeyAuthOptionsRepository) DeleteExpired() error {
	return r.db.Delete(&model.PasskeyAuthOptions{}, "expires_at < datetime('now')").Error
}
