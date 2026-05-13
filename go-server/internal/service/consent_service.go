package service

import (
	"github.com/authnas/authnas/go-server/internal/config"
	"github.com/authnas/authnas/go-server/internal/model"
	"github.com/authnas/authnas/go-server/internal/repository"
)

type ConsentService struct {
	cfg         *config.Config
	consentRepo *repository.ConsentRepository
	clientRepo  *repository.ClientRepository
	userRepo    *repository.UserRepository
}

func NewConsentService(
	cfg *config.Config,
	consentRepo *repository.ConsentRepository,
	clientRepo *repository.ClientRepository,
	userRepo *repository.UserRepository,
) *ConsentService {
	return &ConsentService{
		cfg:         cfg,
		consentRepo: consentRepo,
		clientRepo:  clientRepo,
		userRepo:    userRepo,
	}
}

func (s *ConsentService) Create(consent *model.Consent) error {
	return s.consentRepo.Create(consent)
}

func (s *ConsentService) GetByID(id string) (*model.Consent, error) {
	return s.consentRepo.GetByID(id)
}

func (s *ConsentService) GetByUserAndClient(userID, clientID string) (*model.Consent, error) {
	return s.consentRepo.GetByUserAndClient(userID, clientID)
}

func (s *ConsentService) Update(consent *model.Consent) error {
	return s.consentRepo.Update(consent)
}

func (s *ConsentService) Delete(id string) error {
	return s.consentRepo.Delete(id)
}
