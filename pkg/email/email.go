package email

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
)

type Config struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

type Sender struct {
	cfg Config
}

func NewSender(cfg Config) *Sender {
	return &Sender{cfg: cfg}
}

func (s *Sender) Send(to, subject, body string) error {
	addr := net.JoinHostPort(s.cfg.Host, s.cfg.Port)

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("smtp dial: %w", err)
	}

	client, err := smtp.NewClient(conn, s.cfg.Host)
	if err != nil {
		return fmt.Errorf("smtp client: %w", err)
	}
	defer client.Close()

	if err = client.StartTLS(&tls.Config{
		ServerName:         s.cfg.Host,
		InsecureSkipVerify: true,
	}); err != nil {
		return fmt.Errorf("starttls: %w", err)
	}

	auth := smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)
	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("smtp auth: %w", err)
	}

	if err = client.Mail(s.cfg.From); err != nil {
		return fmt.Errorf("smtp mail from: %w", err)
	}
	if err = client.Rcpt(to); err != nil {
		return fmt.Errorf("smtp rcpt: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}

	msg := fmt.Sprintf(
		"From: Oleron <%s>\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=utf-8\r\n\r\n%s",
		s.cfg.From, to, subject, body,
	)
	if _, err = fmt.Fprint(w, msg); err != nil {
		return fmt.Errorf("smtp write: %w", err)
	}

	return w.Close()
}
