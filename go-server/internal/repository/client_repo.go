package repository

import (
	"github.com/authnas/authnas/go-server/internal/model"
	"gorm.io/gorm"
)

type ClientRepository struct {
	db *gorm.DB
}

func NewClientRepository(db *gorm.DB) *ClientRepository {
	return &ClientRepository{db: db}
}

func (r *ClientRepository) Create(client *model.Client) error {
	return r.db.Create(client).Error
}

func (r *ClientRepository) GetByID(id string) (*model.Client, error) {
	var client model.Client
	err := r.db.First(&client, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &client, nil
}

func (r *ClientRepository) GetByClientID(clientID string) (*model.Client, error) {
	var client model.Client
	err := r.db.First(&client, "client_id = ?", clientID).Error
	if err != nil {
		return nil, err
	}
	return &client, nil
}

func (r *ClientRepository) Update(client *model.Client) error {
	return r.db.Save(client).Error
}

func (r *ClientRepository) Delete(id string) error {
	return r.db.Delete(&model.Client{}, "id = ?", id).Error
}

func (r *ClientRepository) List(offset, limit int) ([]*model.Client, int64, error) {
	var clients []*model.Client
	var total int64

	r.db.Model(&model.Client{}).Count(&total)
	err := r.db.Offset(offset).Limit(limit).Find(&clients).Error
	if err != nil {
		return nil, 0, err
	}
	return clients, total, nil
}
