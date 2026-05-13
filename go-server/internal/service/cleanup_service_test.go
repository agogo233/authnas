package service

import (
	"testing"
	"time"
)

func newTestCleanupServiceNilRepos() *CleanupService {
	return NewCleanupService(
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)
}

func TestCleanupService_NewCleanupService(t *testing.T) {
	svc := newTestCleanupServiceNilRepos()
	if svc == nil {
		t.Fatal("NewCleanupService should not return nil")
	}
}

func TestCleanupService_CleanupExpired_NilRepos(t *testing.T) {
	svc := newTestCleanupServiceNilRepos()

	result, err := svc.CleanupExpired()
	if err != nil {
		t.Fatalf("CleanupExpired should not return error: %v", err)
	}

	if result.KeysDeleted != 0 {
		t.Errorf("Expected 0 KeysDeleted, got %d", result.KeysDeleted)
	}
	if result.PasskeyAuthOptionsDeleted != 0 {
		t.Errorf("Expected 0 PasskeyAuthOptionsDeleted, got %d", result.PasskeyAuthOptionsDeleted)
	}
	if result.OIDCPayloadsDeleted != 0 {
		t.Errorf("Expected 0 OIDCPayloadsDeleted, got %d", result.OIDCPayloadsDeleted)
	}
	if result.EmailVerificationsDeleted != 0 {
		t.Errorf("Expected 0 EmailVerificationsDeleted, got %d", result.EmailVerificationsDeleted)
	}
	if result.PasswordResetsDeleted != 0 {
		t.Errorf("Expected 0 PasswordResetsDeleted, got %d", result.PasswordResetsDeleted)
	}
	if result.EmailLogsDeleted != 0 {
		t.Errorf("Expected 0 EmailLogsDeleted, got %d", result.EmailLogsDeleted)
	}
	if result.CSRFTokensCleaned != 0 {
		t.Errorf("Expected 0 CSRFTokensCleaned, got %d", result.CSRFTokensCleaned)
	}
}

func TestCleanupService_CleanupOldEmailLogs_NilRepo(t *testing.T) {
	svc := newTestCleanupServiceNilRepos()

	deleted, err := svc.CleanupOldEmailLogs(30)
	if err != nil {
		t.Fatalf("CleanupOldEmailLogs should not return error: %v", err)
	}

	if deleted != 0 {
		t.Errorf("Expected 0 deleted, got %d", deleted)
	}
}

func TestCleanupService_StartCleanupScheduler_NilRepos(t *testing.T) {
	svc := newTestCleanupServiceNilRepos()

	svc.StartCleanupScheduler(time.Second)
	svc.Stop()
}

func TestCleanupService_Stop(t *testing.T) {
	svc := newTestCleanupServiceNilRepos()

	svc.Stop()
}

func TestCleanupResult_Structure(t *testing.T) {
	result := &CleanupResult{
		KeysDeleted:               10,
		PasskeyAuthOptionsDeleted: 5,
		OIDCPayloadsDeleted:       3,
		EmailVerificationsDeleted: 7,
		PasswordResetsDeleted:     2,
		EmailLogsDeleted:          15,
		CSRFTokensCleaned:         20,
	}

	if result.KeysDeleted != 10 {
		t.Errorf("Expected KeysDeleted 10, got %d", result.KeysDeleted)
	}
	if result.PasskeyAuthOptionsDeleted != 5 {
		t.Errorf("Expected PasskeyAuthOptionsDeleted 5, got %d", result.PasskeyAuthOptionsDeleted)
	}
	if result.OIDCPayloadsDeleted != 3 {
		t.Errorf("Expected OIDCPayloadsDeleted 3, got %d", result.OIDCPayloadsDeleted)
	}
	if result.EmailVerificationsDeleted != 7 {
		t.Errorf("Expected EmailVerificationsDeleted 7, got %d", result.EmailVerificationsDeleted)
	}
	if result.PasswordResetsDeleted != 2 {
		t.Errorf("Expected PasswordResetsDeleted 2, got %d", result.PasswordResetsDeleted)
	}
	if result.EmailLogsDeleted != 15 {
		t.Errorf("Expected EmailLogsDeleted 15, got %d", result.EmailLogsDeleted)
	}
	if result.CSRFTokensCleaned != 20 {
		t.Errorf("Expected CSRFTokensCleaned 20, got %d", result.CSRFTokensCleaned)
	}
}

func TestCleanupService_CleanupExpired_AllReposNil(t *testing.T) {
	svc := NewCleanupService(nil, nil, nil, nil, nil, nil, nil)

	result, err := svc.CleanupExpired()
	if err != nil {
		t.Fatalf("CleanupExpired should not return error: %v", err)
	}

	if result == nil {
		t.Fatal("Result should not be nil")
	}
}
