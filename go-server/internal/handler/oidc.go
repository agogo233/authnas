package handler

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/authnas/authnas/go-server/internal/config"
	"github.com/authnas/authnas/go-server/internal/middleware"
	"github.com/authnas/authnas/go-server/internal/response"
	"github.com/authnas/authnas/go-server/internal/service"
	"github.com/authnas/authnas/go-server/pkg/cookie"
	"github.com/gin-gonic/gin"
)

type OIDCHandler struct {
	cfg         *config.Config
	oidcService *service.OIDCService
	clientRepo  *service.ClientService
	userService *service.UserService
	csrfService *service.CSRFService
	jwksCache   []byte
}

func NewOIDCHandler(cfg *config.Config, oidcService *service.OIDCService, clientRepo *service.ClientService, userService *service.UserService, csrfService *service.CSRFService) *OIDCHandler {
	h := &OIDCHandler{
		cfg:         cfg,
		oidcService: oidcService,
		clientRepo:  clientRepo,
		userService: userService,
		csrfService: csrfService,
	}
	h.computeJWKSCache()
	return h
}

func (h *OIDCHandler) computeJWKSCache() {
	pubKey := h.oidcService.GetPublicKey()
	if pubKey == nil {
		log.Printf("[WARN] JWKS cache initialization skipped: public key not available")
		return
	}

	jid := base64.RawURLEncoding.EncodeToString([]byte("authnas-key-1"))

	nBytes := pubKey.N.Bytes()
	n := base64.RawURLEncoding.EncodeToString(nBytes)

	e := big.NewInt(int64(pubKey.E))
	eBytes := e.Bytes()
	exponent := base64.RawURLEncoding.EncodeToString(eBytes)

	jwks := JWKS{
		Keys: []JWK{
			{
				Kty: "RSA",
				Use: "sig",
				Kid: jid,
				Alg: "RS256",
				N:   n,
				E:   exponent,
			},
		},
	}

	data, err := json.Marshal(jwks)
	if err != nil {
		log.Printf("[WARN] JWKS cache serialization failed: %v", err)
		return
	}

	h.jwksCache = data
}

func (h *OIDCHandler) RegisterRoutes(r *gin.Engine) {
	oidc := r.Group("/oidc")
	{
		oidc.GET("/.well-known/openid-configuration", h.Discovery)
		oidc.GET("/jwks", h.JWKS)
		oidc.GET("/auth", h.Authorization)
		oidc.GET("/token", h.Token)
		oidc.POST("/token", h.TokenPost)
		oidc.GET("/userinfo", h.UserInfo)
		oidc.POST("/token/revocation", h.Revocation)
		oidc.GET("/logout", h.Logout)
		oidc.GET("/endsession", h.EndSession)
		oidc.POST("/backchannel-logout", h.BackChannelLogout)
		oidc.GET("/interaction/:uid", h.Interaction)
		oidc.POST("/interaction/:uid/confirm", h.InteractionConfirm)
		oidc.DELETE("/interaction/:uid/cancel", h.InteractionCancel)
	}
}

func (h *OIDCHandler) Discovery(c *gin.Context) {
	discovery := h.oidcService.Discovery()
	c.JSON(http.StatusOK, discovery)
}

type JWK struct {
	Kty string `json:"kty"`
	Use string `json:"use"`
	Kid string `json:"kid"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type JWKS struct {
	Keys []JWK `json:"keys"`
}

func (h *OIDCHandler) JWKS(c *gin.Context) {
	if len(h.jwksCache) == 0 {
		response.ServiceUnavailable(c, "public key not available")
		return
	}

	c.Data(http.StatusOK, "application/json", h.jwksCache)
}

func (h *OIDCHandler) Authorization(c *gin.Context) {
	clientID := c.Query("client_id")
	redirectURI := c.Query("redirect_uri")
	responseType := c.Query("response_type")
	scope := c.Query("scope")
	state := c.Query("state")
	nonce := c.Query("nonce")
	codeChallenge := c.Query("code_challenge")
	codeChallengeMethod := c.Query("code_challenge_method")

	if clientID == "" || redirectURI == "" || responseType == "" {
		response.BadRequest(c, "invalid_request")
		return
	}

	if scope == "" {
		scope = "openid"
	}

	_, err := h.oidcService.ValidateAuthorizationRequest(clientID, redirectURI, responseType, scope)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userID := ""
	currentUser := middleware.GetCurrentUser(c)
	if currentUser != nil {
		userID = currentUser.ID
	}

	req := &service.AuthorizationRequest{
		ClientID:            clientID,
		RedirectURI:         redirectURI,
		ResponseType:        responseType,
		Scope:               scope,
		State:               state,
		Nonce:               nonce,
		CodeChallenge:       codeChallenge,
		CodeChallengeMethod: codeChallengeMethod,
	}

	_, uid, err := h.oidcService.CreateAuthorizationSession(req, userID)
	if err != nil {
		response.InternalServerError(c, "failed to create session")
		return
	}

	frontendBase := h.cfg.App.URL

	consentURL := frontendBase + "/consent/" + uid
	if state != "" {
		consentURL += "?state=" + state
	}

	c.Redirect(http.StatusFound, consentURL)
}

func (h *OIDCHandler) Token(c *gin.Context) {
	response.BadRequest(c, "use POST for token endpoint")
}

func (h *OIDCHandler) TokenPost(c *gin.Context) {
	var req service.TokenRequest
	if err := c.ShouldBind(&req); err != nil {
		response.BadRequest(c, "invalid_request")
		return
	}

	if req.GrantType == "authorization_code" {
		tokenResp, err := h.oidcService.ExchangeCode(&req)
		if err != nil {
			response.BadRequest(c, safeErrorMessage(err, "oidc exchange code"))
			return
		}
		response.Success(c, tokenResp)
		return
	}

	if req.GrantType == "refresh_token" {
		tokenResp, err := h.oidcService.RefreshAccessToken(&req)
		if err != nil {
			response.BadRequest(c, safeErrorMessage(err, "oidc refresh token"))
			return
		}
		response.Success(c, tokenResp)
		return
	}

	response.BadRequest(c, "unsupported_grant_type")
}

func (h *OIDCHandler) UserInfo(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		response.Unauthorized(c, "missing authorization header")
		return
	}

	token := ""
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		token = authHeader[7:]
	}

	if token == "" {
		response.Unauthorized(c, "invalid token")
		return
	}

	userInfo, err := h.oidcService.GetUserInfo(token)
	if err != nil {
		response.Unauthorized(c, safeErrorMessage(err, "oidc userinfo"))
		return
	}

	response.Success(c, userInfo)
}

func (h *OIDCHandler) Revocation(c *gin.Context) {
	token := c.PostForm("token")
	if token == "" {
		response.BadRequest(c, "missing token")
		return
	}

	if err := h.oidcService.RevokeToken(token); err != nil {
		response.Success(c, gin.H{"status": "ok"})
		return
	}

	response.Success(c, gin.H{"status": "ok"})
}

func (h *OIDCHandler) Logout(c *gin.Context) {
	idTokenHint := c.Query("id_token_hint")
	postLogoutRedirectURI := c.Query("post_logout_redirect_uri")
	state := c.Query("state")
	clientID := c.Query("client_id")

	if idTokenHint != "" {
		h.oidcService.RevokeTokensByIDTokenHint(idTokenHint)
	}

	cookie.ClearRefreshTokenCookie(c)

	redirectURL := "/"
	if postLogoutRedirectURI != "" {
		if err := h.oidcService.ValidatePostLogoutRedirectURI(postLogoutRedirectURI, clientID); err != nil {
			redirectURL = "/"
		} else {
			redirectURL = postLogoutRedirectURI
		}
	}
	if state != "" {
		u, err := url.Parse(redirectURL)
		if err != nil {
			redirectURL = "/"
		} else {
			q := u.Query()
			q.Set("state", state)
			u.RawQuery = q.Encode()
			redirectURL = u.String()
		}
	}

	c.Redirect(http.StatusFound, redirectURL)
}

func (h *OIDCHandler) Interaction(c *gin.Context) {
	uid := c.Param("uid")

	session, err := h.oidcService.GetAuthorizationSession(uid)
	if err != nil || session == nil {
		response.NotFound(c, "session not found or expired")
		return
	}

	client, err := h.clientRepo.GetByClientID(session.ClientID)
	if err != nil || client == nil {
		response.NotFound(c, "client not found")
		return
	}

	scopes := []string{}
	if session.Scope != "" {
		scopes = strings.Split(session.Scope, " ")
	}

	response.Success(c, gin.H{
		"uid": uid,
		"client": gin.H{
			"clientId": client.ClientID,
			"name":     client.Name,
			"logoUri":  client.LogoURI,
		},
		"scopes": scopes,
		"claims": gin.H{
			"sub": session.UserID,
		},
	})
}

func (h *OIDCHandler) InteractionConfirm(c *gin.Context) {
	uid := c.Param("uid")

	csrfToken := c.PostForm("csrf_token")
	if csrfToken == "" {
		csrfToken = c.GetHeader("X-CSRF-Token")
	}
	if !h.csrfService.ValidateToken(csrfToken, "oidc-consent") {
		response.Forbidden(c, "invalid CSRF token")
		return
	}

	session, err := h.oidcService.GetAuthorizationSession(uid)
	if err != nil || session == nil {
		response.NotFound(c, "session not found or expired")
		return
	}

	if session.UserID != "" && h.userService != nil {
		user, err := h.userService.GetByID(session.UserID)
		if err != nil || user == nil {
			response.Forbidden(c, "user not found")
			return
		}
		if !user.Approved {
			response.Forbidden(c, "user account is disabled")
			return
		}
		if h.cfg.Security.EmailVerification && !user.EmailVerified {
			response.Forbidden(c, "email not verified")
			return
		}
		if user.ExpiresAt != nil && user.ExpiresAt.Before(time.Now()) {
			response.Forbidden(c, "account has expired")
			return
		}
	}

	if !h.oidcService.HasValidConsent(session.UserID, session.ClientID, session.Scope) {
		h.oidcService.SaveConsent(session.UserID, session.ClientID, session.Scope)
	}

	code, err := h.oidcService.CreateAuthorizationCode(session)
	if err != nil {
		response.InternalServerError(c, "failed to create authorization code")
		return
	}

	h.oidcService.DeleteAuthorizationSession(uid)

	redirectURL := h.oidcService.BuildRedirectURL(session.RedirectURI, map[string]string{
		"code":  code,
		"state": session.State,
	})

	response.Success(c, gin.H{
		"redirectTo": redirectURL,
	})
}

func (h *OIDCHandler) InteractionCancel(c *gin.Context) {
	uid := c.Param("uid")

	session, err := h.oidcService.GetAuthorizationSession(uid)
	if err != nil || session == nil {
		response.NotFound(c, "session not found or expired")
		return
	}

	h.oidcService.DeleteAuthorizationSession(uid)

	redirectURL := h.oidcService.BuildRedirectURL(session.RedirectURI, map[string]string{
		"error":        "access_denied",
		"error_reason": "user_denied",
		"state":        session.State,
	})

	response.Success(c, gin.H{
		"redirectTo": redirectURL,
	})
}

func (h *OIDCHandler) EndSession(c *gin.Context) {
	idTokenHint := c.Query("id_token_hint")
	postLogoutRedirectURI := c.Query("post_logout_redirect_uri")
	state := c.Query("state")
	clientID := c.Query("client_id")

	if idTokenHint != "" {
		h.oidcService.RevokeTokensByIDTokenHint(idTokenHint)
	}

	if clientID != "" {
		h.oidcService.RevokeTokensByClientID(clientID)
	}

	redirectURL := "/"
	if postLogoutRedirectURI != "" {
		if err := h.oidcService.ValidatePostLogoutRedirectURI(postLogoutRedirectURI, clientID); err != nil {
			redirectURL = "/"
		} else {
			redirectURL = postLogoutRedirectURI
		}
	}
	if state != "" {
		if strings.Contains(redirectURL, "?") {
			redirectURL += "&state=" + state
		} else {
			redirectURL += "?state=" + state
		}
	}

	c.Redirect(http.StatusFound, redirectURL)
}

func (h *OIDCHandler) BackChannelLogout(c *gin.Context) {
	logoutToken := c.PostForm("logout_token")

	if logoutToken == "" {
		response.BadRequest(c, "logout_token required")
		return
	}

	if err := h.oidcService.ValidateBackChannelLogoutToken(logoutToken); err != nil {
		response.BadRequest(c, "invalid_logout_token")
		return
	}

	response.Success(c, gin.H{"status": "ok"})
}
