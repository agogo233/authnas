package repository

import (
	"github.com/authnas/authnas/go-server/internal/model"
	"gorm.io/gorm"
)

type OIDCPayloadRepository struct {
	db *gorm.DB
}

func NewOIDCPayloadRepository(db *gorm.DB) *OIDCPayloadRepository {
	return &OIDCPayloadRepository{db: db}
}

func (r *OIDCPayloadRepository) Create(payload *model.OIDCPayload) error {
	return r.db.Create(payload).Error
}

func (r *OIDCPayloadRepository) GetByID(id string) (*model.OIDCPayload, error) {
	var payload model.OIDCPayload
	err := r.db.First(&payload, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &payload, nil
}

func (r *OIDCPayloadRepository) GetByUID(uid string) (*model.OIDCPayload, error) {
	var payload model.OIDCPayload
	err := r.db.First(&payload, "uid = ?", uid).Error
	if err != nil {
		return nil, err
	}
	return &payload, nil
}

func (r *OIDCPayloadRepository) Update(payload *model.OIDCPayload) error {
	return r.db.Save(payload).Error
}

func (r *OIDCPayloadRepository) Delete(id string) error {
	return r.db.Delete(&model.OIDCPayload{}, "id = ?", id).Error
}

func (r *OIDCPayloadRepository) DeleteByUID(uid string) error {
	return r.db.Delete(&model.OIDCPayload{}, "uid = ?", uid).Error
}

func (r *OIDCPayloadRepository) GetAndDeleteByUID(uid string) (*model.OIDCPayload, error) {
	var payload model.OIDCPayload
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&payload, "uid = ?", uid).Error; err != nil {
			return err
		}
		if err := tx.Delete(&model.OIDCPayload{}, "uid = ?", uid).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &payload, nil
}

func (r *OIDCPayloadRepository) GetByUIDForUpdate(tx *gorm.DB, uid string) (*model.OIDCPayload, error) {
	var payload model.OIDCPayload
	err := tx.Set("gorm:query_option", "FOR UPDATE").First(&payload, "uid = ?", uid).Error
	if err != nil {
		return nil, err
	}
	return &payload, nil
}

func (r *OIDCPayloadRepository) DeleteByUIDTx(tx *gorm.DB, uid string) error {
	return tx.Delete(&model.OIDCPayload{}, "uid = ?", uid).Error
}

func (r *OIDCPayloadRepository) DeleteExpired() (int64, error) {
	const batchSize = 100
	var totalDeleted int64

	for {
		result := r.db.Exec("DELETE FROM `oidc_payload` WHERE `id` IN (SELECT `id` FROM `oidc_payload` WHERE expires_at < datetime('now') LIMIT ?)", batchSize)
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
