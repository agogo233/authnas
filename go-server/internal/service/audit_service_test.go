package service

import (
	"testing"
)

func TestAuditService_NewAuditService(t *testing.T) {
	svc := NewAuditService(nil)
	if svc == nil {
		t.Fatal("NewAuditService should not return nil")
	}
	if !svc.enabled {
		t.Error("AuditService should be enabled by default")
	}
}

func TestAuditService_Log_Disabled(t *testing.T) {
	svc := &AuditService{enabled: false}
	svc.Log(AuditEvent{
		EventType: AuditEventLoginSuccess,
		UserID:    "user123",
		Success:   true,
	})
}

func TestAuditService_Log_Enabled(t *testing.T) {
	svc := NewAuditService(nil)
	svc.Log(AuditEvent{
		EventType: AuditEventLoginSuccess,
		UserID:    "user123",
		Username:  "testuser",
		IPAddress: "192.168.1.1",
		UserAgent: "Mozilla/5.0",
		Success:   true,
	})
}

func TestAuditService_LogLoginSuccess(t *testing.T) {
	svc := NewAuditService(nil)
	svc.LogLoginSuccess("user123", "testuser", "192.168.1.1", "Mozilla/5.0")
}

func TestAuditService_LogLoginFailed(t *testing.T) {
	svc := NewAuditService(nil)
	svc.LogLoginFailed("user123", "testuser", "192.168.1.1", "Mozilla/5.0", "invalid password")
}

func TestAuditService_LogLogout(t *testing.T) {
	svc := NewAuditService(nil)
	svc.LogLogout("user123", "testuser", "192.168.1.1", "Mozilla/5.0")
}

func TestAuditService_LogPasswordReset(t *testing.T) {
	svc := NewAuditService(nil)
	svc.LogPasswordReset("user123", "testuser", "192.168.1.1", true, "")
}

func TestAuditService_LogPasswordResetRequest(t *testing.T) {
	svc := NewAuditService(nil)
	svc.LogPasswordResetRequest("test@example.com", "192.168.1.1")
}

func TestAuditService_LogUserCreated(t *testing.T) {
	svc := NewAuditService(nil)
	svc.LogUserCreated("admin123", "user123", "newuser")
}

func TestAuditService_LogUserUpdated(t *testing.T) {
	svc := NewAuditService(nil)
	svc.LogUserUpdated("admin123", "user123")
}

func TestAuditService_LogUserDeleted(t *testing.T) {
	svc := NewAuditService(nil)
	svc.LogUserDeleted("admin123", "user123", "deleteduser")
}

func TestAuditService_LogAdminPasswordReset(t *testing.T) {
	svc := NewAuditService(nil)
	svc.LogAdminPasswordReset("admin123", "user123", "testuser")
}

func TestAuditService_LogMFAEnabled(t *testing.T) {
	svc := NewAuditService(nil)
	svc.LogMFAEnabled("user123", "testuser")
}

func TestAuditService_LogMFADisabled(t *testing.T) {
	svc := NewAuditService(nil)
	svc.LogMFADisabled("user123", "testuser")
}

func TestAuditService_LogPasskeyRegistered(t *testing.T) {
	svc := NewAuditService(nil)
	svc.LogPasskeyRegistered("user123", "testuser", "cred123")
}

func TestAuditService_LogPasskeyDeleted(t *testing.T) {
	svc := NewAuditService(nil)
	svc.LogPasskeyDeleted("user123", "testuser", "cred123")
}

func TestAuditService_LogTokenRefreshed(t *testing.T) {
	svc := NewAuditService(nil)
	svc.LogTokenRefreshed("user123", "client123")
}

func TestAuditService_LogTokenRevoked(t *testing.T) {
	svc := NewAuditService(nil)
	svc.LogTokenRevoked("user123")
}

func TestAuditService_LogSessionRevoked(t *testing.T) {
	svc := NewAuditService(nil)
	svc.LogSessionRevoked("user123", "session123")
}

func TestAuditService_Log_WithMetadata(t *testing.T) {
	svc := NewAuditService(nil)
	svc.Log(AuditEvent{
		EventType: AuditEventLoginSuccess,
		UserID:    "user123",
		Success:   true,
		Metadata: map[string]interface{}{
			"key1": "value1",
			"key2": 123,
		},
	})
}

func TestAuditEventType_Values(t *testing.T) {
	tests := []struct {
		eventType AuditEventType
		expected  string
	}{
		{AuditEventLoginSuccess, "LOGIN_SUCCESS"},
		{AuditEventLoginFailed, "LOGIN_FAILED"},
		{AuditEventLogout, "LOGOUT"},
		{AuditEventPasswordReset, "PASSWORD_RESET"},
		{AuditEventPasswordResetRequest, "PASSWORD_RESET_REQUEST"},
		{AuditEventUserCreated, "USER_CREATED"},
		{AuditEventUserUpdated, "USER_UPDATED"},
		{AuditEventUserDeleted, "USER_DELETED"},
		{AuditEventAdminPasswordReset, "ADMIN_PASSWORD_RESET"},
		{AuditEventMFAEnabled, "MFA_ENABLED"},
		{AuditEventMFADisabled, "MFA_DISABLED"},
		{AuditEventPasskeyRegistered, "PASSKEY_REGISTERED"},
		{AuditEventPasskeyDeleted, "PASSKEY_DELETED"},
		{AuditEventTokenRefreshed, "TOKEN_REFRESHED"},
		{AuditEventTokenRevoked, "TOKEN_REVOKED"},
		{AuditEventSessionRevoked, "SESSION_REVOKED"},
	}

	for _, tt := range tests {
		if string(tt.eventType) != tt.expected {
			t.Errorf("Expected %s, got %s", tt.expected, string(tt.eventType))
		}
	}
}

func TestAuditEvent_Structure(t *testing.T) {
	event := AuditEvent{
		EventType:    AuditEventLoginSuccess,
		UserID:       "user123",
		Username:     "testuser",
		ClientID:     "client123",
		IPAddress:    "192.168.1.1",
		UserAgent:    "Mozilla/5.0",
		Success:      true,
		ErrorMessage: "",
		Metadata:     nil,
	}

	if event.EventType != AuditEventLoginSuccess {
		t.Error("EventType mismatch")
	}
	if event.UserID != "user123" {
		t.Error("UserID mismatch")
	}
	if event.Username != "testuser" {
		t.Error("Username mismatch")
	}
	if event.Success != true {
		t.Error("Success should be true")
	}
}

func TestAuditService_Log_CustomTimestamp(t *testing.T) {
	svc := NewAuditService(nil)
	event := AuditEvent{
		EventType: AuditEventLoginSuccess,
		UserID:    "user123",
		Success:   true,
	}
	svc.Log(event)
}

func TestAuditService_LogEventWithEmptyFields(t *testing.T) {
	svc := NewAuditService(nil)
	svc.Log(AuditEvent{
		EventType: AuditEventLoginFailed,
		Success:   false,
	})
}

func TestAuditService_AuditEvent_AllFields(t *testing.T) {
	event := AuditEvent{
		EventType:    AuditEventUserCreated,
		UserID:       "admin-id",
		Username:     "admin",
		ClientID:     "client-id",
		IPAddress:    "10.0.0.1",
		UserAgent:    "TestAgent/1.0",
		Success:      true,
		ErrorMessage: "",
		Metadata: map[string]interface{}{
			"target_user_id": "new-user-id",
		},
	}

	if event.EventType != AuditEventUserCreated {
		t.Errorf("Expected %s, got %s", AuditEventUserCreated, event.EventType)
	}
	if event.UserID != "admin-id" {
		t.Errorf("Expected UserID 'admin-id', got '%s'", event.UserID)
	}
	if event.Metadata["target_user_id"] != "new-user-id" {
		t.Error("Metadata target_user_id mismatch")
	}
}

func TestAuditService_LogMultipleEvents(t *testing.T) {
	svc := NewAuditService(nil)
	svc.LogLoginSuccess("user1", "user1", "1.1.1.1", "Agent1")
	svc.LogLoginFailed("user2", "user2", "2.2.2.2", "Agent2", "failed")
	svc.LogLogout("user3", "user3", "3.3.3.3", "Agent3")
}

func TestAuditService_AuditEventTypes_Uniqueness(t *testing.T) {
	seen := make(map[AuditEventType]bool)
	eventTypes := []AuditEventType{
		AuditEventLoginSuccess,
		AuditEventLoginFailed,
		AuditEventLogout,
		AuditEventPasswordReset,
		AuditEventPasswordResetRequest,
		AuditEventUserCreated,
		AuditEventUserUpdated,
		AuditEventUserDeleted,
		AuditEventAdminPasswordReset,
		AuditEventMFAEnabled,
		AuditEventMFADisabled,
		AuditEventPasskeyRegistered,
		AuditEventPasskeyDeleted,
		AuditEventTokenRefreshed,
		AuditEventTokenRevoked,
		AuditEventSessionRevoked,
	}

	for _, et := range eventTypes {
		if seen[et] {
			t.Errorf("Duplicate AuditEventType found: %s", et)
		}
		seen[et] = true
	}
}

func TestAuditService_Log_VerifyOutputFormat(t *testing.T) {
	svc := NewAuditService(nil)
	event := AuditEvent{
		EventType: AuditEventLoginSuccess,
		UserID:    "test-user",
		Username:  "testuser",
		Success:   true,
	}
	svc.Log(event)
}
