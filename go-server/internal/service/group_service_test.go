package service

import (
	"testing"

	"github.com/authnas/authnas/go-server/internal/config"
	"github.com/authnas/authnas/go-server/internal/database"
	"github.com/authnas/authnas/go-server/internal/model"
	"github.com/authnas/authnas/go-server/internal/repository"
	"gorm.io/gorm"
)

func setupGroupServiceTestDB(t *testing.T) (*gorm.DB, func()) {
	tmpDir := t.TempDir()
	dbPath := tmpDir + "/test.db"

	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Path: dbPath,
		},
	}

	db, err := database.New(cfg)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	if err := db.AutoMigrate(&model.Group{}, &model.User{}, &model.UserGroup{}); err != nil {
		t.Fatalf("Failed to auto migrate: %v", err)
	}

	cleanup := func() {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
	}

	return db, cleanup
}

func newTestGroupService(db *gorm.DB) *GroupService {
	cfg := &config.Config{}
	groupRepo := repository.NewGroupRepository(db)
	userRepo := repository.NewUserRepository(db)
	return NewGroupService(cfg, groupRepo, userRepo)
}

func TestGroupService_Create(t *testing.T) {
	db, cleanup := setupGroupServiceTestDB(t)
	defer cleanup()

	svc := newTestGroupService(db)

	group, err := svc.Create("Test Group", "A test group description")
	if err != nil {
		t.Fatalf("Failed to create group: %v", err)
	}

	if group.Name != "Test Group" {
		t.Errorf("Expected name 'Test Group', got '%s'", group.Name)
	}

	if group.Description == nil || *group.Description != "A test group description" {
		t.Error("Description mismatch")
	}

	if group.ID == "" {
		t.Error("Group ID should be set")
	}
}

func TestGroupService_GetByID(t *testing.T) {
	db, cleanup := setupGroupServiceTestDB(t)
	defer cleanup()

	svc := newTestGroupService(db)

	created, _ := svc.Create("Test Group", "description")

	found, err := svc.GetByID(created.ID)
	if err != nil {
		t.Fatalf("Failed to get group by ID: %v", err)
	}

	if found.ID != created.ID {
		t.Errorf("Expected ID '%s', got '%s'", created.ID, found.ID)
	}
}

func TestGroupService_GetByID_NotFound(t *testing.T) {
	db, cleanup := setupGroupServiceTestDB(t)
	defer cleanup()

	svc := newTestGroupService(db)

	_, err := svc.GetByID("non-existent")
	if err == nil {
		t.Error("Expected error for non-existent group")
	}
}

func TestGroupService_List(t *testing.T) {
	db, cleanup := setupGroupServiceTestDB(t)
	defer cleanup()

	svc := newTestGroupService(db)

	svc.Create("Group A", "description A")
	svc.Create("Group B", "description B")
	svc.Create("Group C", "description C")

	groups, total, err := svc.List(0, 10)
	if err != nil {
		t.Fatalf("Failed to list groups: %v", err)
	}

	if total != 3 {
		t.Errorf("Expected total 3, got %d", total)
	}

	if len(groups) != 3 {
		t.Errorf("Expected 3 groups, got %d", len(groups))
	}
}

func TestGroupService_List_Pagination(t *testing.T) {
	db, cleanup := setupGroupServiceTestDB(t)
	defer cleanup()

	svc := newTestGroupService(db)

	for i := 0; i < 5; i++ {
		svc.Create("Group"+string(rune('A'+i)), "description")
	}

	groups, total, err := svc.List(0, 2)
	if err != nil {
		t.Fatalf("Failed to list groups: %v", err)
	}

	if total != 5 {
		t.Errorf("Expected total 5, got %d", total)
	}

	if len(groups) != 2 {
		t.Errorf("Expected 2 groups, got %d", len(groups))
	}
}

func TestGroupService_List_Empty(t *testing.T) {
	db, cleanup := setupGroupServiceTestDB(t)
	defer cleanup()

	svc := newTestGroupService(db)

	groups, total, err := svc.List(0, 10)
	if err != nil {
		t.Fatalf("Failed to list groups: %v", err)
	}

	if total != 0 {
		t.Errorf("Expected total 0, got %d", total)
	}

	if len(groups) != 0 {
		t.Errorf("Expected 0 groups, got %d", len(groups))
	}
}

func TestGroupService_Update(t *testing.T) {
	db, cleanup := setupGroupServiceTestDB(t)
	defer cleanup()

	svc := newTestGroupService(db)

	created, _ := svc.Create("Original Name", "Original Description")

	updated, err := svc.Update(created.ID, "New Name", "New Description")
	if err != nil {
		t.Fatalf("Failed to update group: %v", err)
	}

	if updated.Name != "New Name" {
		t.Errorf("Expected name 'New Name', got '%s'", updated.Name)
	}

	if updated.Description == nil || *updated.Description != "New Description" {
		t.Error("Description mismatch")
	}
}

func TestGroupService_Update_NoDescriptionChange(t *testing.T) {
	db, cleanup := setupGroupServiceTestDB(t)
	defer cleanup()

	svc := newTestGroupService(db)

	created, _ := svc.Create("Original Name", "Original Description")

	updated, err := svc.Update(created.ID, "New Name", "")
	if err != nil {
		t.Fatalf("Failed to update group: %v", err)
	}

	if updated.Name != "New Name" {
		t.Errorf("Expected name 'New Name', got '%s'", updated.Name)
	}

	if updated.Description == nil || *updated.Description != "Original Description" {
		t.Error("Description should not have changed")
	}
}

func TestGroupService_Delete(t *testing.T) {
	db, cleanup := setupGroupServiceTestDB(t)
	defer cleanup()

	svc := newTestGroupService(db)

	created, _ := svc.Create("Test Group", "description")

	err := svc.Delete(created.ID)
	if err != nil {
		t.Fatalf("Failed to delete group: %v", err)
	}

	_, err = svc.GetByID(created.ID)
	if err == nil {
		t.Error("Expected error after deleting group")
	}
}

func TestGroupService_AddUser(t *testing.T) {
	db, cleanup := setupGroupServiceTestDB(t)
	defer cleanup()

	svc := newTestGroupService(db)
	group, _ := svc.Create("Test Group", "description")

	email := "test@example.com"
	user := &model.User{
		ID:       "user-1",
		Username: "testuser",
		Email:    &email,
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	err := svc.AddUser(group.ID, user.ID)
	if err != nil {
		t.Fatalf("Failed to add user to group: %v", err)
	}
}

func TestGroupService_RemoveUser(t *testing.T) {
	db, cleanup := setupGroupServiceTestDB(t)
	defer cleanup()

	svc := newTestGroupService(db)
	group, _ := svc.Create("Test Group", "description")

	email := "test@example.com"
	user := &model.User{
		ID:       "user-1",
		Username: "testuser",
		Email:    &email,
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	svc.AddUser(group.ID, user.ID)
	err := svc.RemoveUser(group.ID, user.ID)
	if err != nil {
		t.Fatalf("Failed to remove user from group: %v", err)
	}
}

func TestGroupService_GetUserGroups(t *testing.T) {
	db, cleanup := setupGroupServiceTestDB(t)
	defer cleanup()

	svc := newTestGroupService(db)

	group1, _ := svc.Create("Group 1", "description")
	group2, _ := svc.Create("Group 2", "description")

	email := "test@example.com"
	user := &model.User{
		ID:       "user-1",
		Username: "testuser",
		Email:    &email,
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	svc.AddUser(group1.ID, user.ID)
	svc.AddUser(group2.ID, user.ID)

	groups, err := svc.GetUserGroups(user.ID)
	if err != nil {
		t.Fatalf("Failed to get user groups: %v", err)
	}

	if len(groups) != 2 {
		t.Errorf("Expected 2 groups, got %d", len(groups))
	}
}

func TestGroupService_GetUserGroups_Empty(t *testing.T) {
	db, cleanup := setupGroupServiceTestDB(t)
	defer cleanup()

	svc := newTestGroupService(db)

	email := "test@example.com"
	user := &model.User{
		ID:       "user-1",
		Username: "testuser",
		Email:    &email,
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	groups, err := svc.GetUserGroups(user.ID)
	if err != nil {
		t.Fatalf("Failed to get user groups: %v", err)
	}

	if len(groups) != 0 {
		t.Errorf("Expected 0 groups, got %d", len(groups))
	}
}
