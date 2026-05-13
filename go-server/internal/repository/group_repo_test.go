package repository

import (
	"testing"

	"github.com/authnas/authnas/go-server/internal/model"
	"gorm.io/gorm"
)

func setupTestDBForGroup(t *testing.T) *gorm.DB {
	return SetupTestDB(t)
}

func TestGroupRepository_Create(t *testing.T) {
	db := setupTestDBForGroup(t)
	defer CleanupDB(db)

	repo := NewGroupRepository(db)

	group := &model.Group{
		ID:   "group-1",
		Name: "Test Group",
	}

	if err := repo.Create(group); err != nil {
		t.Fatalf("Failed to create group: %v", err)
	}

	retrieved, err := repo.GetByID("group-1")
	if err != nil {
		t.Fatalf("Failed to get group: %v", err)
	}

	if retrieved.Name != "Test Group" {
		t.Errorf("Expected name 'Test Group', got '%s'", retrieved.Name)
	}
}

func TestGroupRepository_GetByID(t *testing.T) {
	db := setupTestDBForGroup(t)
	defer CleanupDB(db)

	repo := NewGroupRepository(db)

	group := &model.Group{
		ID:   "group-get-by-id",
		Name: "Get By ID Group",
	}

	repo.Create(group)

	retrieved, err := repo.GetByID("group-get-by-id")
	if err != nil {
		t.Fatalf("Failed to get group by ID: %v", err)
	}

	if retrieved.Name != "Get By ID Group" {
		t.Errorf("Expected name 'Get By ID Group', got '%s'", retrieved.Name)
	}
}

func TestGroupRepository_GetByName(t *testing.T) {
	db := setupTestDBForGroup(t)
	defer CleanupDB(db)

	repo := NewGroupRepository(db)

	group := &model.Group{
		ID:   "group-get-by-name",
		Name: "Unique Group Name",
	}

	repo.Create(group)

	retrieved, err := repo.GetByName("Unique Group Name")
	if err != nil {
		t.Fatalf("Failed to get group by name: %v", err)
	}

	if retrieved.ID != "group-get-by-name" {
		t.Errorf("Expected ID 'group-get-by-name', got '%s'", retrieved.ID)
	}
}

func TestGroupRepository_GetByName_NotFound(t *testing.T) {
	db := setupTestDBForGroup(t)
	defer CleanupDB(db)

	repo := NewGroupRepository(db)

	_, err := repo.GetByName("Non-existent Group")
	if err == nil {
		t.Error("Expected error when getting non-existent group")
	}
}

func TestGroupRepository_Update(t *testing.T) {
	db := setupTestDBForGroup(t)
	defer CleanupDB(db)

	repo := NewGroupRepository(db)

	group := &model.Group{
		ID:          "group-update",
		Name:        "Original Name",
		Description: StringPtr("Original Description"),
	}

	repo.Create(group)

	group.Name = "Updated Name"
	group.Description = StringPtr("Updated Description")
	if err := repo.Update(group); err != nil {
		t.Fatalf("Failed to update group: %v", err)
	}

	retrieved, _ := repo.GetByID("group-update")
	if retrieved.Name != "Updated Name" {
		t.Errorf("Expected name 'Updated Name', got '%s'", retrieved.Name)
	}
	if retrieved.Description == nil || *retrieved.Description != "Updated Description" {
		t.Errorf("Expected description 'Updated Description', got '%v'", retrieved.Description)
	}
}

func TestGroupRepository_Delete(t *testing.T) {
	db := setupTestDBForGroup(t)
	defer CleanupDB(db)

	repo := NewGroupRepository(db)

	group := &model.Group{
		ID:   "group-delete",
		Name: "Delete Group",
	}

	repo.Create(group)

	if err := repo.Delete("group-delete"); err != nil {
		t.Fatalf("Failed to delete group: %v", err)
	}

	_, err := repo.GetByID("group-delete")
	if err == nil {
		t.Error("Expected error when getting deleted group")
	}
}

func TestGroupRepository_List(t *testing.T) {
	db := setupTestDBForGroup(t)
	defer CleanupDB(db)

	repo := NewGroupRepository(db)

	for i := 0; i < 5; i++ {
		group := &model.Group{
			ID:   "group-list-" + string(rune('a'+i)),
			Name: "List Group " + string(rune('a'+i)),
		}
		repo.Create(group)
	}

	groups, total, err := repo.List(0, 10)
	if err != nil {
		t.Fatalf("Failed to list groups: %v", err)
	}

	if total != 5 {
		t.Errorf("Expected total 5, got %d", total)
	}

	if len(groups) != 5 {
		t.Errorf("Expected 5 groups, got %d", len(groups))
	}
}

func TestGroupRepository_AddUser(t *testing.T) {
	db := setupTestDBForGroup(t)
	defer CleanupDB(db)

	groupRepo := NewGroupRepository(db)
	userRepo := NewUserRepository(db)

	group := &model.Group{
		ID:   "group-add-user",
		Name: "Add User Group",
	}
	groupRepo.Create(group)

	user := &model.User{
		ID:       "user-add-to-group",
		Username: "adduser",
		Email:    StringPtr("adduser@example.com"),
	}
	userRepo.Create(user)

	if err := groupRepo.AddUser("group-add-user", "user-add-to-group"); err != nil {
		t.Fatalf("Failed to add user to group: %v", err)
	}

	groups, err := groupRepo.GetUserGroups("user-add-to-group")
	if err != nil {
		t.Fatalf("Failed to get user groups: %v", err)
	}

	if len(groups) != 1 {
		t.Errorf("Expected 1 group, got %d", len(groups))
	}

	if groups[0].ID != "group-add-user" {
		t.Errorf("Expected group ID 'group-add-user', got '%s'", groups[0].ID)
	}
}

func TestGroupRepository_RemoveUser(t *testing.T) {
	db := setupTestDBForGroup(t)
	defer CleanupDB(db)

	groupRepo := NewGroupRepository(db)
	userRepo := NewUserRepository(db)

	group := &model.Group{
		ID:   "group-remove-user",
		Name: "Remove User Group",
	}
	groupRepo.Create(group)

	user := &model.User{
		ID:       "user-remove-from-group",
		Username: "removeuser",
		Email:    StringPtr("removeuser@example.com"),
	}
	userRepo.Create(user)

	groupRepo.AddUser("group-remove-user", "user-remove-from-group")

	if err := groupRepo.RemoveUser("group-remove-user", "user-remove-from-group"); err != nil {
		t.Fatalf("Failed to remove user from group: %v", err)
	}

	groups, err := groupRepo.GetUserGroups("user-remove-from-group")
	if err != nil {
		t.Fatalf("Failed to get user groups: %v", err)
	}

	if len(groups) != 0 {
		t.Errorf("Expected 0 groups, got %d", len(groups))
	}
}

func TestGroupRepository_GetUserGroups(t *testing.T) {
	db := setupTestDBForGroup(t)
	defer CleanupDB(db)

	groupRepo := NewGroupRepository(db)
	userRepo := NewUserRepository(db)

	user := &model.User{
		ID:       "user-get-groups",
		Username: "getgroupsuser",
		Email:    StringPtr("getgroups@example.com"),
	}
	userRepo.Create(user)

	for i := 0; i < 3; i++ {
		group := &model.Group{
			ID:   "group-for-user-" + string(rune('a'+i)),
			Name: "User Group " + string(rune('a'+i)),
		}
		groupRepo.Create(group)
		groupRepo.AddUser(group.ID, user.ID)
	}

	groups, err := groupRepo.GetUserGroups(user.ID)
	if err != nil {
		t.Fatalf("Failed to get user groups: %v", err)
	}

	if len(groups) != 3 {
		t.Errorf("Expected 3 groups, got %d", len(groups))
	}
}
