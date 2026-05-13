package service

import (
	"github.com/authnas/authnas/go-server/internal/config"
	"github.com/authnas/authnas/go-server/internal/model"
	"github.com/authnas/authnas/go-server/internal/repository"
)

type GroupService struct {
	cfg       *config.Config
	groupRepo *repository.GroupRepository
	userRepo  *repository.UserRepository
}

func NewGroupService(cfg *config.Config, groupRepo *repository.GroupRepository, userRepo *repository.UserRepository) *GroupService {
	return &GroupService{
		cfg:       cfg,
		groupRepo: groupRepo,
		userRepo:  userRepo,
	}
}

func (s *GroupService) Create(name, description string) (*model.Group, error) {
	group := &model.Group{
		ID:          generateID(),
		Name:        name,
		Description: &description,
		CreatedAt:   now(),
		UpdatedAt:   now(),
	}
	if err := s.groupRepo.Create(group); err != nil {
		return nil, err
	}
	return group, nil
}

func (s *GroupService) GetByID(id string) (*model.Group, error) {
	return s.groupRepo.GetByID(id)
}

func (s *GroupService) List(offset, limit int) ([]*model.Group, int64, error) {
	return s.groupRepo.List(offset, limit)
}

func (s *GroupService) Update(id, name, description string) (*model.Group, error) {
	group, err := s.groupRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	group.Name = name
	if description != "" {
		group.Description = &description
	}
	group.UpdatedAt = now()
	if err := s.groupRepo.Update(group); err != nil {
		return nil, err
	}
	return group, nil
}

func (s *GroupService) Delete(id string) error {
	return s.groupRepo.Delete(id)
}

func (s *GroupService) AddUser(groupID, userID string) error {
	return s.groupRepo.AddUser(groupID, userID)
}

func (s *GroupService) RemoveUser(groupID, userID string) error {
	return s.groupRepo.RemoveUser(groupID, userID)
}

func (s *GroupService) GetUserGroups(userID string) ([]*model.Group, error) {
	return s.groupRepo.GetUserGroups(userID)
}
