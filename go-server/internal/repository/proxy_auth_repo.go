package repository

import (
	"github.com/authnas/authnas/go-server/internal/model"
	"gorm.io/gorm"
)

type ProxyAuthRepository struct {
	db *gorm.DB
}

func NewProxyAuthRepository(db *gorm.DB) *ProxyAuthRepository {
	return &ProxyAuthRepository{db: db}
}

func (r *ProxyAuthRepository) Create(pa *model.ProxyAuth) error {
	return r.db.Create(pa).Error
}

func (r *ProxyAuthRepository) GetByID(id string) (*model.ProxyAuth, error) {
	var pa model.ProxyAuth
	err := r.db.First(&pa, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &pa, nil
}

func (r *ProxyAuthRepository) GetEnabled() ([]*model.ProxyAuth, error) {
	var pas []*model.ProxyAuth
	err := r.db.Where("enabled = ?", true).Find(&pas).Error
	if err != nil {
		return nil, err
	}
	return pas, nil
}

func (r *ProxyAuthRepository) Update(pa *model.ProxyAuth) error {
	return r.db.Save(pa).Error
}

func (r *ProxyAuthRepository) Delete(id string) error {
	return r.db.Delete(&model.ProxyAuth{}, "id = ?", id).Error
}

func (r *ProxyAuthRepository) List(offset, limit int) ([]*model.ProxyAuth, int64, error) {
	var pas []*model.ProxyAuth
	var total int64

	r.db.Model(&model.ProxyAuth{}).Count(&total)
	err := r.db.Offset(offset).Limit(limit).Find(&pas).Error
	if err != nil {
		return nil, 0, err
	}
	return pas, total, nil
}

type EmailLogRepository struct {
	db *gorm.DB
}

func NewEmailLogRepository(db *gorm.DB) *EmailLogRepository {
	return &EmailLogRepository{db: db}
}

func (r *EmailLogRepository) Create(log *model.EmailLog) error {
	return r.db.Create(log).Error
}

func (r *EmailLogRepository) GetByID(id string) (*model.EmailLog, error) {
	var log model.EmailLog
	err := r.db.First(&log, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}

func (r *EmailLogRepository) List(offset, limit int) ([]*model.EmailLog, int64, error) {
	var logs []*model.EmailLog
	var total int64

	r.db.Model(&model.EmailLog{}).Count(&total)
	err := r.db.Offset(offset).Limit(limit).Order("created_at DESC").Find(&logs).Error
	if err != nil {
		return nil, 0, err
	}
	return logs, total, nil
}

func (r *EmailLogRepository) DeleteOlderThan(days int) error {
	return r.db.Delete(&model.EmailLog{}, "created_at < datetime('now', '-' || ? || ' days')", days).Error
}
