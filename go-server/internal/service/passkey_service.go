package service

import (
	"encoding/json"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/authnas/authnas/go-server/internal/config"
	"github.com/authnas/authnas/go-server/internal/model"
	"github.com/authnas/authnas/go-server/internal/repository"
	"github.com/authnas/authnas/go-server/pkg/utils"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
)

type PasskeyService struct {
	cfg          *config.Config
	passkeyRepo  *repository.PasskeyRepository
	userRepo     *repository.UserRepository
	authOptsRepo *repository.PasskeyAuthOptionsRepository
	webauthn     *webauthn.WebAuthn
	random       *utils.RandomUtil
	rpID         string
}

type passkeyUser struct {
	id    []byte
	name  string
	creds []webauthn.Credential
}

func (u *passkeyUser) WebAuthnID() []byte                         { return u.id }
func (u *passkeyUser) WebAuthnName() string                       { return u.name }
func (u *passkeyUser) WebAuthnDisplayName() string                { return u.name }
func (u *passkeyUser) WebAuthnCredentials() []webauthn.Credential { return u.creds }

func NewPasskeyService(
	cfg *config.Config,
	passkeyRepo *repository.PasskeyRepository,
	userRepo *repository.UserRepository,
	authOptsRepo *repository.PasskeyAuthOptionsRepository,
	random *utils.RandomUtil,
) *PasskeyService {
	rpID := extractDomain(cfg.App.URL)

	waConfig := &webauthn.Config{
		RPDisplayName: cfg.App.Name,
		RPID:          rpID,
		RPOrigins:     []string{cfg.App.URL},
	}

	wa, err := webauthn.New(waConfig)
	if err != nil {
		panic(err)
	}

	return &PasskeyService{
		cfg:          cfg,
		passkeyRepo:  passkeyRepo,
		userRepo:     userRepo,
		authOptsRepo: authOptsRepo,
		webauthn:     wa,
		random:       random,
		rpID:         rpID,
	}
}

func extractDomain(urlStr string) string {
	u, err := url.Parse(urlStr)
	if err != nil {
		return urlStr
	}
	host := u.Host
	if idx := strings.LastIndex(host, ":"); idx != -1 {
		host = host[:idx]
	}
	return host
}

func (s *PasskeyService) Create(passkey *model.Passkey) error {
	return s.passkeyRepo.Create(passkey)
}

func (s *PasskeyService) GetByID(id string) (*model.Passkey, error) {
	return s.passkeyRepo.GetByID(id)
}

func (s *PasskeyService) GetByCredentialID(credentialID string) (*model.Passkey, error) {
	return s.passkeyRepo.GetByCredentialID(credentialID)
}

func (s *PasskeyService) GetByUserID(userID string) ([]*model.Passkey, error) {
	return s.passkeyRepo.GetByUserID(userID)
}

func (s *PasskeyService) Update(passkey *model.Passkey) error {
	return s.passkeyRepo.Update(passkey)
}

func (s *PasskeyService) Delete(id string) error {
	return s.passkeyRepo.Delete(id)
}

func (s *PasskeyService) HasPasskeys(userID string) bool {
	passkeys, err := s.passkeyRepo.GetByUserID(userID)
	if err != nil {
		return false
	}
	return len(passkeys) > 0
}

type RegistrationOptions struct {
	Challenge string          `json:"challenge"`
	Options   json.RawMessage `json:"options"`
}

func (s *PasskeyService) GenerateRegistrationOptions(userID, username string) (*RegistrationOptions, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	existingCreds, err := s.passkeyRepo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}

	var creds []protocol.CredentialDescriptor
	for _, cred := range existingCreds {
		creds = append(creds, protocol.CredentialDescriptor{
			CredentialID: []byte(cred.CredentialID),
		})
	}

	options, sessionData, err := s.webauthn.BeginRegistration(
		&passkeyUser{
			id:    []byte(user.ID),
			name:  username,
			creds: []webauthn.Credential{},
		},
		webauthn.WithExclusions(creds),
	)
	if err != nil {
		return nil, err
	}

	authOpts := &model.PasskeyAuthOptions{
		ID:        generateID(),
		UserID:    &userID,
		Challenge: string(sessionData.Challenge),
		Options:   mustMarshal(sessionData),
		ExpiresAt: time.Now().Add(5 * time.Minute),
		CreatedAt: time.Now(),
	}
	if err := s.authOptsRepo.Create(authOpts); err != nil {
		return nil, err
	}

	optionsData, err := json.Marshal(options)
	if err != nil {
		return nil, err
	}

	return &RegistrationOptions{
		Challenge: string(sessionData.Challenge),
		Options:   optionsData,
	}, nil
}

func (s *PasskeyService) CreateCredentialFromResponse(userID, username, name string, parsedData *protocol.ParsedCredentialCreationData) (*model.Passkey, error) {
	authOpts, err := s.authOptsRepo.GetByUserID(userID)
	if err != nil || authOpts == nil {
		return nil, err
	}

	if authOpts.ExpiresAt.Before(time.Now()) {
		s.authOptsRepo.Delete(authOpts.ID)
		return nil, err
	}

	var sessionData webauthn.SessionData
	if err := json.Unmarshal([]byte(authOpts.Options), &sessionData); err != nil {
		return nil, err
	}

	credential, err := s.webauthn.CreateCredential(
		&passkeyUser{
			id:    []byte(userID),
			name:  username,
			creds: []webauthn.Credential{},
		},
		sessionData,
		parsedData,
	)
	if err != nil {
		return nil, err
	}

	s.authOptsRepo.Delete(authOpts.ID)

	transports := ""
	if parsedData.Response.Transports != nil {
		transports = mustMarshal(parsedData.Response.Transports)
	}

	attType := ""
	if parsedData.Response.AttestationObject.Format != "" {
		attType = parsedData.Response.AttestationObject.Format
	}

	passkeyName := name
	if passkeyName == "" {
		passkeyName = username
	}

	passkey := &model.Passkey{
		ID:              generateID(),
		UserID:          userID,
		Name:            &passkeyName,
		CredentialID:    string(credential.ID),
		PublicKey:       string(credential.PublicKey),
		AttestationType: &attType,
		Transports:      &transports,
		LastUsedAt:      nil,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := s.passkeyRepo.Create(passkey); err != nil {
		return nil, err
	}

	return passkey, nil
}

type AuthenticationOptions struct {
	Challenge string          `json:"challenge"`
	Options   json.RawMessage `json:"options"`
}

func (s *PasskeyService) GenerateAuthenticationOptions(userID string) (*AuthenticationOptions, error) {
	challenge, err := s.random.GenerateRandomBytes(32)
	if err != nil {
		return nil, err
	}

	allowedCreds := make([]protocol.CredentialDescriptor, 0)
	if userID != "" {
		passkeys, err := s.passkeyRepo.GetByUserID(userID)
		if err == nil {
			for _, pk := range passkeys {
				allowedCreds = append(allowedCreds, protocol.CredentialDescriptor{
					CredentialID: []byte(pk.CredentialID),
				})
			}
		}
	}

	options := protocol.PublicKeyCredentialRequestOptions{
		Challenge:          challenge,
		Timeout:            60000,
		RelyingPartyID:     s.rpID,
		AllowedCredentials: allowedCreds,
		UserVerification:   protocol.VerificationPreferred,
	}

	authOpts := &model.PasskeyAuthOptions{
		ID:        generateID(),
		UserID:    &userID,
		Challenge: string(challenge),
		Options:   mustMarshal(options),
		ExpiresAt: time.Now().Add(5 * time.Minute),
		CreatedAt: time.Now(),
	}
	if err := s.authOptsRepo.Create(authOpts); err != nil {
		return nil, err
	}

	optionsData, err := json.Marshal(options)
	if err != nil {
		return nil, err
	}

	return &AuthenticationOptions{
		Challenge: string(challenge),
		Options:   optionsData,
	}, nil
}

func (s *PasskeyService) ValidateAuthentication(credentialID string, response *protocol.ParsedCredentialAssertionData) (*model.Passkey, error) {
	passkey, err := s.passkeyRepo.GetByCredentialID(credentialID)
	if err != nil || passkey == nil {
		return nil, err
	}

	authOpts, err := s.authOptsRepo.GetByChallenge(string(response.Response.CollectedClientData.Challenge))
	if err != nil || authOpts == nil {
		return nil, err
	}

	if authOpts.ExpiresAt.Before(time.Now()) {
		s.authOptsRepo.Delete(authOpts.ID)
		return nil, err
	}

	if authOpts.UserID == nil || *authOpts.UserID == "" {
		s.authOptsRepo.Delete(authOpts.ID)
		return nil, errors.New("auth options user id missing")
	}

	if *authOpts.UserID != passkey.UserID {
		s.authOptsRepo.Delete(authOpts.ID)
		return nil, errors.New("user id mismatch")
	}

	s.authOptsRepo.Delete(authOpts.ID)

	existingCreds, err := s.passkeyRepo.GetByUserID(passkey.UserID)
	if err != nil {
		return nil, err
	}

	var creds []webauthn.Credential
	for _, pk := range existingCreds {
		creds = append(creds, webauthn.Credential{
			ID: []byte(pk.CredentialID),
		})
	}

	userName := passkey.UserID
	if passkey.User != nil {
		userName = passkey.User.Username
	} else {
		if u, err := s.userRepo.GetByID(passkey.UserID); err == nil && u != nil {
			userName = u.Username
		}
	}

	user := &passkeyUser{
		id:    []byte(passkey.UserID),
		name:  userName,
		creds: creds,
	}

	sessionData := webauthn.SessionData{
		Challenge:            authOpts.Challenge,
		RelyingPartyID:       s.rpID,
		UserID:               []byte(passkey.UserID),
		AllowedCredentialIDs: [][]byte{[]byte(credentialID)},
		Expires:              authOpts.ExpiresAt,
		UserVerification:     protocol.VerificationPreferred,
	}

	_, err = s.webauthn.ValidateLogin(user, sessionData, response)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	passkey.LastUsedAt = &now
	s.passkeyRepo.Update(passkey)

	return passkey, nil
}

func mustMarshal(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(data)
}
