package repository

import (
	"github.com/authnas/authnas/go-server/internal/model"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepository) GetByID(id string) (*model.User, error) {
	var user model.User
	err := r.db.First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByIDForUpdate(tx *gorm.DB, id string) (*model.User, error) {
	var user model.User
	err := tx.Set("gorm:query_option", "FOR UPDATE").First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByEmail(email string) (*model.User, error) {
	var user model.User
	err := r.db.First(&user, "email = ?", email).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByUsername(username string) (*model.User, error) {
	var user model.User
	err := r.db.First(&user, "username = ?", username).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByInput(input string) (*model.User, error) {
	var user model.User
	err := r.db.Where("username = ? OR email = ?", input, input).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) Update(user *model.User) error {
	return r.db.Save(user).Error
}

func (r *UserRepository) Delete(id string) error {
	return r.db.Delete(&model.User{}, "id = ?", id).Error
}

func (r *UserRepository) List(offset, limit int) ([]*model.User, int64, error) {
	var users []*model.User
	var total int64

	r.db.Model(&model.User{}).Count(&total)
	err := r.db.Order("created_at DESC").Offset(offset).Limit(limit).Find(&users).Error
	if err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

func (r *UserRepository) Count() (int64, error) {
	var total int64
	err := r.db.Model(&model.User{}).Count(&total).Error
	return total, err
}

func (r *UserRepository) CountAdmins() (int64, error) {
	var total int64
	err := r.db.Model(&model.User{}).Where("is_admin = ?", true).Count(&total).Error
	return total, err
}

func (r *UserRepository) Search(query string, offset, limit int) ([]*model.User, int64, error) {
	var users []*model.User
	var total int64

	searchQuery := "%" + query + "%"
	db := r.db.Model(&model.User{}).Where("username LIKE ? OR email LIKE ? OR name LIKE ?", searchQuery, searchQuery, searchQuery)
	db.Count(&total)
	err := db.Order("created_at DESC").Offset(offset).Limit(limit).Find(&users).Error
	if err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

func (r *UserRepository) IncrementTokenVersion(id string) error {
	return r.db.Model(&model.User{}).Where("id = ?", id).Update("token_version", gorm.Expr("token_version + 1")).Error
}
