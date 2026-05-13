package service

import (
	"github.com/authnas/authnas/go-server/internal/config"
	"github.com/authnas/authnas/go-server/internal/model"
	"github.com/authnas/authnas/go-server/internal/repository"
)

type ProxyAuthService struct {
	cfg           *config.Config
	proxyAuthRepo *repository.ProxyAuthRepository
	userRepo      *repository.UserRepository
}

func NewProxyAuthService(
	cfg *config.Config,
	proxyAuthRepo *repository.ProxyAuthRepository,
	userRepo *repository.UserRepository,
) *ProxyAuthService {
	return &ProxyAuthService{
		cfg:           cfg,
		proxyAuthRepo: proxyAuthRepo,
		userRepo:      userRepo,
	}
}

func (s *ProxyAuthService) Create(pa *model.ProxyAuth) error {
	return s.proxyAuthRepo.Create(pa)
}

func (s *ProxyAuthService) GetByID(id string) (*model.ProxyAuth, error) {
	return s.proxyAuthRepo.GetByID(id)
}

func (s *ProxyAuthService) GetEnabled() ([]*model.ProxyAuth, error) {
	return s.proxyAuthRepo.GetEnabled()
}

func (s *ProxyAuthService) Update(pa *model.ProxyAuth) error {
	return s.proxyAuthRepo.Update(pa)
}

func (s *ProxyAuthService) Delete(id string) error {
	return s.proxyAuthRepo.Delete(id)
}

func (s *ProxyAuthService) List(offset, limit int) ([]*model.ProxyAuth, int64, error) {
	return s.proxyAuthRepo.List(offset, limit)
}
