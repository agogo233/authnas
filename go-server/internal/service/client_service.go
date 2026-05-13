package service

import (
	"github.com/authnas/authnas/go-server/internal/config"
	"github.com/authnas/authnas/go-server/internal/model"
	"github.com/authnas/authnas/go-server/internal/repository"
)

type ClientService struct {
	cfg        *config.Config
	clientRepo *repository.ClientRepository
}

func NewClientService(cfg *config.Config, clientRepo *repository.ClientRepository) *ClientService {
	return &ClientService{
		cfg:        cfg,
		clientRepo: clientRepo,
	}
}

func (s *ClientService) Create(client *model.Client) error {
	return s.clientRepo.Create(client)
}

func (s *ClientService) GetByID(id string) (*model.Client, error) {
	return s.clientRepo.GetByID(id)
}

func (s *ClientService) GetByClientID(clientID string) (*model.Client, error) {
	return s.clientRepo.GetByClientID(clientID)
}

func (s *ClientService) Update(client *model.Client) error {
	return s.clientRepo.Update(client)
}

func (s *ClientService) Delete(id string) error {
	return s.clientRepo.Delete(id)
}

func (s *ClientService) List(offset, limit int) ([]*model.Client, int64, error) {
	return s.clientRepo.List(offset, limit)
}
