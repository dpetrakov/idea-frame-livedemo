package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strconv"
)

func main() {
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	user := os.Getenv("SMTP_USERNAME")
	pass := os.Getenv("SMTP_PASSWORD")
	from := os.Getenv("SMTP_FROM")
	to := os.Getenv("SMTP_TEST_TO")
	if host == "" || port == "" || user == "" || pass == "" || from == "" || to == "" {
		log.Fatalf("Missing env vars. Required: SMTP_HOST, SMTP_PORT, SMTP_USERNAME, SMTP_PASSWORD, SMTP_FROM, SMTP_TEST_TO")
	}

	tlsServerName := os.Getenv("SMTP_TLS_SERVER_NAME")
	if tlsServerName == "" {
		tlsServerName = host
	}
	insecureStr := os.Getenv("SMTP_TLS_INSECURE_SKIP_VERIFY")
	insecure, _ := strconv.ParseBool(insecureStr)
	useSSL, _ := strconv.ParseBool(os.Getenv("SMTP_USE_SSL"))

	addr := host + ":" + port
	subject := "SMTP test from MeetAx Next"
	body := "Hello! This is SMTP test message."
	msg := "From: " + from + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n\r\n" +
		body

	var c *smtp.Client
	var err error
	if useSSL {
		// Implicit TLS (port 465)
		tlsCfg := &tls.Config{ServerName: tlsServerName, InsecureSkipVerify: insecure}
		conn, dErr := tls.Dial("tcp", addr, tlsCfg)
		if dErr != nil {
			log.Fatalf("smtp ssl dial: %v", dErr)
		}
		c, err = smtp.NewClient(conn, host)
		if err != nil {
			log.Fatalf("smtp new client: %v", err)
		}
		if err := c.Hello("localhost"); err != nil {
			log.Fatalf("smtp ehlo: %v", err)
		}
	} else {
		// STARTTLS (port 587)
		c, err = smtp.Dial(addr)
		if err != nil {
			log.Fatalf("smtp dial: %v", err)
		}
		if err := c.Hello("localhost"); err != nil {
			log.Fatalf("smtp ehlo: %v", err)
		}
		if ok, _ := c.Extension("STARTTLS"); ok {
			tlsCfg := &tls.Config{ServerName: tlsServerName, InsecureSkipVerify: insecure}
			if err := c.StartTLS(tlsCfg); err != nil {
				log.Fatalf("smtp starttls: %v", err)
			}
			// Не выполняем повторный EHLO — поведение выровнено с продакшн-кодом отправки писем
		}
	}
	if ok, _ := c.Extension("AUTH"); ok {
		auth := smtp.PlainAuth("", user, pass, host)
		if err := c.Auth(auth); err != nil {
			log.Fatalf("smtp auth: %v", err)
		}
	}
	if err := c.Mail(from); err != nil {
		log.Fatalf("smtp mail from: %v", err)
	}
	if err := c.Rcpt(to); err != nil {
		log.Fatalf("smtp rcpt to: %v", err)
	}
	wc, err := c.Data()
	if err != nil {
		log.Fatalf("smtp data: %v", err)
	}
	if _, err := wc.Write([]byte(msg)); err != nil {
		log.Fatalf("smtp write: %v", err)
	}
	if err := wc.Close(); err != nil {
		log.Fatalf("smtp data close: %v", err)
	}
	if err := c.Quit(); err != nil {
		log.Fatalf("smtp quit: %v", err)
	}
	fmt.Println("SMTP test message sent successfully")
}
