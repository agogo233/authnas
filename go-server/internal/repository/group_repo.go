package repository

import (
	"github.com/authnas/authnas/go-server/internal/model"
	"gorm.io/gorm"
)

type GroupRepository struct {
	db *gorm.DB
}

func NewGroupRepository(db *gorm.DB) *GroupRepository {
	return &GroupRepository{db: db}
}

func (r *GroupRepository) Create(group *model.Group) error {
	return r.db.Create(group).Error
}

func (r *GroupRepository) GetByID(id string) (*model.Group, error) {
	var group model.Group
	err := r.db.First(&group, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *GroupRepository) GetByName(name string) (*model.Group, error) {
	var group model.Group
	err := r.db.First(&group, "name = ?", name).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *GroupRepository) Update(group *model.Group) error {
	return r.db.Save(group).Error
}

func (r *GroupRepository) Delete(id string) error {
	return r.db.Delete(&model.Group{}, "id = ?", id).Error
}

func (r *GroupRepository) List(offset, limit int) ([]*model.Group, int64, error) {
	var groups []*model.Group
	var total int64

	r.db.Model(&model.Group{}).Count(&total)
	err := r.db.Offset(offset).Limit(limit).Find(&groups).Error
	if err != nil {
		return nil, 0, err
	}
	return groups, total, nil
}

func (r *GroupRepository) AddUser(groupID, userID string) error {
	ug := model.UserGroup{
		ID:        generateID(),
		UserID:    userID,
		GroupID:   groupID,
		CreatedAt: now(),
	}
	return r.db.Create(&ug).Error
}

func (r *GroupRepository) RemoveUser(groupID, userID string) error {
	return r.db.Delete(&model.UserGroup{}, "group_id = ? AND user_id = ?", groupID, userID).Error
}

func (r *GroupRepository) GetUserGroups(userID string) ([]*model.Group, error) {
	var groups []*model.Group
	err := r.db.Joins("JOIN user_group ON user_group.group_id = groups.id").
		Where("user_group.user_id = ?", userID).
		Find(&groups).Error
	if err != nil {
		return nil, err
	}
	return groups, nil
}

func (r *GroupRepository) DeleteByUserID(userID string) error {
	return r.db.Delete(&model.UserGroup{}, "user_id = ?", userID).Error
}
