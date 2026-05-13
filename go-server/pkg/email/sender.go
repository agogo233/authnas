package email

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net/smtp"

	"github.com/authnas/authnas/go-server/internal/config"
)

type Sender struct {
	cfg *config.Config
}

func NewSender(cfg *config.Config) *Sender {
	return &Sender{cfg: cfg}
}

func (s *Sender) Send(to, subject, body string) error {
	if !s.cfg.Email.Enabled {
		return nil
	}

	addr := fmt.Sprintf("%s:%d", s.cfg.Email.SMTPHost, s.cfg.Email.SMTPPort)

	auth := smtp.PlainAuth("", s.cfg.Email.SMTPUser, s.cfg.Email.SMTPPassword, s.cfg.Email.SMTPHost)

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		s.cfg.Email.FromAddress, to, subject, body)

	err := smtp.SendMail(addr, auth, s.cfg.Email.FromAddress, []string{to}, []byte(msg))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func (s *Sender) SendHTML(to, subject, htmlBody string) error {
	if !s.cfg.Email.Enabled {
		return nil
	}

	addr := fmt.Sprintf("%s:%d", s.cfg.Email.SMTPHost, s.cfg.Email.SMTPPort)

	auth := smtp.PlainAuth("", s.cfg.Email.SMTPUser, s.cfg.Email.SMTPPassword, s.cfg.Email.SMTPHost)

	headers := make(map[string]string)
	headers["From"] = s.cfg.Email.FromAddress
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=\"utf-8\""

	var msg bytes.Buffer
	for k, v := range headers {
		msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	msg.WriteString("\r\n")
	msg.WriteString(htmlBody)

	err := smtp.SendMail(addr, auth, s.cfg.Email.FromAddress, []string{to}, msg.Bytes())
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func (s *Sender) SendTLS(to, subject, body string) error {
	if !s.cfg.Email.Enabled {
		return nil
	}

	addr := fmt.Sprintf("%s:%d", s.cfg.Email.SMTPHost, s.cfg.Email.SMTPPort)

	auth := smtp.PlainAuth("", s.cfg.Email.SMTPUser, s.cfg.Email.SMTPPassword, s.cfg.Email.SMTPHost)

	tlsConfig := &tls.Config{
		ServerName: s.cfg.Email.SMTPHost,
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to connect via TLS: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.cfg.Email.SMTPHost)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Close()

	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	if err := client.Mail(s.cfg.Email.FromAddress); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("failed to set recipient: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to open data writer: %w", err)
	}

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		s.cfg.Email.FromAddress, to, subject, body)

	if _, err := w.Write([]byte(msg)); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("failed to close data writer: %w", err)
	}

	return nil
}
