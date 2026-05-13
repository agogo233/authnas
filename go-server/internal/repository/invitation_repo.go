package repository

import (
	"github.com/authnas/authnas/go-server/internal/model"
	"gorm.io/gorm"
)

type InvitationRepository struct {
	db *gorm.DB
}

func NewInvitationRepository(db *gorm.DB) *InvitationRepository {
	return &InvitationRepository{db: db}
}

func (r *InvitationRepository) Create(invitation *model.Invitation) error {
	return r.db.Create(invitation).Error
}

func (r *InvitationRepository) GetByID(id string) (*model.Invitation, error) {
	var invitation model.Invitation
	err := r.db.First(&invitation, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &invitation, nil
}

func (r *InvitationRepository) GetByCode(code string) (*model.Invitation, error) {
	var invitation model.Invitation
	err := r.db.First(&invitation, "code = ?", code).Error
	if err != nil {
		return nil, err
	}
	return &invitation, nil
}

func (r *InvitationRepository) GetByEmail(email string) ([]*model.Invitation, error) {
	var invitations []*model.Invitation
	err := r.db.Where("email = ?", email).Find(&invitations).Error
	if err != nil {
		return nil, err
	}
	return invitations, nil
}

func (r *InvitationRepository) Update(invitation *model.Invitation) error {
	return r.db.Save(invitation).Error
}

func (r *InvitationRepository) Delete(id string) error {
	return r.db.Delete(&model.Invitation{}, "id = ?", id).Error
}

func (r *InvitationRepository) List(offset, limit int) ([]*model.Invitation, int64, error) {
	var invitations []*model.Invitation
	var total int64

	r.db.Model(&model.Invitation{}).Count(&total)
	err := r.db.Offset(offset).Limit(limit).Find(&invitations).Error
	if err != nil {
		return nil, 0, err
	}
	return invitations, total, nil
}

func (r *InvitationRepository) DeleteExpired() error {
	return r.db.Delete(&model.Invitation{}, "expires_at < datetime('now')").Error
}

func (r *InvitationRepository) GetValidByID(id string) (*model.Invitation, error) {
	var invitation model.Invitation
	err := r.db.First(&invitation, "id = ? AND expires_at > datetime('now')", id).Error
	if err != nil {
		return nil, err
	}
	if invitation.MaxUses != nil && invitation.UsedCount >= *invitation.MaxUses {
		return nil, gorm.ErrRecordNotFound
	}
	return &invitation, nil
}

func (r *InvitationRepository) IncrementUsedCount(tx *gorm.DB, id string) error {
	result := tx.Model(&model.Invitation{}).Where("id = ?", id).
		Where("expires_at > datetime('now')").
		Where("COALESCE(max_uses, 0) = 0 OR used_count < max_uses").
		Update("used_count", gorm.Expr("used_count + 1"))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
