package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/authnas/authnas/go-server/internal/config"
	"github.com/authnas/authnas/go-server/internal/model"
	"github.com/authnas/authnas/go-server/internal/repository"
	"github.com/authnas/authnas/go-server/pkg/utils"
	"gorm.io/gorm"
)

type InvitationService struct {
	cfg            *config.Config
	invitationRepo *repository.InvitationRepository
	userRepo       *repository.UserRepository
}

func NewInvitationService(
	cfg *config.Config,
	invitationRepo *repository.InvitationRepository,
	userRepo *repository.UserRepository,
) *InvitationService {
	return &InvitationService{
		cfg:            cfg,
		invitationRepo: invitationRepo,
		userRepo:       userRepo,
	}
}

type InvitationValidation struct {
	Valid        bool
	Invitation   *model.Invitation
	ErrorMessage string
}

func (s *InvitationService) ValidateInvitation(id, code string) (*InvitationValidation, error) {
	invitation, err := s.invitationRepo.GetByID(id)
	if err != nil || invitation == nil {
		return &InvitationValidation{Valid: false, ErrorMessage: "invitation not found"}, nil
	}

	if invitation.Code != code {
		return &InvitationValidation{Valid: false, ErrorMessage: "invalid invitation code"}, nil
	}

	if invitation.ExpiresAt.Before(time.Now()) {
		return &InvitationValidation{Valid: false, ErrorMessage: "invitation has expired"}, nil
	}

	if invitation.MaxUses != nil && invitation.UsedCount >= *invitation.MaxUses {
		return &InvitationValidation{Valid: false, ErrorMessage: "invitation has reached maximum uses"}, nil
	}

	return &InvitationValidation{
		Valid:      true,
		Invitation: invitation,
	}, nil
}

func (s *InvitationService) ConsumeInvitation(tx *gorm.DB, id string) error {
	return s.invitationRepo.IncrementUsedCount(tx, id)
}

func (s *InvitationService) Create(email, username string, expiresIn time.Duration, groupID *string, scopes *string, maxUses *int, createdBy string) (*model.Invitation, error) {
	code, _ := utils.NewRandom().GenerateToken(32)

	if expiresIn == 0 {
		expiresIn = 7 * 24 * time.Hour
	}

	invitation := &model.Invitation{
		ID:        uuid.New().String(),
		Email:     email,
		Username:  &username,
		Code:      code,
		Scopes:    scopes,
		GroupID:   groupID,
		MaxUses:   maxUses,
		UsedCount: 0,
		ExpiresAt: time.Now().Add(expiresIn),
		CreatedAt: time.Now(),
		CreatedBy: &createdBy,
	}

	if err := s.invitationRepo.Create(invitation); err != nil {
		return nil, err
	}
	return invitation, nil
}

func (s *InvitationService) GetByID(id string) (*model.Invitation, error) {
	return s.invitationRepo.GetByID(id)
}

func (s *InvitationService) GetByCode(code string) (*model.Invitation, error) {
	return s.invitationRepo.GetByCode(code)
}

func (s *InvitationService) List(offset, limit int) ([]*model.Invitation, int64, error) {
	return s.invitationRepo.List(offset, limit)
}

func (s *InvitationService) Delete(id string) error {
	return s.invitationRepo.Delete(id)
}
