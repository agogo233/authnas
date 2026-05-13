package model

import (
	"testing"
	"time"
)

func TestUserTableName(t *testing.T) {
	user := User{}
	if user.TableName() != "user" {
		t.Errorf("Expected table name 'user', got '%s'", user.TableName())
	}
}

func TestGroupTableName(t *testing.T) {
	group := Group{}
	if group.TableName() != "groups" {
		t.Errorf("Expected table name 'groups', got '%s'", group.TableName())
	}
}

func TestTOTPTableName(t *testing.T) {
	totp := TOTP{}
	if totp.TableName() != "totp" {
		t.Errorf("Expected table name 'totp', got '%s'", totp.TableName())
	}
}

func TestPasskeyTableName(t *testing.T) {
	passkey := Passkey{}
	if passkey.TableName() != "passkey" {
		t.Errorf("Expected table name 'passkey', got '%s'", passkey.TableName())
	}
}

func TestPasskeyAuthOptionsTableName(t *testing.T) {
	opts := PasskeyAuthOptions{}
	if opts.TableName() != "passkey_auth_options" {
		t.Errorf("Expected table name 'passkey_auth_options', got '%s'", opts.TableName())
	}
}

func TestClientTableName(t *testing.T) {
	client := Client{}
	if client.TableName() != "client" {
		t.Errorf("Expected table name 'client', got '%s'", client.TableName())
	}
}

func TestConsentTableName(t *testing.T) {
	consent := Consent{}
	if consent.TableName() != "consent" {
		t.Errorf("Expected table name 'consent', got '%s'", consent.TableName())
	}
}

func TestInvitationTableName(t *testing.T) {
	invitation := Invitation{}
	if invitation.TableName() != "invitation" {
		t.Errorf("Expected table name 'invitation', got '%s'", invitation.TableName())
	}
}

func TestKeyTableName(t *testing.T) {
	key := Key{}
	if key.TableName() != "key" {
		t.Errorf("Expected table name 'key', got '%s'", key.TableName())
	}
}

func TestUserGroupTableName(t *testing.T) {
	ug := UserGroup{}
	if ug.TableName() != "user_group" {
		t.Errorf("Expected table name 'user_group', got '%s'", ug.TableName())
	}
}

func TestUserModel(t *testing.T) {
	email := "test@example.com"
	username := "testuser"
	name := "Test User"
	now := time.Now()

	user := User{
		ID:            "test-id-123",
		Email:         &email,
		Username:      username,
		Name:          &name,
		PasswordHash:  nil,
		EmailVerified: true,
		Approved:      true,
		MFARequired:   false,
		TokenVersion:  0,
		ExpiresAt:     nil,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if user.ID != "test-id-123" {
		t.Errorf("Expected ID 'test-id-123', got '%s'", user.ID)
	}

	if user.Email == nil || *user.Email != email {
		t.Errorf("Expected Email '%s', got '%v'", email, user.Email)
	}

	if user.Username != username {
		t.Errorf("Expected Username '%s', got '%s'", username, user.Username)
	}

	if !user.EmailVerified {
		t.Error("Expected EmailVerified to be true")
	}

	if !user.Approved {
		t.Error("Expected Approved to be true")
	}
}

func TestTOTPModel(t *testing.T) {
	now := time.Now()
	totp := TOTP{
		ID:        "totp-id-123",
		UserID:    "user-id-123",
		Secret:    "JBSWY3DPEHPK3PXP",
		Issuer:    "AuthNas",
		CreatedAt: now,
		UpdatedAt: now,
	}

	if totp.ID != "totp-id-123" {
		t.Errorf("Expected ID 'totp-id-123', got '%s'", totp.ID)
	}

	if totp.UserID != "user-id-123" {
		t.Errorf("Expected UserID 'user-id-123', got '%s'", totp.UserID)
	}

	if totp.Secret != "JBSWY3DPEHPK3PXP" {
		t.Errorf("Expected Secret 'JBSWY3DPEHPK3PXP', got '%s'", totp.Secret)
	}

	if totp.Issuer != "AuthNas" {
		t.Errorf("Expected Issuer 'AuthNas', got '%s'", totp.Issuer)
	}
}

func TestPasskeyModel(t *testing.T) {
	now := time.Now()
	transports := "usb,ble,nfc"
	name := "My Passkey"

	passkey := Passkey{
		ID:              "passkey-id-123",
		UserID:          "user-id-123",
		Name:            &name,
		CredentialID:    "cred-123",
		PublicKey:       "public-key-data",
		AttestationType: nil,
		Transports:      &transports,
		LastUsedAt:      nil,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if passkey.ID != "passkey-id-123" {
		t.Errorf("Expected ID 'passkey-id-123', got '%s'", passkey.ID)
	}

	if passkey.CredentialID != "cred-123" {
		t.Errorf("Expected CredentialID 'cred-123', got '%s'", passkey.CredentialID)
	}

	if passkey.Name == nil || *passkey.Name != name {
		t.Errorf("Expected Name '%s', got '%v'", name, passkey.Name)
	}

	if passkey.Transports == nil || *passkey.Transports != transports {
		t.Errorf("Expected Transports '%s', got '%v'", transports, passkey.Transports)
	}
}

func TestPasskeyAuthOptionsModel(t *testing.T) {
	now := time.Now()
	userID := "user-id-123"

	opts := PasskeyAuthOptions{
		ID:        "opts-id-123",
		UserID:    &userID,
		Challenge: "challenge-data",
		Options:   "{}",
		ExpiresAt: now.Add(5 * time.Minute),
		CreatedAt: now,
	}

	if opts.ID != "opts-id-123" {
		t.Errorf("Expected ID 'opts-id-123', got '%s'", opts.ID)
	}

	if opts.UserID == nil || *opts.UserID != userID {
		t.Errorf("Expected UserID '%s', got '%v'", userID, opts.UserID)
	}

	if opts.Challenge != "challenge-data" {
		t.Errorf("Expected Challenge 'challenge-data', got '%s'", opts.Challenge)
	}

	if opts.ExpiresAt.Before(now) {
		t.Error("Expected ExpiresAt to be in the future")
	}
}
