package service

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/subtle"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/authnas/authnas/go-server/internal/config"
	"github.com/authnas/authnas/go-server/internal/model"
	"github.com/authnas/authnas/go-server/internal/repository"
	"github.com/authnas/authnas/go-server/pkg/utils"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type OIDCService struct {
	cfg             *config.Config
	db              *gorm.DB
	clientRepo      *repository.ClientRepository
	consentRepo     *repository.ConsentRepository
	oidcPayloadRepo *repository.OIDCPayloadRepository
	userRepo        *repository.UserRepository
	groupRepo       *repository.GroupRepository
	keyRepo         *repository.KeyRepository
	authService     *AuthService
	random          *utils.RandomUtil
	privateKey      *rsa.PrivateKey
	stopCleanup     chan struct{}
}

func NewOIDCService(
	cfg *config.Config,
	db *gorm.DB,
	clientRepo *repository.ClientRepository,
	consentRepo *repository.ConsentRepository,
	oidcPayloadRepo *repository.OIDCPayloadRepository,
	userRepo *repository.UserRepository,
	groupRepo *repository.GroupRepository,
	keyRepo *repository.KeyRepository,
	authService *AuthService,
	random *utils.RandomUtil,
) *OIDCService {
	svc := &OIDCService{
		cfg:             cfg,
		db:              db,
		clientRepo:      clientRepo,
		consentRepo:     consentRepo,
		oidcPayloadRepo: oidcPayloadRepo,
		userRepo:        userRepo,
		groupRepo:       groupRepo,
		keyRepo:         keyRepo,
		authService:     authService,
		random:          random,
		stopCleanup:     make(chan struct{}),
	}

	if cfg.OIDC.PrivateKey != "" && cfg.OIDC.Certificate != "" {
		if pk, err := loadPrivateKey(cfg.OIDC.PrivateKey); err == nil {
			svc.privateKey = pk
		}
	}

	if svc.privateKey == nil {
		pk, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			log.Fatalf("Failed to generate RSA private key for OIDC: %v", err)
		}
		svc.privateKey = pk
	}

	go svc.startCleanupTask()

	return svc
}

func (s *OIDCService) startCleanupTask() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if s.oidcPayloadRepo != nil {
				s.oidcPayloadRepo.DeleteExpired()
			}
			if s.keyRepo != nil {
				s.keyRepo.DeleteExpired()
			}
		case <-s.stopCleanup:
			return
		}
	}
}

func (s *OIDCService) Stop() {
	close(s.stopCleanup)
}

func loadPrivateKey(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	if block.Type != "RSA PRIVATE KEY" && block.Type != "PRIVATE KEY" {
		return nil, fmt.Errorf("unsupported PEM block type: %s", block.Type)
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		keyInterface, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
		var ok bool
		key, ok = keyInterface.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("private key is not RSA key")
		}
	}

	return key, nil
}

type OIDCDiscovery struct {
	Issuer                            string   `json:"issuer"`
	AuthorizationEndpoint             string   `json:"authorization_endpoint"`
	TokenEndpoint                     string   `json:"token_endpoint"`
	UserInfoEndpoint                  string   `json:"userinfo_endpoint"`
	JwksURI                           string   `json:"jwks_uri"`
	RevocationEndpoint                string   `json:"revocation_endpoint"`
	EndSessionEndpoint                string   `json:"endsession_endpoint,omitempty"`
	BackchannelLogoutSupported        bool     `json:"backchannel_logout_supported"`
	BackchannelLogoutURISupported     bool     `json:"backchannel_logout_uri_supported"`
	ResponseTypesSupported            []string `json:"response_types_supported"`
	SubjectTypesSupported             []string `json:"subject_types_supported"`
	IDTokenSigningAlgValuesSupported  []string `json:"id_token_signing_alg_values_supported"`
	ScopesSupported                   []string `json:"scopes_supported"`
	TokenEndpointAuthMethodsSupported []string `json:"token_endpoint_auth_methods_supported"`
	ClaimsSupported                   []string `json:"claims_supported"`
	CodeChallengeMethodsSupported     []string `json:"code_challenge_methods_supported"`
	GrantTypesSupported               []string `json:"grant_types_supported"`
}

func (s *OIDCService) Discovery() *OIDCDiscovery {
	return &OIDCDiscovery{
		Issuer:                            s.cfg.OIDC.Issuer,
		AuthorizationEndpoint:             s.cfg.App.URL + "/oidc/auth",
		TokenEndpoint:                     s.cfg.App.URL + "/oidc/token",
		UserInfoEndpoint:                  s.cfg.App.URL + "/oidc/userinfo",
		JwksURI:                           s.cfg.App.URL + "/oidc/jwks",
		RevocationEndpoint:                s.cfg.App.URL + "/oidc/token/revocation",
		EndSessionEndpoint:                s.cfg.App.URL + "/oidc/endsession",
		BackchannelLogoutSupported:        true,
		BackchannelLogoutURISupported:     true,
		ResponseTypesSupported:            []string{"code", "id_token", "code id_token"},
		SubjectTypesSupported:             []string{"public"},
		IDTokenSigningAlgValuesSupported:  []string{"RS256"},
		ScopesSupported:                   []string{"openid", "profile", "email", "groups"},
		TokenEndpointAuthMethodsSupported: []string{"client_secret_basic", "client_secret_post", "none"},
		ClaimsSupported:                   []string{"sub", "name", "email", "email_verified", "groups"},
		CodeChallengeMethodsSupported:     []string{"S256"},
		GrantTypesSupported:               []string{"authorization_code", "refresh_token", "client_credentials"},
	}
}

type AuthorizationRequest struct {
	ClientID            string `json:"client_id"`
	RedirectURI         string `json:"redirect_uri"`
	ResponseType        string `json:"response_type"`
	Scope               string `json:"scope"`
	State               string `json:"state"`
	Nonce               string `json:"nonce"`
	CodeChallenge       string `json:"code_challenge"`
	CodeChallengeMethod string `json:"code_challenge_method"`
}

type AuthorizationSession struct {
	UID           string `json:"uid"`
	ClientID      string `json:"client_id"`
	UserID        string `json:"user_id"`
	RedirectURI   string `json:"redirect_uri"`
	Scope         string `json:"scope"`
	State         string `json:"state"`
	Nonce         string `json:"nonce"`
	CodeChallenge string `json:"code_challenge"`
	AuthTime      int64  `json:"auth_time"`
}

func (s *AuthorizationSession) Payload() string {
	data, _ := json.Marshal(s)
	return string(data)
}

func (s *OIDCService) validateRedirectURI(redirectURI, registeredURI string) bool {
	if redirectURI == "" || registeredURI == "" {
		return false
	}

	if redirectURI == registeredURI {
		return true
	}

	reqURL, err := url.Parse(redirectURI)
	if err != nil {
		return false
	}

	regURL, err := url.Parse(registeredURI)
	if err != nil {
		return false
	}

	if reqURL.Scheme != regURL.Scheme {
		return false
	}
	if reqURL.Scheme != "https" && reqURL.Scheme != "http" {
		return false
	}

	if reqURL.Host != regURL.Host {
		return false
	}

	if reqURL.Path != regURL.Path {
		return false
	}

	return true
}

func (s *OIDCService) ValidateAuthorizationRequest(clientID, redirectURI, responseType, scope string) (*model.Client, error) {
	client, err := s.clientRepo.GetByClientID(clientID)
	if err != nil || client == nil {
		return nil, errors.New("invalid client")
	}

	if !s.validateRedirectURI(redirectURI, client.RedirectURIs) {
		return nil, errors.New("invalid redirect_uri")
	}

	var validResponseTypes []string
	if client.ResponseTypes != nil && *client.ResponseTypes != "" {
		validResponseTypes = strings.Split(*client.ResponseTypes, " ")
	} else {
		validResponseTypes = []string{"code"}
	}
	hasValidResponseType := false
	for _, rt := range validResponseTypes {
		if rt == responseType {
			hasValidResponseType = true
			break
		}
	}
	if !hasValidResponseType {
		return nil, errors.New("unsupported response_type")
	}

	var validScopes []string
	if client.Scopes != nil && *client.Scopes != "" {
		validScopes = strings.Split(*client.Scopes, " ")
	} else {
		validScopes = []string{"openid"}
	}
	requestedScopes := strings.Split(scope, " ")
	for _, rs := range requestedScopes {
		found := false
		for _, vs := range validScopes {
			if rs == vs {
				found = true
				break
			}
		}
		if !found && rs != "openid" {
			return nil, fmt.Errorf("unsupported scope: %s", rs)
		}
	}

	return client, nil
}

func (s *OIDCService) CreateAuthorizationSession(req *AuthorizationRequest, userID string) (*AuthorizationSession, string, error) {
	uid, err := s.random.GenerateToken(32)
	if err != nil {
		return nil, "", err
	}

	session := &AuthorizationSession{
		UID:           uid,
		ClientID:      req.ClientID,
		UserID:        userID,
		RedirectURI:   req.RedirectURI,
		Scope:         req.Scope,
		State:         req.State,
		Nonce:         req.Nonce,
		CodeChallenge: req.CodeChallenge,
		AuthTime:      time.Now().Unix(),
	}

	sessionData, err := json.Marshal(session)
	if err != nil {
		return nil, "", err
	}

	payload := &model.OIDCPayload{
		ID:        generateID(),
		UID:       uid,
		Payload:   string(sessionData),
		ExpiresAt: time.Now().Add(10 * time.Minute),
		CreatedAt: time.Now(),
	}

	if err := s.oidcPayloadRepo.Create(payload); err != nil {
		return nil, "", err
	}

	return session, uid, nil
}

func (s *OIDCService) GetAuthorizationSession(uid string) (*AuthorizationSession, error) {
	payload, err := s.oidcPayloadRepo.GetByUID(uid)
	if err != nil || payload == nil {
		return nil, errors.New("session not found")
	}

	if payload.ExpiresAt.Before(time.Now()) {
		s.oidcPayloadRepo.Delete(payload.ID)
		return nil, errors.New("session expired")
	}

	var session AuthorizationSession
	if err := json.Unmarshal([]byte(payload.Payload), &session); err != nil {
		return nil, err
	}

	return &session, nil
}

func (s *OIDCService) DeleteAuthorizationSession(uid string) error {
	payload, err := s.oidcPayloadRepo.GetByUID(uid)
	if err != nil || payload == nil {
		return nil
	}
	return s.oidcPayloadRepo.Delete(payload.ID)
}

func (s *OIDCService) CreateAuthorizationCode(session *AuthorizationSession) (string, error) {
	code, err := s.random.GenerateToken(32)
	if err != nil {
		return "", err
	}

	if err := s.oidcPayloadRepo.DeleteByUID(session.UID); err != nil {
		return "", err
	}

	codePayload := &model.OIDCPayload{
		ID:        generateID(),
		UID:       code,
		Payload:   session.Payload(),
		ExpiresAt: time.Now().Add(10 * time.Minute),
		CreatedAt: time.Now(),
	}

	if err := s.oidcPayloadRepo.Create(codePayload); err != nil {
		return "", err
	}

	return code, nil
}

type TokenRequest struct {
	GrantType    string `json:"grantType"`
	Code         string `json:"code"`
	RedirectURI  string `json:"redirectUri"`
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
	CodeVerifier string `json:"codeVerifier"`
	RefreshToken string `json:"refreshToken"`
}

type TokenResponse struct {
	AccessToken  string `json:"accessToken"`
	TokenType    string `json:"tokenType"`
	ExpiresIn    int    `json:"expiresIn"`
	ExpiresAt    string `json:"expiresAt,omitempty"`
	RefreshToken string `json:"refreshToken"`
	IDToken      string `json:"idToken"`
	Scope        string `json:"scope"`
}

func (s *OIDCService) ExchangeCode(req *TokenRequest) (*TokenResponse, error) {
	var session *AuthorizationSession
	var payload *model.OIDCPayload

	err := s.db.Transaction(func(tx *gorm.DB) error {
		var err error
		payload, err = s.oidcPayloadRepo.GetByUIDForUpdate(tx, req.Code)
		if err != nil || payload == nil {
			return errors.New("invalid authorization code")
		}

		if payload.ExpiresAt.Before(time.Now()) {
			tx.Delete(&model.OIDCPayload{}, "id = ?", payload.ID)
			return errors.New("authorization code expired")
		}

		if err := json.Unmarshal([]byte(payload.Payload), &session); err != nil {
			return err
		}

		if session.RedirectURI != req.RedirectURI {
			return errors.New("redirect_uri mismatch")
		}

		client, err := s.clientRepo.GetByClientID(session.ClientID)
		if err != nil || client == nil {
			return errors.New("invalid client")
		}

		isPublicClient := client.ClientSecret == nil || *client.ClientSecret == ""

		if isPublicClient && session.CodeChallenge == "" {
			return errors.New("code_challenge required for public clients")
		}

		if session.CodeChallenge != "" {
			if req.CodeVerifier == "" {
				return errors.New("code_verifier required when code_challenge is present")
			}
			hash := sha256Hash(req.CodeVerifier)
			if subtle.ConstantTimeCompare([]byte(hash), []byte(session.CodeChallenge)) != 1 {
				return errors.New("invalid code_verifier")
			}
		} else if isPublicClient {
			return errors.New("code_challenge required for public clients")
		}

		if client.ClientSecret != nil && *client.ClientSecret != "" && *client.ClientSecret != req.ClientSecret {
			return errors.New("invalid client_secret")
		}

		if err := tx.Delete(&model.OIDCPayload{}, "id = ?", payload.ID).Error; err != nil {
			return err
		}

		if session.UID != "" {
			if err := tx.Delete(&model.OIDCPayload{}, "uid = ?", session.UID).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	var user *model.User
	if session.UserID != "" {
		user, err = s.userRepo.GetByID(session.UserID)
		if err != nil {
			return nil, errors.New("user not found")
		}
	}

	refreshToken, err := s.random.GenerateToken(32)
	if err != nil {
		return nil, err
	}

	refreshExpiry := s.cfg.JWT.GetRefreshTokenExpiry()
	expiresAt := time.Now().Add(refreshExpiry)

	if user != nil {
		groups, _ := s.groupRepo.GetUserGroups(user.ID)
		var groupNames []string
		for _, g := range groups {
			groupNames = append(groupNames, g.Name)
		}

		email := derefString(user.Email)
		name := derefString(user.Name)

		claims := &OIDCClaims{
			Issuer:    s.cfg.OIDC.Issuer,
			Subject:   user.ID,
			Audience:  session.ClientID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			AuthTime:  session.AuthTime,
			Nonce:     session.Nonce,
			Email:     email,
			Name:      name,
			Groups:    groupNames,
		}

		idToken := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		idTokenString, err := idToken.SignedString(s.privateKey)
		if err != nil {
			return nil, err
		}

		accessClaims := jwt.MapClaims{
			"sub":    user.ID,
			"aud":    session.ClientID,
			"exp":    time.Now().Add(time.Hour).Unix(),
			"iat":    time.Now().Unix(),
			"email":  email,
			"name":   name,
			"groups": groupNames,
		}
		accessToken := jwt.NewWithClaims(jwt.SigningMethodRS256, accessClaims)
		accessTokenString, err := accessToken.SignedString(s.privateKey)
		if err != nil {
			return nil, err
		}

		hashedRefreshToken, err := bcrypt.GenerateFromPassword([]byte(refreshToken), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}

		h := sha256.Sum256([]byte(refreshToken))
		lookupHash := base64.RawStdEncoding.EncodeToString(h[:])

		keyModel := &model.Key{
			ID:                     generateID(),
			UserID:                 user.ID,
			ClientID:               session.ClientID,
			TokenVersion:           user.TokenVersion,
			RefreshTokenHash:       string(hashedRefreshToken),
			RefreshTokenLookupHash: lookupHash,
			ExpiresAt:              expiresAt,
			CreatedAt:              time.Now(),
			GroupNames:             strings.Join(groupNames, ","),
			Nonce:                  session.Nonce,
		}
		s.keyRepo.Create(keyModel)

		accessExpiry := s.cfg.JWT.GetAccessTokenExpiry()
		expiresIn := int(accessExpiry.Seconds())
		expiresAtStr := time.Now().Add(accessExpiry).Format("2006-01-02T15:04:05Z07:00")

		return &TokenResponse{
			AccessToken:  accessTokenString,
			TokenType:    "Bearer",
			ExpiresIn:    expiresIn,
			ExpiresAt:    expiresAtStr,
			RefreshToken: refreshToken,
			IDToken:      idTokenString,
			Scope:        session.Scope,
		}, nil
	}

	accessClaims := jwt.MapClaims{
		"sub": session.ClientID,
		"aud": session.ClientID,
		"exp": time.Now().Add(time.Hour).Unix(),
		"iat": time.Now().Unix(),
		"cid": session.ClientID,
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodRS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(s.privateKey)
	if err != nil {
		return nil, err
	}

	accessExpiry := s.cfg.JWT.GetAccessTokenExpiry()
	expiresIn := int(accessExpiry.Seconds())
	expiresAtStr := time.Now().Add(accessExpiry).Format("2006-01-02T15:04:05Z07:00")

	return &TokenResponse{
		AccessToken:  accessTokenString,
		TokenType:    "Bearer",
		ExpiresIn:    expiresIn,
		ExpiresAt:    expiresAtStr,
		RefreshToken: refreshToken,
		Scope:        session.Scope,
	}, nil
}

func (s *OIDCService) RefreshAccessToken(req *TokenRequest) (*TokenResponse, error) {
	if req.RefreshToken == "" {
		return nil, errors.New("refresh token required")
	}

	key, err := s.keyRepo.FindByRefreshToken(req.RefreshToken)
	if err != nil || key == nil {
		return nil, errors.New("invalid refresh token")
	}

	if key.ExpiresAt.Before(time.Now()) {
		s.keyRepo.Delete(key.ID)
		return nil, errors.New("refresh token expired")
	}

	user, err := s.userRepo.GetByID(key.UserID)
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}

	if key.TokenVersion != user.TokenVersion {
		s.keyRepo.Delete(key.ID)
		return nil, errors.New("session has been revoked")
	}

	if key.ClientID != "" {
		if key.ClientID != req.ClientID {
			return nil, errors.New("client_id mismatch: refresh token was not issued to this client")
		}
		client, err := s.clientRepo.GetByClientID(key.ClientID)
		if err != nil || client == nil {
			return nil, errors.New("invalid client")
		}
		if client.ClientSecret != nil && *client.ClientSecret != "" {
			if req.ClientSecret == "" || *client.ClientSecret != req.ClientSecret {
				return nil, errors.New("invalid client_secret")
			}
		}
	}

	cachedGroupNames := key.GroupNames
	var groupNames []string
	if cachedGroupNames != "" {
		groupNames = strings.Split(cachedGroupNames, ",")
	} else {
		groups, _ := s.groupRepo.GetUserGroups(user.ID)
		for _, g := range groups {
			groupNames = append(groupNames, g.Name)
		}
	}

	email := derefString(user.Email)
	name := derefString(user.Name)

	accessClaims := jwt.MapClaims{
		"sub":    user.ID,
		"aud":    req.ClientID,
		"exp":    time.Now().Add(time.Hour).Unix(),
		"iat":    time.Now().Unix(),
		"email":  email,
		"name":   name,
		"groups": groupNames,
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodRS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(s.privateKey)
	if err != nil {
		return nil, err
	}

	refreshTokenStr, err := s.random.GenerateToken(32)
	if err != nil {
		return nil, err
	}

	refreshExpiry := s.cfg.JWT.GetRefreshTokenExpiry()
	clientID := key.ClientID
	if req.ClientID != "" {
		clientID = req.ClientID
	}

	hashedNewRefreshToken, err := bcrypt.GenerateFromPassword([]byte(refreshTokenStr), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	h := sha256.Sum256([]byte(refreshTokenStr))
	lookupHash := base64.RawStdEncoding.EncodeToString(h[:])

	var groupNamesToCache string
	if cachedGroupNames != "" {
		groupNamesToCache = cachedGroupNames
	} else {
		groupNamesToCache = strings.Join(groupNames, ",")
	}

	newKey := &model.Key{
		ID:                     generateID(),
		UserID:                 user.ID,
		ClientID:               clientID,
		TokenVersion:           user.TokenVersion,
		RefreshTokenHash:       string(hashedNewRefreshToken),
		RefreshTokenLookupHash: lookupHash,
		ExpiresAt:              time.Now().Add(refreshExpiry),
		CreatedAt:              time.Now(),
		GroupNames:             groupNamesToCache,
	}

	idTokenClaims := &OIDCClaims{
		Issuer:    s.cfg.OIDC.Issuer,
		Subject:   user.ID,
		Audience:  clientID,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		AuthTime:  key.CreatedAt.Unix(),
		Nonce:     key.Nonce,
		Email:     email,
		Name:      name,
		Groups:    groupNames,
	}
	idToken := jwt.NewWithClaims(jwt.SigningMethodRS256, idTokenClaims)
	idTokenString, err := idToken.SignedString(s.privateKey)
	if err != nil {
		return nil, err
	}

	var finalRefreshToken string
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(newKey).Error; err != nil {
			return err
		}
		if err := tx.Delete(&model.Key{}, "id = ?", key.ID).Error; err != nil {
			return err
		}
		finalRefreshToken = refreshTokenStr
		return nil
	})

	if err != nil {
		return nil, errors.New("failed to refresh token")
	}

	accessExpiry := s.cfg.JWT.GetAccessTokenExpiry()
	expiresIn := int(accessExpiry.Seconds())
	expiresAt := time.Now().Add(accessExpiry).Format("2006-01-02T15:04:05Z07:00")

	return &TokenResponse{
		AccessToken:  accessTokenString,
		TokenType:    "Bearer",
		ExpiresIn:    expiresIn,
		ExpiresAt:    expiresAt,
		RefreshToken: finalRefreshToken,
		IDToken:      idTokenString,
		Scope:        "openid profile email groups",
	}, nil
}

type OIDCClaims struct {
	Issuer    string           `json:"iss"`
	Subject   string           `json:"sub"`
	Audience  string           `json:"aud"`
	ExpiresAt *jwt.NumericDate `json:"exp"`
	IssuedAt  *jwt.NumericDate `json:"iat"`
	Nonce     string           `json:"nonce,omitempty"`
	AuthTime  int64            `json:"auth_time,omitempty"`
	Email     string           `json:"email,omitempty"`
	Name      string           `json:"name,omitempty"`
	Groups    []string         `json:"groups,omitempty"`
}

func (c *OIDCClaims) GetExpirationTime() (*jwt.NumericDate, error) {
	return c.ExpiresAt, nil
}

func (c *OIDCClaims) GetIssuedAt() (*jwt.NumericDate, error) {
	return c.IssuedAt, nil
}

func (c *OIDCClaims) GetNotBefore() (*jwt.NumericDate, error) {
	return nil, nil
}

func (c *OIDCClaims) GetIssuer() (string, error) {
	return c.Issuer, nil
}

func (c *OIDCClaims) GetSubject() (string, error) {
	return c.Subject, nil
}

func (c *OIDCClaims) GetAudience() (jwt.ClaimStrings, error) {
	return jwt.ClaimStrings{c.Audience}, nil
}

func sha256Hash(input string) string {
	h := sha256Sum(input)
	return base64.RawURLEncoding.EncodeToString(h)
}

func sha256Sum(input string) []byte {
	hash := sha256.Sum256([]byte(input))
	return hash[:]
}

type UserInfo struct {
	Sub           string   `json:"sub"`
	Name          string   `json:"name"`
	Email         string   `json:"email"`
	EmailVerified bool     `json:"email_verified"`
	Groups        []string `json:"groups,omitempty"`
}

func (s *OIDCService) GetUserInfo(accessToken string) (*UserInfo, error) {
	token, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return &s.privateKey.PublicKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		sub, _ := claims["sub"].(string)
		if sub == "" {
			return nil, errors.New("invalid token: missing sub claim")
		}

		userInfo := &UserInfo{
			Sub:    sub,
			Groups: []string{},
		}

		if name, ok := claims["name"].(string); ok {
			userInfo.Name = name
		}
		if email, ok := claims["email"].(string); ok {
			userInfo.Email = email
		}

		if groups, ok := claims["groups"].([]interface{}); ok {
			for _, g := range groups {
				if gs, ok := g.(string); ok {
					userInfo.Groups = append(userInfo.Groups, gs)
				}
			}
		}

		if userInfo.Email != "" {
			user, err := s.userRepo.GetByID(userInfo.Sub)
			if err == nil && user != nil {
				userInfo.EmailVerified = user.EmailVerified
			}
		}

		return userInfo, nil
	}

	return nil, errors.New("invalid token")
}

func (s *OIDCService) RevokeToken(token string) error {
	key, err := s.keyRepo.FindByRefreshToken(token)
	if err != nil || key == nil {
		return nil
	}
	return s.keyRepo.Delete(key.ID)
}

func (s *OIDCService) RevokeTokensByIDTokenHint(idTokenHint string) error {
	token, err := jwt.Parse(idTokenHint, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return &s.privateKey.PublicKey, nil
	})

	if err != nil || !token.Valid {
		return nil
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil
	}

	sub, ok := claims["sub"].(string)
	if !ok || sub == "" {
		return nil
	}

	return s.keyRepo.DeleteByUserID(sub)
}

func (s *OIDCService) RevokeTokensByClientID(clientID string) error {
	return s.keyRepo.DeleteByClientID(clientID)
}

func (s *OIDCService) ValidateBackChannelLogoutToken(logoutToken string) error {
	token, err := jwt.Parse(logoutToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return &s.privateKey.PublicKey, nil
	})

	if err != nil {
		return err
	}

	if !token.Valid {
		return errors.New("invalid logout token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return errors.New("invalid logout token claims")
	}

	events, ok := claims["events"].(map[string]interface{})
	if !ok {
		return errors.New("missing events in logout token")
	}

	_, ok = events["http://schemas.openid.net/event/backchannel-logout"]
	if !ok {
		return errors.New("invalid event type in logout token")
	}

	return nil
}

func (s *OIDCService) GetPublicKey() *rsa.PublicKey {
	if s.privateKey != nil {
		return &s.privateKey.PublicKey
	}
	return nil
}

func (s *OIDCService) ValidatePostLogoutRedirectURI(postLogoutURI, clientID string) error {
	if postLogoutURI == "" {
		return nil
	}

	if clientID != "" {
		client, err := s.clientRepo.GetByClientID(clientID)
		if err != nil || client == nil {
			return errors.New("invalid client")
		}

		if client.PostLogoutRedirectURIs != nil && *client.PostLogoutRedirectURIs != "" {
			allowed := strings.Split(*client.PostLogoutRedirectURIs, " ")
			valid := false
			for _, uri := range allowed {
				if uri == postLogoutURI {
					valid = true
					break
				}
			}
			if !valid {
				return errors.New("post_logout_redirect_uri not allowed for this client")
			}
			return nil
		}
	}

	if strings.HasPrefix(postLogoutURI, "javascript:") {
		return errors.New("javascript: protocol not allowed")
	}

	parsed, err := url.Parse(postLogoutURI)
	if err != nil {
		return errors.New("invalid post_logout_redirect_uri")
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return errors.New("post_logout_redirect_uri must use http or https")
	}

	return nil
}

func (s *OIDCService) BuildRedirectURL(base string, params map[string]string) string {
	u, _ := url.Parse(base)
	q := u.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()
	return u.String()
}

func (s *OIDCService) HasValidConsent(userID, clientID, scope string) bool {
	if s.consentRepo == nil {
		return false
	}
	consent, err := s.consentRepo.GetByUserAndClient(userID, clientID)
	if err != nil || consent == nil {
		return false
	}
	if consent.ExpiresAt != nil && consent.ExpiresAt.Before(time.Now()) {
		return false
	}
	return true
}

func (s *OIDCService) SaveConsent(userID, clientID, scope string) error {
	if s.consentRepo == nil {
		return nil
	}

	existing, err := s.consentRepo.GetByUserAndClient(userID, clientID)
	if err == nil && existing != nil {
		existing.Scopes = scope
		return s.consentRepo.Update(existing)
	}

	consent := &model.Consent{
		ID:        generateID(),
		UserID:    userID,
		ClientID:  clientID,
		Scopes:    scope,
		CreatedAt: time.Now(),
	}
	return s.consentRepo.Create(consent)
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
