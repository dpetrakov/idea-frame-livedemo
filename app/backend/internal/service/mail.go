package service

import (
	"context"
	"crypto/tls"
	"encoding/base64"
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
	subject := fmt.Sprintf("%s: одноразовый код для входа в MeetAx Next", code)
	plainBody := fmt.Sprintf("Ваш код для входа в MeetAx Next: %s\r\nДействует %d минут. Если вы не запрашивали вход — игнорируйте это письмо.\r\n\r\nКоманда MeetAx\r\nКанал проекта: https://myteam.mail.ru/profile/AoLJns1GcHktTLbfCmc\r\nПоддержка: https://myteam.mail.ru/profile/AoLJwqc5Q6QgFvo3ED0\r\n", code, s.ttlMin)
	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html lang="ru">
  <body style="margin:0;padding:24px;background:#f6f8fa;font-family:-apple-system,BlinkMacSystemFont,Segoe UI,Roboto,Helvetica,Arial,sans-serif;color:#111827;line-height:1.6;">
    <div style="max-width:560px;margin:0 auto;background:#ffffff;border-radius:12px;padding:24px;box-shadow:0 4px 16px rgba(0,0,0,0.06);">
      <h1 style="margin:0 0 12px;font-size:20px;color:#111827;">Вход в MeetAx Next</h1>
      <p style="margin:0 0 16px;color:#374151;">Ваш одноразовый код для входа. Не сообщайте его никому.</p>
      <div style="display:inline-block;margin:8px 0 16px;padding:12px 16px;border-radius:12px;background:#f3f4f6;border:1px solid #e5e7eb;">
        <span style="font-size:32px;letter-spacing:6px;font-weight:700;color:#111827;">%s</span>
      </div>
      <p style="margin:0 0 12px;color:#374151;">Код действует <strong>%d минут</strong>.</p>
      <p style="margin:0 0 12px;color:#6b7280;">Если вы не запрашивали вход в систему MeetAx Next, просто проигнорируйте это письмо.</p>
      <hr style="border:none;border-top:1px solid #e5e7eb;margin:20px 0;" />
      <p style="margin:0 0 8px;color:#374151;">С уважением,<br/>Команда MeetAx</p>
      <p style="margin:0 0 4px;">
        <a href="https://myteam.mail.ru/profile/AoLJns1GcHktTLbfCmc" style="color:#2563eb;text-decoration:none;">Канал проекта</a>
        <span style="color:#9ca3af;"> · </span>
        <a href="https://myteam.mail.ru/profile/AoLJwqc5Q6QgFvo3ED0" style="color:#2563eb;text-decoration:none;">Поддержка проекта</a>
      </p>
      <p style="margin:8px 0 0;color:#9ca3af;font-size:12px;">Письмо отправлено автоматически, отвечать на него не нужно.</p>
    </div>
  </body>
</html>`, code, s.ttlMin)

	// multipart/alternative: text/plain + text/html, корректная кодировка Subject
	b64 := func(s string) string { return base64.StdEncoding.EncodeToString([]byte(s)) }
	subjectEnc := fmt.Sprintf("=?UTF-8?B?%s?=", b64(subject))
	boundary := fmt.Sprintf("=_MeetAxNext_%s", code)
	msg := "From: " + s.from + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subjectEnc + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: multipart/alternative; boundary=\"" + boundary + "\"\r\n\r\n" +
		"--" + boundary + "\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n\r\n" +
		plainBody + "\r\n" +
		"--" + boundary + "\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n\r\n" +
		htmlBody + "\r\n" +
		"--" + boundary + "--\r\n"

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
