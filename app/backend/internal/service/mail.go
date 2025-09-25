package service

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/smtp"

	"github.com/ideaframe/backend/internal/config"
)

// MailSender интерфейс отправки писем с кодом подтверждения
type MailSender interface {
	SendVerificationCode(ctx context.Context, to string, code string) error
}

// Mailgun удалён: используем только SMTP

// SMTPSender реализация через SMTP
type SMTPSender struct {
	host                  string
	port                  int
	user                  string
	pass                  string
	from                  string
	ttlMin                int
	tlsServerName         string
	tlsInsecureSkipVerify bool
	useSSL                bool
	ehloDomain            string
}

func NewSMTPSender(cfg *config.Config) *SMTPSender {
	return &SMTPSender{
		host:                  cfg.SMTPHost,
		port:                  cfg.SMTPPort,
		user:                  cfg.SMTPUsername,
		pass:                  cfg.SMTPPassword,
		from:                  cfg.SMTPFrom,
		ttlMin:                cfg.EmailCodesTTLMinutes,
		tlsServerName:         cfg.SMTPTLSServerName,
		tlsInsecureSkipVerify: cfg.SMTPTLSInsecureSkipVerify,
		useSSL:                cfg.SMTPUseSSL,
		ehloDomain:            cfg.SMTPEhloDomain,
	}
}

func (s *SMTPSender) SendVerificationCode(ctx context.Context, to string, code string) error {
	if s.host == "" || s.user == "" || s.pass == "" || s.from == "" || s.port == 0 {
		return errors.New("smtp is not configured")
	}
	subject := "Код подтверждения e‑mail"
	body := fmt.Sprintf("Ваш код подтверждения: %s. Он действителен %d минут.", code, s.ttlMin)
	msg := "From: " + s.from + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n\r\n" +
		body

	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	var c *smtp.Client
	var err error
	ehlo := s.ehloDomain
	if ehlo == "" {
		ehlo = "localhost"
	}
	if s.useSSL {
		// Имплицитный TLS (порт 465)
		serverName := s.tlsServerName
		if serverName == "" {
			serverName = s.host
		}
		tlsCfg := &tls.Config{ServerName: serverName, InsecureSkipVerify: s.tlsInsecureSkipVerify}
		conn, dErr := tls.Dial("tcp", addr, tlsCfg)
		if dErr != nil {
			return fmt.Errorf("smtp ssl dial: %w", dErr)
		}
		c, err = smtp.NewClient(conn, s.host)
		if err != nil {
			return fmt.Errorf("smtp new client: %w", err)
		}
		if err := c.Hello(ehlo); err != nil {
			return fmt.Errorf("smtp ehlo: %w", err)
		}
	} else {
		// STARTTLS (порт 587): EHLO -> STARTTLS, без повторного EHLO
		c, err = smtp.Dial(addr)
		if err != nil {
			return fmt.Errorf("smtp dial: %w", err)
		}
		if err := c.Hello(ehlo); err != nil {
			return fmt.Errorf("smtp ehlo: %w", err)
		}
		serverName := s.tlsServerName
		if serverName == "" {
			serverName = s.host
		}
		tlsCfg := &tls.Config{ServerName: serverName, InsecureSkipVerify: s.tlsInsecureSkipVerify}
		if err := c.StartTLS(tlsCfg); err != nil {
			return fmt.Errorf("smtp starttls: %w", err)
		}
		// Не вызываем повторный EHLO
	}
	defer c.Close()
	if ok, _ := c.Extension("AUTH"); ok {
		auth := smtp.PlainAuth("", s.user, s.pass, s.host)
		if err := c.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth: %w", err)
		}
	}
	if err := c.Mail(s.from); err != nil {
		return fmt.Errorf("smtp mail from: %w", err)
	}
	if err := c.Rcpt(to); err != nil {
		return fmt.Errorf("smtp rcpt to: %w", err)
	}
	wc, err := c.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	if _, err := wc.Write([]byte(msg)); err != nil {
		return fmt.Errorf("smtp write: %w", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("smtp data close: %w", err)
	}
	if err := c.Quit(); err != nil {
		return fmt.Errorf("smtp quit: %w", err)
	}
	return nil
}
