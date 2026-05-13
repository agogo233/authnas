package repository

import (
	"crypto/sha256"
	"encoding/base64"
	"time"

	"github.com/authnas/authnas/go-server/internal/model"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type KeyRepository struct {
	db *gorm.DB
}

func NewKeyRepository(db *gorm.DB) *KeyRepository {
	return &KeyRepository{db: db}
}

func (r *KeyRepository) Create(key *model.Key) error {
	return r.db.Create(key).Error
}

func (r *KeyRepository) GetByID(id string) (*model.Key, error) {
	var key model.Key
	err := r.db.First(&key, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &key, nil
}

func (r *KeyRepository) GetByRefreshTokenHash(hash string) (*model.Key, error) {
	var key model.Key
	err := r.db.First(&key, "refresh_token_hash = ?", hash).Error
	if err != nil {
		return nil, err
	}
	return &key, nil
}

func (r *KeyRepository) FindByRefreshToken(token string) (*model.Key, error) {
	h := sha256.Sum256([]byte(token))
	lookupHash := base64.RawStdEncoding.EncodeToString(h[:])

	var key model.Key
	err := r.db.Where("refresh_token_lookup_hash = ? AND expires_at > ?", lookupHash, time.Now()).First(&key).Error
	if err != nil {
		return nil, gorm.ErrRecordNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(key.RefreshTokenHash), []byte(token)); err != nil {
		return nil, gorm.ErrRecordNotFound
	}

	return &key, nil
}

func (r *KeyRepository) FindByRefreshTokenForUpdate(tx *gorm.DB, token string) (*model.Key, error) {
	h := sha256.Sum256([]byte(token))
	lookupHash := base64.RawStdEncoding.EncodeToString(h[:])

	var key model.Key
	err := tx.Set("gorm:query_option", "FOR UPDATE").
		Where("refresh_token_lookup_hash = ? AND expires_at > ?", lookupHash, time.Now()).
		First(&key).Error
	if err != nil {
		return nil, gorm.ErrRecordNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(key.RefreshTokenHash), []byte(token)); err != nil {
		return nil, gorm.ErrRecordNotFound
	}

	return &key, nil
}

func (r *KeyRepository) GetByUserID(userID string) ([]*model.Key, error) {
	var keys []*model.Key
	err := r.db.Where("user_id = ?", userID).Find(&keys).Error
	if err != nil {
		return nil, err
	}
	return keys, nil
}

func (r *KeyRepository) Delete(id string) error {
	return r.db.Delete(&model.Key{}, "id = ?", id).Error
}

func (r *KeyRepository) DeleteByUserID(userID string) error {
	return r.db.Delete(&model.Key{}, "user_id = ?", userID).Error
}

func (r *KeyRepository) DeleteByClientID(clientID string) error {
	return r.db.Delete(&model.Key{}, "client_id = ?", clientID).Error
}

func (r *KeyRepository) DeleteExpired() (int64, error) {
	const batchSize = 100
	var totalDeleted int64

	for {
		result := r.db.Exec("DELETE FROM `key` WHERE `id` IN (SELECT `id` FROM `key` WHERE expires_at < datetime('now') LIMIT ?)", batchSize)
		if result.Error != nil {
			return totalDeleted, result.Error
		}
		deleted := result.RowsAffected
		totalDeleted += deleted
		if deleted < batchSize {
			break
		}
	}
	return totalDeleted, nil
}

func (r *KeyRepository) CreateWithTx(tx *gorm.DB, key *model.Key) error {
	return tx.Create(key).Error
}

func (r *KeyRepository) DeleteWithTx(tx *gorm.DB, id string) error {
	return tx.Delete(&model.Key{}, "id = ?", id).Error
}
