package service

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"

	"github.com/authnas/authnas/go-server/internal/config"
	"github.com/authnas/authnas/go-server/internal/model"
	"github.com/authnas/authnas/go-server/internal/repository"
	"github.com/authnas/authnas/go-server/pkg/email"
)

type EmailService struct {
	cfg                   *config.Config
	emailVerificationRepo *repository.EmailVerificationRepository
	passwordResetRepo     *repository.PasswordResetRepository
	emailLogRepo          *repository.EmailLogRepository
	sender                *email.Sender
}

func NewEmailService(
	cfg *config.Config,
	emailVerificationRepo *repository.EmailVerificationRepository,
	passwordResetRepo *repository.PasswordResetRepository,
	emailLogRepo *repository.EmailLogRepository,
	sender *email.Sender,
) *EmailService {
	return &EmailService{
		cfg:                   cfg,
		emailVerificationRepo: emailVerificationRepo,
		passwordResetRepo:     passwordResetRepo,
		emailLogRepo:          emailLogRepo,
		sender:                sender,
	}
}

type EmailTemplateData struct {
	AppName  string
	UserName string
	Link     string
	Code     string
}

func (s *EmailService) SendVerificationEmail(user *model.User, code string) error {
	if user.Email == nil {
		return errors.New("user email is nil")
	}

	link := fmt.Sprintf("%s/verify-email?code=%s", s.cfg.App.URL, code)
	subject := "Verify your email"

	htmlBody, err := s.renderTemplate("verification", EmailTemplateData{
		AppName:  s.cfg.App.Name,
		UserName: user.Username,
		Link:     link,
		Code:     code,
	})
	if err != nil {
		return err
	}

	if s.sender == nil {
		s.logEmail(*user.Email, subject, "verification", "skipped", "email sender not configured")
		return nil
	}

	if err := s.sender.SendHTML(*user.Email, subject, htmlBody); err != nil {
		s.logEmail(*user.Email, subject, "verification", "failed", err.Error())
		return err
	}

	s.logEmail(*user.Email, subject, "verification", "sent", "")
	return nil
}

func (s *EmailService) SendPasswordResetEmail(user *model.User, code string) error {
	if user.Email == nil {
		return errors.New("user email is nil")
	}

	link := fmt.Sprintf("%s/reset-password?code=%s", s.cfg.App.URL, code)
	subject := "Reset your password"

	htmlBody, err := s.renderTemplate("password_reset", EmailTemplateData{
		AppName:  s.cfg.App.Name,
		UserName: user.Username,
		Link:     link,
		Code:     code,
	})
	if err != nil {
		return err
	}

	if s.sender == nil {
		s.logEmail(*user.Email, subject, "password_reset", "skipped", "email sender not configured")
		return nil
	}

	if err := s.sender.SendHTML(*user.Email, subject, htmlBody); err != nil {
		s.logEmail(*user.Email, subject, "password_reset", "failed", err.Error())
		return err
	}

	s.logEmail(*user.Email, subject, "password_reset", "sent", "")
	return nil
}

func (s *EmailService) renderTemplate(templateName string, data EmailTemplateData) (string, error) {
	tmpl := s.getEmailTemplate(templateName)

	var buf bytes.Buffer
	err := tmpl.Execute(&buf, data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (s *EmailService) getEmailTemplate(name string) *template.Template {
	templates := map[string]string{
		"verification":   verificationTemplate,
		"password_reset": passwordResetTemplate,
	}

	tmplStr, ok := templates[name]
	if !ok {
		tmplStr = defaultTemplate
	}

	return template.Must(template.New(name).Parse(tmplStr))
}

func (s *EmailService) logEmail(recipient, subject, template, status, errorMsg string) {
	if s.emailLogRepo == nil {
		return
	}
	log := &model.EmailLog{
		ID:        generateID(),
		Recipient: recipient,
		Subject:   subject,
		Template:  template,
		Status:    status,
		CreatedAt: now(),
	}
	if errorMsg != "" {
		log.Error = &errorMsg
	}
	s.emailLogRepo.Create(log)
}

func (s *EmailService) SendTestEmail(to string) error {
	subject := "Test Email from AuthNas"
	htmlBody := `
<!DOCTYPE html>
<html>
<head><meta charset="utf-8"><title>Test Email</title></head>
<body>
<h1>Test Email</h1>
<p>This is a test email from AuthNas. If you received this, your SMTP configuration is working correctly.</p>
</body>
</html>
`
	if s.sender == nil {
		s.logEmail(to, subject, "test", "skipped", "email sender not configured")
		return errors.New("email sender not configured")
	}
	if err := s.sender.SendHTML(to, subject, htmlBody); err != nil {
		s.logEmail(to, subject, "test", "failed", err.Error())
		return err
	}
	s.logEmail(to, subject, "test", "sent", "")
	return nil
}

const defaultTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>{{.AppName}}</title>
</head>
<body>
    <h1>{{.AppName}}</h1>
    <p>Hello {{.UserName}},</p>
    <p>{{.Link}}</p>
</body>
</html>
`

const verificationTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>Email Verification</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .button { display: inline-block; padding: 10px 20px; background-color: #007bff; color: white; text-decoration: none; border-radius: 5px; }
        .footer { margin-top: 20px; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Verify Your Email</h1>
        <p>Hello {{.UserName}},</p>
        <p>Thank you for signing up! Please verify your email address by clicking the button below:</p>
        <p><a href="{{.Link}}" class="button">Verify Email</a></p>
        <p>Or copy and paste this link into your browser:</p>
        <p>{{.Link}}</p>
        <p>This link will expire in 24 hours.</p>
        <div class="footer">
            <p>If you didn't create an account, please ignore this email.</p>
        </div>
    </div>
</body>
</html>
`

const passwordResetTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>Password Reset</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .button { display: inline-block; padding: 10px 20px; background-color: #dc3545; color: white; text-decoration: none; border-radius: 5px; }
        .footer { margin-top: 20px; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Reset Your Password</h1>
        <p>Hello {{.UserName}},</p>
        <p>We received a request to reset your password. Click the button below to reset it:</p>
        <p><a href="{{.Link}}" class="button">Reset Password</a></p>
        <p>Or copy and paste this link into your browser:</p>
        <p>{{.Link}}</p>
        <p>This link will expire in 1 hour.</p>
        <div class="footer">
            <p>If you didn't request a password reset, please ignore this email or contact support if you have concerns.</p>
        </div>
    </div>
</body>
</html>
`
