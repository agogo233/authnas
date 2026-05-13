package repository

import (
	"testing"
	"time"

	"github.com/authnas/authnas/go-server/internal/model"
	"gorm.io/gorm"
)

func setupTestDBForInvitation(t *testing.T) *gorm.DB {
	return SetupTestDB(t)
}

func TestInvitationRepository_Create(t *testing.T) {
	db := setupTestDBForInvitation(t)
	defer CleanupDB(db)

	repo := NewInvitationRepository(db)

	invitation := &model.Invitation{
		ID:        "invitation-1",
		Email:     "test@example.com",
		Code:      "TESTCODE123",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	if err := repo.Create(invitation); err != nil {
		t.Fatalf("Failed to create invitation: %v", err)
	}

	retrieved, err := repo.GetByID("invitation-1")
	if err != nil {
		t.Fatalf("Failed to get invitation: %v", err)
	}

	if retrieved.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%s'", retrieved.Email)
	}
}

func TestInvitationRepository_GetByID(t *testing.T) {
	db := setupTestDBForInvitation(t)
	defer CleanupDB(db)

	repo := NewInvitationRepository(db)

	invitation := &model.Invitation{
		ID:        "invitation-get-by-id",
		Email:     "getbyid@example.com",
		Code:      "GETBYIDCODE",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	repo.Create(invitation)

	retrieved, err := repo.GetByID("invitation-get-by-id")
	if err != nil {
		t.Fatalf("Failed to get invitation by ID: %v", err)
	}

	if retrieved.Code != "GETBYIDCODE" {
		t.Errorf("Expected code 'GETBYIDCODE', got '%s'", retrieved.Code)
	}
}

func TestInvitationRepository_GetByCode(t *testing.T) {
	db := setupTestDBForInvitation(t)
	defer CleanupDB(db)

	repo := NewInvitationRepository(db)

	invitation := &model.Invitation{
		ID:        "invitation-get-by-code",
		Email:     "getbycode@example.com",
		Code:      "UNIQUECODE456",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	repo.Create(invitation)

	retrieved, err := repo.GetByCode("UNIQUECODE456")
	if err != nil {
		t.Fatalf("Failed to get invitation by code: %v", err)
	}

	if retrieved.ID != "invitation-get-by-code" {
		t.Errorf("Expected ID 'invitation-get-by-code', got '%s'", retrieved.ID)
	}
}

func TestInvitationRepository_GetByCode_NotFound(t *testing.T) {
	db := setupTestDBForInvitation(t)
	defer CleanupDB(db)

	repo := NewInvitationRepository(db)

	_, err := repo.GetByCode("NONEXISTENTCODE")
	if err == nil {
		t.Error("Expected error when getting invitation with non-existent code")
	}
}

func TestInvitationRepository_GetByEmail(t *testing.T) {
	db := setupTestDBForInvitation(t)
	defer CleanupDB(db)

	repo := NewInvitationRepository(db)

	invitation1 := &model.Invitation{
		ID:        "invitation-email-1",
		Email:     "multiple@example.com",
		Code:      "CODETWO",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	invitation2 := &model.Invitation{
		ID:        "invitation-email-2",
		Email:     "multiple@example.com",
		Code:      "CODETHREE",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	repo.Create(invitation1)
	repo.Create(invitation2)

	retrieved, err := repo.GetByEmail("multiple@example.com")
	if err != nil {
		t.Fatalf("Failed to get invitations by email: %v", err)
	}

	if len(retrieved) != 2 {
		t.Errorf("Expected 2 invitations, got %d", len(retrieved))
	}
}

func TestInvitationRepository_Update(t *testing.T) {
	db := setupTestDBForInvitation(t)
	defer CleanupDB(db)

	repo := NewInvitationRepository(db)

	invitation := &model.Invitation{
		ID:        "invitation-update",
		Email:     "update@example.com",
		Code:      "UPDATECODE",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	repo.Create(invitation)

	invitation.UsedCount = 1
	if err := repo.Update(invitation); err != nil {
		t.Fatalf("Failed to update invitation: %v", err)
	}

	retrieved, _ := repo.GetByID("invitation-update")
	if retrieved.UsedCount != 1 {
		t.Error("Expected used_count to be 1 after update")
	}
}

func TestInvitationRepository_Delete(t *testing.T) {
	db := setupTestDBForInvitation(t)
	defer CleanupDB(db)

	repo := NewInvitationRepository(db)

	invitation := &model.Invitation{
		ID:        "invitation-delete",
		Email:     "delete@example.com",
		Code:      "DELETECODE",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	repo.Create(invitation)

	if err := repo.Delete("invitation-delete"); err != nil {
		t.Fatalf("Failed to delete invitation: %v", err)
	}

	_, err := repo.GetByID("invitation-delete")
	if err == nil {
		t.Error("Expected error when getting deleted invitation")
	}
}

func TestInvitationRepository_List(t *testing.T) {
	db := setupTestDBForInvitation(t)
	defer CleanupDB(db)

	repo := NewInvitationRepository(db)

	for i := 0; i < 5; i++ {
		invitation := &model.Invitation{
			ID:        "invitation-list-" + string(rune('a'+i)),
			Email:     "list" + string(rune('a'+i)) + "@example.com",
			Code:      "LISTCODE" + string(rune('a'+i)),
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}
		repo.Create(invitation)
	}

	invitations, total, err := repo.List(0, 10)
	if err != nil {
		t.Fatalf("Failed to list invitations: %v", err)
	}

	if total != 5 {
		t.Errorf("Expected total 5, got %d", total)
	}

	if len(invitations) != 5 {
		t.Errorf("Expected 5 invitations, got %d", len(invitations))
	}
}

func TestInvitationRepository_DeleteExpired(t *testing.T) {
	db := setupTestDBForInvitation(t)
	defer CleanupDB(db)

	repo := NewInvitationRepository(db)

	expiredInvitation := &model.Invitation{
		ID:        "invitation-expired",
		Email:     "expired@example.com",
		Code:      "EXPIREDCODE",
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}
	validInvitation := &model.Invitation{
		ID:        "invitation-valid",
		Email:     "valid@example.com",
		Code:      "VALIDCODE",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	repo.Create(expiredInvitation)
	repo.Create(validInvitation)

	if err := repo.DeleteExpired(); err != nil {
		t.Fatalf("Failed to delete expired invitations: %v", err)
	}

	_, err := repo.GetByID("invitation-expired")
	if err == nil {
		t.Error("Expected expired invitation to be deleted")
	}

	_, err = repo.GetByID("invitation-valid")
	if err != nil {
		t.Error("Expected valid invitation to still exist")
	}
}
