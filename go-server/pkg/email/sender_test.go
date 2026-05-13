package email

import (
	"testing"

	"github.com/authnas/authnas/go-server/internal/config"
)

func TestNewSender(t *testing.T) {
	cfg := &config.Config{
		Email: config.EmailConfig{
			Enabled:  false,
			SMTPHost: "localhost",
			SMTPPort: 587,
		},
	}

	sender := NewSender(cfg)
	if sender == nil {
		t.Fatal("NewSender() returned nil")
	}

	if sender.cfg != cfg {
		t.Error("NewSender() did not store config correctly")
	}
}

func TestSender_Send_Disabled(t *testing.T) {
	cfg := &config.Config{
		Email: config.EmailConfig{
			Enabled:  false,
			SMTPHost: "localhost",
			SMTPPort: 587,
		},
	}

	sender := NewSender(cfg)

	err := sender.Send("test@example.com", "Subject", "Body")
	if err != nil {
		t.Errorf("Send() returned error when email disabled: %v", err)
	}
}

func TestSender_SendHTML_Disabled(t *testing.T) {
	cfg := &config.Config{
		Email: config.EmailConfig{
			Enabled:  false,
			SMTPHost: "localhost",
			SMTPPort: 587,
		},
	}

	sender := NewSender(cfg)

	err := sender.SendHTML("test@example.com", "Subject", "<html><body>Body</body></html>")
	if err != nil {
		t.Errorf("SendHTML() returned error when email disabled: %v", err)
	}
}

func TestSender_SendTLS_Disabled(t *testing.T) {
	cfg := &config.Config{
		Email: config.EmailConfig{
			Enabled:  false,
			SMTPHost: "localhost",
			SMTPPort: 587,
		},
	}

	sender := NewSender(cfg)

	err := sender.SendTLS("test@example.com", "Subject", "Body")
	if err != nil {
		t.Errorf("SendTLS() returned error when email disabled: %v", err)
	}
}

func TestSender_Send_Enabled_NoSMTPServer(t *testing.T) {
	cfg := &config.Config{
		Email: config.EmailConfig{
			Enabled:      true,
			SMTPHost:     "localhost",
			SMTPPort:     6525,
			SMTPUser:     "user",
			SMTPPassword: "password",
			FromAddress:  "from@example.com",
		},
	}

	sender := NewSender(cfg)

	err := sender.Send("test@example.com", "Subject", "Body")
	if err == nil {
		t.Error("Send() should return error when SMTP server is not available")
	}
}

func TestSender_SendHTML_Enabled_NoSMTPServer(t *testing.T) {
	cfg := &config.Config{
		Email: config.EmailConfig{
			Enabled:      true,
			SMTPHost:     "localhost",
			SMTPPort:     6525,
			SMTPUser:     "user",
			SMTPPassword: "password",
			FromAddress:  "from@example.com",
		},
	}

	sender := NewSender(cfg)

	err := sender.SendHTML("test@example.com", "Subject", "<html><body>Body</body></html>")
	if err == nil {
		t.Error("SendHTML() should return error when SMTP server is not available")
	}
}
