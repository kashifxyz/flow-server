package mail

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"
)

type SMTPConfig struct {
	Host       string
	Port       int
	User       string
	Password   string
	From       string
	FromName   string
	Encryption string
}

type SMTPSender struct {
	Config SMTPConfig
}

func (s SMTPSender) Send(ctx context.Context, msg Message) error {
	body := render(msg)
	from := s.Config.From
	headerFrom := from
	if s.Config.FromName != "" {
		headerFrom = fmt.Sprintf("%s <%s>", s.Config.FromName, from)
	}
	payload := strings.Join([]string{
		"From: " + headerFrom,
		"To: " + msg.To,
		"Subject: " + subject(msg),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
	}, "\r\n")

	addr := net.JoinHostPort(s.Config.Host, fmt.Sprintf("%d", s.Config.Port))
	auth := smtp.PlainAuth("", s.Config.User, s.Config.Password, s.Config.Host)

	dialer := &net.Dialer{Timeout: 10 * time.Second}
	if deadline, ok := ctx.Deadline(); ok {
		dialer.Deadline = deadline
	}

	if s.Config.Encryption == "starttls" {
		return smtp.SendMail(addr, auth, from, []string{msg.To}, []byte(payload))
	}

	conn, err := tls.DialWithDialer(dialer, "tcp", addr, &tls.Config{
		ServerName: s.Config.Host,
		MinVersion: tls.VersionTLS12,
	})
	if err != nil {
		return fmt.Errorf("smtp tls dial: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.Config.Host)
	if err != nil {
		return fmt.Errorf("smtp client: %w", err)
	}
	defer func() { _ = client.Close() }()

	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("smtp auth: %w", err)
	}
	if err := client.Mail(from); err != nil {
		return err
	}
	if err := client.Rcpt(msg.To); err != nil {
		return err
	}
	w, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write([]byte(payload)); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return client.Quit()
}
