package email

import (
	"context"
	"fmt"
	"net"
	"net/smtp"
	"strings"
)

type Message struct {
	To      string
	Subject string
	Body    string
}

type Sender interface {
	Enabled() bool
	Send(context.Context, Message) error
}

type SMTPSender struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

func NewSMTPSender(host string, port int, username, password, from string) *SMTPSender {
	return &SMTPSender{Host: host, Port: port, Username: username, Password: password, From: from}
}

func (s *SMTPSender) Enabled() bool {
	return strings.TrimSpace(s.Host) != "" && strings.TrimSpace(s.From) != ""
}

func (s *SMTPSender) Send(ctx context.Context, msg Message) error {
	if !s.Enabled() {
		return fmt.Errorf("smtp email is not configured")
	}
	if strings.TrimSpace(msg.To) == "" {
		return fmt.Errorf("email recipient is required")
	}
	addr := net.JoinHostPort(s.Host, fmt.Sprintf("%d", s.Port))
	headers := map[string]string{
		"From":         s.From,
		"To":           msg.To,
		"Subject":      msg.Subject,
		"MIME-Version": "1.0",
		"Content-Type": "text/plain; charset=UTF-8",
	}
	var raw strings.Builder
	for key, value := range headers {
		raw.WriteString(key)
		raw.WriteString(": ")
		raw.WriteString(value)
		raw.WriteString("\r\n")
	}
	raw.WriteString("\r\n")
	raw.WriteString(msg.Body)
	raw.WriteString("\r\n")

	var auth smtp.Auth
	if s.Username != "" || s.Password != "" {
		auth = smtp.PlainAuth("", s.Username, s.Password, s.Host)
	}
	errCh := make(chan error, 1)
	go func() {
		errCh <- smtp.SendMail(addr, auth, s.From, []string{msg.To}, []byte(raw.String()))
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errCh:
		return err
	}
}
