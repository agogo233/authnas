package service

import (
	"encoding/json"
	"log"
	"time"

	"github.com/authnas/authnas/go-server/internal/model"
	"github.com/authnas/authnas/go-server/internal/repository"
	"github.com/google/uuid"
)

type AuditEventType string

const (
	AuditEventLoginSuccess         AuditEventType = "LOGIN_SUCCESS"
	AuditEventLoginFailed          AuditEventType = "LOGIN_FAILED"
	AuditEventLogout               AuditEventType = "LOGOUT"
	AuditEventPasswordReset        AuditEventType = "PASSWORD_RESET"
	AuditEventPasswordResetRequest AuditEventType = "PASSWORD_RESET_REQUEST"
	AuditEventUserCreated          AuditEventType = "USER_CREATED"
	AuditEventUserUpdated          AuditEventType = "USER_UPDATED"
	AuditEventUserDeleted          AuditEventType = "USER_DELETED"
	AuditEventAdminPasswordReset   AuditEventType = "ADMIN_PASSWORD_RESET"
	AuditEventMFAEnabled           AuditEventType = "MFA_ENABLED"
	AuditEventMFADisabled          AuditEventType = "MFA_DISABLED"
	AuditEventPasskeyRegistered    AuditEventType = "PASSKEY_REGISTERED"
	AuditEventPasskeyDeleted       AuditEventType = "PASSKEY_DELETED"
	AuditEventTokenRefreshed       AuditEventType = "TOKEN_REFRESHED"
	AuditEventTokenRevoked         AuditEventType = "TOKEN_REVOKED"
	AuditEventSessionRevoked       AuditEventType = "SESSION_REVOKED"
)

type AuditEvent struct {
	Timestamp    time.Time              `json:"timestamp"`
	EventType    AuditEventType         `json:"event_type"`
	UserID       string                 `json:"user_id,omitempty"`
	Username     string                 `json:"username,omitempty"`
	ClientID     string                 `json:"client_id,omitempty"`
	IPAddress    string                 `json:"ip_address,omitempty"`
	UserAgent    string                 `json:"user_agent,omitempty"`
	Success      bool                   `json:"success"`
	ErrorMessage string                 `json:"error_message,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

type AuditService struct {
	enabled   bool
	auditRepo *repository.AuditLogRepository
}

func NewAuditService(auditRepo *repository.AuditLogRepository) *AuditService {
	return &AuditService{
		enabled:   true,
		auditRepo: auditRepo,
	}
}

func (s *AuditService) Log(event AuditEvent) {
	if !s.enabled {
		return
	}

	event.Timestamp = time.Now()

	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("[AUDIT ERROR] failed to marshal audit event: %v", err)
		return
	}

	log.Printf("[AUDIT] %s", string(data))

	if s.auditRepo != nil {
		metadataJSON := ""
		if len(event.Metadata) > 0 {
			metaBytes, _ := json.Marshal(event.Metadata)
			metadataJSON = string(metaBytes)
		}

		auditLog := &model.AuditLog{
			ID:           uuid.New().String(),
			Timestamp:    event.Timestamp,
			EventType:    string(event.EventType),
			UserID:       event.UserID,
			Username:     event.Username,
			ClientID:     event.ClientID,
			IPAddress:    event.IPAddress,
			UserAgent:    event.UserAgent,
			Success:      event.Success,
			ErrorMessage: event.ErrorMessage,
			Metadata:     metadataJSON,
		}
		if err := s.auditRepo.Create(auditLog); err != nil {
			log.Printf("[AUDIT ERROR] failed to persist audit event: %v", err)
		}
	}
}

func (s *AuditService) LogLoginSuccess(userID, username, ipAddress, userAgent string) {
	s.Log(AuditEvent{
		EventType: AuditEventLoginSuccess,
		UserID:    userID,
		Username:  username,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Success:   true,
	})
}

func (s *AuditService) LogLoginFailed(userID, username, ipAddress, userAgent string, reason string) {
	s.Log(AuditEvent{
		EventType:    AuditEventLoginFailed,
		UserID:       userID,
		Username:     username,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		Success:      false,
		ErrorMessage: reason,
	})
}

func (s *AuditService) LogLogout(userID, username, ipAddress, userAgent string) {
	s.Log(AuditEvent{
		EventType: AuditEventLogout,
		UserID:    userID,
		Username:  username,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Success:   true,
	})
}

func (s *AuditService) LogPasswordReset(userID, username, ipAddress string, success bool, reason string) {
	s.Log(AuditEvent{
		EventType:    AuditEventPasswordReset,
		UserID:       userID,
		Username:     username,
		IPAddress:    ipAddress,
		Success:      success,
		ErrorMessage: reason,
	})
}

func (s *AuditService) LogPasswordResetRequest(email, ipAddress string) {
	s.Log(AuditEvent{
		EventType: AuditEventPasswordResetRequest,
		Username:  email,
		IPAddress: ipAddress,
		Success:   true,
	})
}

func (s *AuditService) LogUserCreated(adminID, targetUserID, targetUsername string) {
	s.Log(AuditEvent{
		EventType: AuditEventUserCreated,
		UserID:    adminID,
		Username:  targetUsername,
		Metadata: map[string]interface{}{
			"target_user_id": targetUserID,
		},
		Success: true,
	})
}

func (s *AuditService) LogUserUpdated(adminID, targetUserID string) {
	s.Log(AuditEvent{
		EventType: AuditEventUserUpdated,
		UserID:    adminID,
		Metadata: map[string]interface{}{
			"target_user_id": targetUserID,
		},
		Success: true,
	})
}

func (s *AuditService) LogUserDeleted(adminID, targetUserID, targetUsername string) {
	s.Log(AuditEvent{
		EventType: AuditEventUserDeleted,
		UserID:    adminID,
		Username:  targetUsername,
		Metadata: map[string]interface{}{
			"target_user_id": targetUserID,
		},
		Success: true,
	})
}

func (s *AuditService) LogAdminPasswordReset(adminID, targetUserID, targetUsername string) {
	s.Log(AuditEvent{
		EventType: AuditEventAdminPasswordReset,
		UserID:    adminID,
		Metadata: map[string]interface{}{
			"target_user_id":  targetUserID,
			"target_username": targetUsername,
		},
		Success: true,
	})
}

func (s *AuditService) LogMFAEnabled(userID, username string) {
	s.Log(AuditEvent{
		EventType: AuditEventMFAEnabled,
		UserID:    userID,
		Username:  username,
		Success:   true,
	})
}

func (s *AuditService) LogMFADisabled(userID, username string) {
	s.Log(AuditEvent{
		EventType: AuditEventMFADisabled,
		UserID:    userID,
		Username:  username,
		Success:   true,
	})
}

func (s *AuditService) LogPasskeyRegistered(userID, username, credentialID string) {
	s.Log(AuditEvent{
		EventType: AuditEventPasskeyRegistered,
		UserID:    userID,
		Username:  username,
		Metadata: map[string]interface{}{
			"credential_id": credentialID,
		},
		Success: true,
	})
}

func (s *AuditService) LogPasskeyDeleted(userID, username, credentialID string) {
	s.Log(AuditEvent{
		EventType: AuditEventPasskeyDeleted,
		UserID:    userID,
		Username:  username,
		Metadata: map[string]interface{}{
			"credential_id": credentialID,
		},
		Success: true,
	})
}

func (s *AuditService) LogTokenRefreshed(userID, clientID string) {
	s.Log(AuditEvent{
		EventType: AuditEventTokenRefreshed,
		UserID:    userID,
		ClientID:  clientID,
		Success:   true,
	})
}

func (s *AuditService) LogTokenRevoked(userID string) {
	s.Log(AuditEvent{
		EventType: AuditEventTokenRevoked,
		UserID:    userID,
		Success:   true,
	})
}

func (s *AuditService) LogSessionRevoked(userID, sessionID string) {
	s.Log(AuditEvent{
		EventType: AuditEventSessionRevoked,
		UserID:    userID,
		Metadata: map[string]interface{}{
			"session_id": sessionID,
		},
		Success: true,
	})
}
