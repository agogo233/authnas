package handler

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/authnas/authnas/go-server/internal/config"
	"github.com/authnas/authnas/go-server/internal/middleware"
	"github.com/authnas/authnas/go-server/internal/model"
	"github.com/authnas/authnas/go-server/internal/response"
	"github.com/authnas/authnas/go-server/internal/service"
	"github.com/authnas/authnas/go-server/pkg/cookie"
	"github.com/gin-gonic/gin"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/nbutton23/zxcvbn-go"
)

var (
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	htmlRegex     = regexp.MustCompile(`<[^>]*>`)
	scriptRegex   = regexp.MustCompile(`(?i)(javascript|vbscript|onerror|onload|onclick)`)
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
)

var dummyPasswordHash = "$argon2id$v=19$m=65536,t=3,p=4$XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX$XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"

func safeErrorMessage(err error, context string) string {
	log.Printf("[ERROR] %s: %v", context, err)
	return "an internal error occurred"
}

type AuthHandler struct {
	authService       *service.AuthService
	userService       *service.UserService
	totpService       *service.TOTPService
	passkeyService    *service.PasskeyService
	invitationService *service.InvitationService
	emailService      *service.EmailService
	settingService    *service.SystemSettingService
	csrfService       *service.CSRFService
	AuthMiddleware    *middleware.AuthMiddleware
	cfg               *config.Config
}

func NewAuthHandler(
	cfg *config.Config,
	authService *service.AuthService,
	userService *service.UserService,
	totpService *service.TOTPService,
	passkeyService *service.PasskeyService,
	invitationService *service.InvitationService,
	emailService *service.EmailService,
	settingService *service.SystemSettingService,
	csrfService *service.CSRFService,
	authMiddleware *middleware.AuthMiddleware,
) *AuthHandler {
	return &AuthHandler{
		authService:       authService,
		userService:       userService,
		totpService:       totpService,
		passkeyService:    passkeyService,
		invitationService: invitationService,
		emailService:      emailService,
		settingService:    settingService,
		csrfService:       csrfService,
		AuthMiddleware:    authMiddleware,
		cfg:               cfg,
	}
}

func (h *AuthHandler) isEmailVerificationRequired() bool {
	if h.settingService != nil {
		return h.settingService.IsEmailVerificationRequired()
	}
	return h.cfg.Security.EmailVerification
}

func (h *AuthHandler) isSignupRequiresApproval() bool {
	if h.settingService != nil {
		return h.settingService.IsSignupRequiresApproval()
	}
	return h.cfg.Security.SignupRequiresApproval
}

func (h *AuthHandler) isMFARequired() bool {
	if h.settingService != nil {
		settings, err := h.settingService.GetSecurity()
		if err == nil {
			return settings.MFARequired
		}
	}
	return h.cfg.Security.MFARequired
}

func (h *AuthHandler) validatePassword(password string) error {
	if h.settingService != nil {
		if err := h.settingService.ValidatePassword(password); err != nil {
			return err
		}
		return nil
	}

	result := zxcvbn.PasswordStrength(password, nil)
	if h.cfg.Security.PasswordStrength > 0 && result.Score < h.cfg.Security.PasswordStrength {
		return fmt.Errorf("password is too weak")
	}

	if h.cfg.Security.PasswordMinLength > 0 && len(password) < h.cfg.Security.PasswordMinLength {
		return fmt.Errorf("password must be at least %d characters", h.cfg.Security.PasswordMinLength)
	}

	return nil
}

func (h *AuthHandler) GetPublicConfig(c *gin.Context) {
	appName := h.cfg.App.Name
	defaultRedirect := h.cfg.App.URL
	contactEmail := h.cfg.Email.FromAddress
	passwordMinLength := h.cfg.Security.PasswordMinLength
	passwordStrength := h.cfg.Security.PasswordStrength

	if h.settingService != nil {
		appName = h.settingService.GetAppName()
		defaultRedirect = h.settingService.GetAppURL()
		if minLength, strength, err := h.settingService.GetPasswordPolicy(); err == nil {
			passwordMinLength = minLength
			passwordStrength = strength
		}
		if emailSettings, err := h.settingService.GetEmail(); err == nil && emailSettings.FromEmail != "" {
			contactEmail = emailSettings.FromEmail
		}
	}

	response.Success(c, PublicConfig{
		AppName:                appName,
		SignupRequiresApproval: h.isSignupRequiresApproval(),
		EmailVerification:      h.isEmailVerificationRequired(),
		MFARequired:            h.isMFARequired(),
		PasswordMinLength:      passwordMinLength,
		PasswordStrength:       passwordStrength,
		DefaultRedirect:        defaultRedirect,
		ContactEmail:           contactEmail,
	})
}

func (h *AuthHandler) RegisterRoutes(r *gin.RouterGroup) {
	auth := r.Group("/auth")
	{
		auth.GET("/csrf", h.GetCSRFToken)
		auth.GET("/me", h.AuthMiddleware.Authenticate(), h.GetMe)
		auth.POST("/refresh", h.RefreshToken)
		auth.POST("/login", h.Login)
		auth.POST("/register", h.Register)
		auth.POST("/passkey/start", h.PasskeyStart)
		auth.POST("/passkey/end", h.PasskeyEnd)
		auth.POST("/totp", h.TOTPVerify)
		auth.POST("/verify_email", h.VerifyEmail)
		auth.POST("/send_verify_email", h.SendVerifyEmail)
		auth.GET("/invitation/:id/:challenge", h.GetInvitation)
		auth.POST("/forgot_password", h.ForgotPassword)
		auth.POST("/reset_password", h.ResetPassword)
	}
}

func (h *AuthHandler) setRefreshTokenCookie(c *gin.Context, refreshToken string) {
	refreshExpiry := h.cfg.JWT.GetRefreshTokenExpiry()
	cookie.SetRefreshTokenCookie(c, refreshToken, refreshExpiry)
}

type LoginRequest struct {
	Input    string `json:"input" binding:"required"`
	Password string `json:"password" binding:"required"`
	Remember bool   `json:"remember"`
}

type LoginData struct {
	AccessToken string                      `json:"accessToken,omitempty"`
	ExpiresAt   string                      `json:"expiresAt,omitempty"`
	User        *response.LoginUserResponse `json:"user,omitempty"`
	MFARequired bool                        `json:"mfaRequired,omitempty"`
	MFAUserID   string                      `json:"userId,omitempty"`
	MFAToken    string                      `json:"mfaToken,omitempty"`
}

func buildLoginUserResponse(user *model.User, hasTotp, hasPasskeys bool) *response.LoginUserResponse {
	resp := &response.LoginUserResponse{
		ID:            user.ID,
		Email:         user.Email,
		Username:      user.Username,
		Name:          user.Name,
		EmailVerified: user.EmailVerified,
		Approved:      user.Approved,
		IsAdmin:       user.IsAdmin,
		MFARequired:   user.MFARequired,
		HasTotp:       hasTotp,
		HasPasskeys:   hasPasskeys,
		HasPassword:   user.PasswordHash != nil && *user.PasswordHash != "",
		TokenVersion:  user.TokenVersion,
		CreatedAt:     user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if user.ExpiresAt != nil {
		expStr := user.ExpiresAt.Format("2006-01-02T15:04:05Z07:00")
		resp.ExpiresAt = &expStr
	}
	if !user.UpdatedAt.IsZero() {
		updStr := user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00")
		resp.UpdatedAt = &updStr
	}

	return resp
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	clientIP := middleware.GetClientIP(c)

	user, err := h.userService.GetByInput(req.Input)
	if err != nil || user == nil {
		h.authService.VerifyPassword(dummyPasswordHash, req.Password)
		response.Unauthorized(c, "invalid credentials")
		return
	}

	if !middleware.RateLimitByUser(h.cfg, user.ID) {
		response.TooManyRequests(c, "too many login attempts, please try again later")
		return
	}

	if h.authService.IsLoginLockedOut(user.ID, clientIP) {
		remaining := h.authService.GetLoginLockoutRemainingTime(user.ID, clientIP)
		response.TooManyRequests(c, fmt.Sprintf("account is locked due to too many failed login attempts. Try again in %v", remaining.Round(time.Second)))
		return
	}

	if user.PasswordHash == nil || !h.authService.VerifyPassword(*user.PasswordHash, req.Password) {
		h.authService.RecordFailedLogin(user.ID, clientIP)
		response.Unauthorized(c, "invalid credentials")
		return
	}

	h.authService.RecordSuccessfulLogin(user.ID, clientIP)

	if user.MFARequired || h.isMFARequired() {
		totpEnabled := h.totpService.HasTOTP(user.ID)
		if totpEnabled {
			mfaToken, err := h.authService.GenerateMFAToken(user.ID)
			if err != nil {
				response.InternalServerError(c, "failed to generate MFA token")
				return
			}
			response.Success(c, LoginData{
				MFARequired: true,
				MFAUserID:   user.ID,
				MFAToken:    mfaToken,
			})
			return
		}
	}

	tokens, err := h.authService.GenerateTokenPair(user, c.GetHeader("User-Agent"))
	if err != nil {
		response.InternalServerError(c, "failed to generate tokens")
		return
	}

	hasTotp := h.totpService.HasTOTP(user.ID)
	hasPasskeys := h.passkeyService.HasPasskeys(user.ID)

	accessExpiry := h.cfg.JWT.GetAccessTokenExpiry()
	expiresAtTime := time.Now().Add(accessExpiry)

	h.setRefreshTokenCookie(c, tokens.RefreshToken)

	resp := &LoginData{
		AccessToken: tokens.AccessToken,
		ExpiresAt:   expiresAtTime.Format("2006-01-02T15:04:05Z07:00"),
		User:        buildLoginUserResponse(user, hasTotp, hasPasskeys),
	}

	response.Success(c, resp)
}

type RegisterRequest struct {
	Email     string `json:"email" binding:"max=250"`
	Username  string `json:"username" binding:"required"`
	Password  string `json:"password" binding:"required"`
	Name      string `json:"name"`
	InviteID  string `json:"inviteId"`
	Challenge string `json:"challenge"`
}

type RegisterResponse struct {
	AccessToken string `json:"accessToken,omitempty"`
	ExpiresAt   string `json:"expiresAt,omitempty"`
	Message     string `json:"message,omitempty"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	clientIP := middleware.GetClientIP(c)
	if req.InviteID != "" || req.Challenge != "" {
		if allowed, msg := middleware.CheckInviteRateLimit(req.Email, clientIP); !allowed {
			response.TooManyRequests(c, msg)
			return
		}
	}

	if h.isEmailVerificationRequired() || h.isSignupRequiresApproval() {
		if req.Email == "" {
			response.BadRequest(c, "email is required when email verification or approval is enabled")
			return
		}
	}

	if err := validateRegistrationInput(req.Username, req.Email); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.validatePassword(req.Password); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	var user *model.User
	var err error

	if req.InviteID != "" || req.Challenge != "" {
		if req.InviteID == "" || req.Challenge == "" {
			response.BadRequest(c, "both invite_id and challenge are required for invitation registration")
			return
		}
		user, _, err = h.userService.CreateWithInvitation(req.InviteID, req.Challenge, req.Email, req.Username, req.Password)
	} else {
		if h.isSignupRequiresApproval() {
			response.BadRequest(c, "registration requires a valid invitation")
			return
		}
		user, err = h.userService.Create(req.Email, req.Username, req.Password)
	}

	if err != nil {
		response.BadRequest(c, safeErrorMessage(err, "user registration"))
		return
	}

	tokens, err := h.authService.GenerateTokenPair(user, c.GetHeader("User-Agent"))
	if err != nil {
		response.InternalServerError(c, "failed to generate tokens")
		return
	}

	accessExpiry := h.cfg.JWT.GetAccessTokenExpiry()
	expiresAtStr := time.Now().Add(accessExpiry).Format("2006-01-02T15:04:05Z07:00")

	h.setRefreshTokenCookie(c, tokens.RefreshToken)

	response.Success(c, RegisterResponse{
		AccessToken: tokens.AccessToken,
		ExpiresAt:   expiresAtStr,
	})
}

func validateRegistrationInput(username, email string) error {
	if username == "" {
		return fmt.Errorf("username is required")
	}

	if len(username) < 3 || len(username) > 32 {
		return fmt.Errorf("username must be between 3 and 32 characters")
	}

	if !usernameRegex.MatchString(username) {
		return fmt.Errorf("username can only contain letters, numbers, and underscores")
	}

	if htmlRegex.MatchString(username) {
		return fmt.Errorf("username contains invalid characters")
	}

	if scriptRegex.MatchString(username) {
		return fmt.Errorf("username contains invalid characters")
	}

	if email != "" && !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format")
	}

	return nil
}

type PasskeyStartRequest struct {
	Username string `json:"username"`
}

type PasskeyStartResponse struct {
	Challenge string `json:"challenge,omitempty"`
	Options   string `json:"options,omitempty"`
	Message   string `json:"message,omitempty"`
}

func (h *AuthHandler) PasskeyStart(c *gin.Context) {
	var req PasskeyStartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	var userID string
	if req.Username != "" {
		user, err := h.userService.GetByUsername(req.Username)
		if err != nil {
			response.BadRequest(c, "user not found")
			return
		}
		userID = user.ID
	}

	opts, err := h.passkeyService.GenerateAuthenticationOptions(userID)
	if err != nil {
		response.InternalServerError(c, "failed to generate options")
		return
	}

	response.Success(c, PasskeyStartResponse{
		Challenge: opts.Challenge,
		Options:   string(opts.Options),
	})
}

type PasskeyEndRequest struct {
	CredentialID string `json:"credentialId" binding:"required"`
	Challenge    string `json:"challenge" binding:"required"`
	Response     string `json:"response" binding:"required"`
}

type PasskeyEndResponse struct {
	AccessToken string                      `json:"accessToken,omitempty"`
	ExpiresAt   string                      `json:"expiresAt,omitempty"`
	User        *response.LoginUserResponse `json:"user,omitempty"`
	Message     string                      `json:"message,omitempty"`
}

type PasskeyEndMFAResponse struct {
	MFARequired bool   `json:"mfaRequired"`
	MFAUserID   string `json:"userId"`
	MFAToken    string `json:"mfaToken"`
}

func (h *AuthHandler) PasskeyEnd(c *gin.Context) {
	var req PasskeyEndRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, safeErrorMessage(err, "passkey end validation"))
		return
	}

	responseData, err := protocol.ParseCredentialRequestResponseBytes([]byte(req.Response))
	if err != nil {
		response.BadRequest(c, "invalid response")
		return
	}

	passkey, err := h.passkeyService.ValidateAuthentication(req.CredentialID, responseData)
	if err != nil || passkey == nil {
		response.Unauthorized(c, "authentication failed")
		return
	}

	user, err := h.userService.GetByID(passkey.UserID)
	if err != nil || user == nil {
		response.Unauthorized(c, "user not found")
		return
	}

	if h.isEmailVerificationRequired() && !user.EmailVerified {
		response.Forbidden(c, "email not verified")
		return
	}

	if h.isSignupRequiresApproval() && !user.Approved {
		response.Forbidden(c, "account not approved")
		return
	}

	if user.ExpiresAt != nil && user.ExpiresAt.Before(time.Now()) {
		response.Forbidden(c, "account has expired")
		return
	}

	if user.MFARequired || h.isMFARequired() {
		totpEnabled := h.totpService.HasTOTP(user.ID)
		if totpEnabled {
			mfaToken, err := h.authService.GenerateMFAToken(user.ID)
			if err != nil {
				response.InternalServerError(c, "failed to generate MFA token")
				return
			}
			response.Success(c, PasskeyEndMFAResponse{
				MFARequired: true,
				MFAUserID:   user.ID,
				MFAToken:    mfaToken,
			})
			return
		}
	}

	tokens, err := h.authService.GenerateTokenPair(user, c.GetHeader("User-Agent"))
	if err != nil {
		response.InternalServerError(c, "failed to generate tokens")
		return
	}

	accessExpiry := h.cfg.JWT.GetAccessTokenExpiry()
	expiresAtStr := time.Now().Add(accessExpiry).Format("2006-01-02T15:04:05Z07:00")
	hasTotp := h.totpService.HasTOTP(user.ID)
	hasPasskeys := h.passkeyService.HasPasskeys(user.ID)

	h.setRefreshTokenCookie(c, tokens.RefreshToken)

	response.Success(c, PasskeyEndResponse{
		AccessToken: tokens.AccessToken,
		ExpiresAt:   expiresAtStr,
		User:        buildLoginUserResponse(user, hasTotp, hasPasskeys),
	})
}

type TOTPVerifyRequest struct {
	Token    string `json:"token" binding:"required"`
	MFAToken string `json:"mfaToken"`
}

type TOTPVerifyResponse struct {
	AccessToken string                      `json:"accessToken,omitempty"`
	ExpiresAt   string                      `json:"expiresAt,omitempty"`
	User        *response.LoginUserResponse `json:"user,omitempty"`
}

func (h *AuthHandler) TOTPVerify(c *gin.Context) {
	var req TOTPVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	var userID string
	var err error

	if req.MFAToken != "" {
		userID, err = h.authService.ValidateMFAToken(req.MFAToken)
		if err != nil {
			response.Unauthorized(c, "invalid or expired MFA session")
			return
		}
	} else {
		userID = c.GetString("user_id")
		if userID == "" {
			response.Unauthorized(c, "authentication required")
			return
		}
	}

	if !h.totpService.Validate(userID, req.Token) {
		response.Unauthorized(c, "invalid token")
		return
	}

	user, err := h.userService.GetByID(userID)
	if err != nil || user == nil {
		response.Unauthorized(c, "user not found")
		return
	}

	tokens, err := h.authService.GenerateTokenPair(user, c.GetHeader("User-Agent"))
	if err != nil {
		response.InternalServerError(c, "failed to generate tokens")
		return
	}

	accessExpiry := h.cfg.JWT.GetAccessTokenExpiry()
	expiresAtStr := time.Now().Add(accessExpiry).Format("2006-01-02T15:04:05Z07:00")
	hasTotp := h.totpService.HasTOTP(user.ID)
	hasPasskeys := h.passkeyService.HasPasskeys(user.ID)

	h.setRefreshTokenCookie(c, tokens.RefreshToken)

	response.Success(c, TOTPVerifyResponse{
		AccessToken: tokens.AccessToken,
		ExpiresAt:   expiresAtStr,
		User:        buildLoginUserResponse(user, hasTotp, hasPasskeys),
	})
}

type VerifyEmailRequest struct {
	UserID    string `json:"userId"`
	Challenge string `json:"challenge"`
}

func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	var req VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, safeErrorMessage(err, "verify email binding"))
		return
	}

	userID := req.UserID
	challenge := req.Challenge

	if userID == "" || challenge == "" {
		response.BadRequest(c, "userId and challenge are required")
		return
	}

	err := h.userService.VerifyEmail(userID, challenge)
	if err != nil {
		response.BadRequest(c, safeErrorMessage(err, "verify email"))
		return
	}

	response.SuccessWithMessage(c, "email verified successfully")
}

type SendVerifyEmailRequest struct {
	Email string `json:"email" binding:"required"`
}

func (h *AuthHandler) SendVerifyEmail(c *gin.Context) {
	var req SendVerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, safeErrorMessage(err, "send verify email binding"))
		return
	}

	clientIP := middleware.GetClientIP(c)
	if allowed, msg := middleware.CheckEmailVerifyRateLimit(req.Email, clientIP); !allowed {
		response.TooManyRequests(c, msg)
		return
	}

	var userID string

	if req.Email != "" {
		user, err := h.userService.GetByInput(req.Email)
		if err != nil || user == nil {
			response.NotFound(c, "user not found")
			return
		}
		userID = user.ID
	} else {
		currentUser := middleware.GetCurrentUser(c)
		if currentUser == nil {
			response.Unauthorized(c, "user not found")
			return
		}
		userID = currentUser.ID
	}

	_, err := h.userService.SendEmailVerification(userID)
	if err != nil {
		response.InternalServerError(c, safeErrorMessage(err, "send email verification"))
		return
	}

	middleware.RecordEmailVerifyAttempt(req.Email, clientIP)
	response.SuccessWithMessage(c, "verification email sent")
}

type InvitationResponse struct {
	Email    string  `json:"email,omitempty"`
	Username *string `json:"username,omitempty"`
}

func (h *AuthHandler) GetInvitation(c *gin.Context) {
	id := c.Param("id")
	challenge := c.Param("challenge")

	invitation, err := h.invitationService.GetByID(id)
	if err != nil || invitation == nil {
		response.NotFound(c, "invitation not found")
		return
	}

	if invitation.Code != challenge {
		response.NotFound(c, "invalid challenge")
		return
	}

	if invitation.ExpiresAt.Before(time.Now()) {
		response.NotFound(c, "invitation has expired")
		return
	}

	response.Success(c, InvitationResponse{
		Email:    invitation.Email,
		Username: invitation.Username,
	})
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required"`
}

func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, safeErrorMessage(err, "forgot password binding"))
		return
	}

	clientIP := middleware.GetClientIP(c)
	if allowed, msg := middleware.CheckPasswordResetRateLimit(req.Email, clientIP); !allowed {
		response.TooManyRequests(c, msg)
		return
	}

	_ = h.userService.ForgotPassword(req.Email)
	middleware.RecordPasswordResetAttempt(req.Email, clientIP)

	response.SuccessWithMessage(c, "If an account with that email exists, a password reset link has been sent")
}

type AuthResetPasswordRequest struct {
	Code        string `json:"code" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required"`
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req AuthResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, safeErrorMessage(err, "reset password binding"))
		return
	}

	clientIP := middleware.GetClientIP(c)
	if allowed, msg := middleware.CheckResetCodeRateLimit(req.Code, clientIP); !allowed {
		response.TooManyRequests(c, msg)
		return
	}

	err := h.userService.ResetPasswordByCode(req.Code, req.NewPassword)
	if err != nil {
		middleware.RecordResetCodeAttempt(req.Code, clientIP)
		response.BadRequest(c, safeErrorMessage(err, "reset password"))
		return
	}

	middleware.ResetResetCodeRateLimit(req.Code, clientIP)
	response.SuccessWithMessage(c, "password reset successfully")
}

type MeResponse struct {
	ID            string  `json:"id"`
	Email         *string `json:"email"`
	Username      string  `json:"username"`
	Name          *string `json:"name"`
	EmailVerified bool    `json:"emailVerified"`
	Approved      bool    `json:"approved"`
	IsAdmin       bool    `json:"isAdmin"`
	MFARequired   bool    `json:"mfaRequired"`
	HasTotp       bool    `json:"hasTotp"`
	HasPasskeys   bool    `json:"hasPasskeys"`
	HasPassword   bool    `json:"hasPassword"`
	TokenVersion  int     `json:"tokenVersion"`
	CreatedAt     string  `json:"createdAt"`
	UpdatedAt     *string `json:"updatedAt"`
	ExpiresAt     *string `json:"expiresAt"`
}

func (h *AuthHandler) GetMe(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		response.Unauthorized(c, "user not found")
		return
	}

	hasTotp := h.totpService.HasTOTP(user.ID)
	hasPasskeys := h.passkeyService.HasPasskeys(user.ID)

	resp := &MeResponse{
		ID:            user.ID,
		Email:         user.Email,
		Username:      user.Username,
		Name:          user.Name,
		EmailVerified: user.EmailVerified,
		Approved:      user.Approved,
		IsAdmin:       user.IsAdmin,
		MFARequired:   user.MFARequired,
		HasTotp:       hasTotp,
		HasPasskeys:   hasPasskeys,
		HasPassword:   user.PasswordHash != nil && *user.PasswordHash != "",
		TokenVersion:  user.TokenVersion,
		CreatedAt:     user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if !user.UpdatedAt.IsZero() {
		updStr := user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00")
		resp.UpdatedAt = &updStr
	}
	if user.ExpiresAt != nil {
		expStr := user.ExpiresAt.Format("2006-01-02T15:04:05Z07:00")
		resp.ExpiresAt = &expStr
	}

	response.Success(c, resp)
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	refreshToken := cookie.GetRefreshTokenFromCookie(c)
	if refreshToken == "" {
		response.Unauthorized(c, "refresh token not found")
		return
	}

	tokenPair, err := h.authService.RefreshToken(refreshToken)
	if err != nil {
		cookie.ClearRefreshTokenCookie(c)
		response.Unauthorized(c, "invalid or expired refresh token")
		return
	}

	h.setRefreshTokenCookie(c, tokenPair.RefreshToken)

	accessExpiry := h.cfg.JWT.GetAccessTokenExpiry()
	expiresAtStr := time.Now().Add(accessExpiry).Format("2006-01-02T15:04:05Z07:00")

	response.Success(c, gin.H{
		"accessToken": tokenPair.AccessToken,
		"expiresAt":   expiresAtStr,
	})
}

type CSRFTokenResponse struct {
	Token     string `json:"token"`
	ExpiresAt string `json:"expiresAt"`
}

func (h *AuthHandler) GetCSRFToken(c *gin.Context) {
	userID := ""
	currentUser := middleware.GetCurrentUser(c)
	if currentUser != nil {
		userID = currentUser.ID
	}

	csrfToken, err := h.csrfService.GenerateToken(userID)
	if err != nil {
		response.InternalServerError(c, "failed to generate CSRF token")
		return
	}

	maxAge := int(h.csrfService.GetExpiry().Seconds())
	secure := strings.HasPrefix(h.cfg.App.URL, "https://")
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		"csrf_token",
		csrfToken.Token,
		maxAge,
		"/",
		"",
		secure,
		true,
	)

	response.Success(c, CSRFTokenResponse{
		Token:     csrfToken.Token,
		ExpiresAt: csrfToken.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}
