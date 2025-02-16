package helpers

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"html/template"
	"os"
	"strconv"
	"time"

	"gopkg.in/gomail.v2"
)

type EmailData struct {
	Email string
	OTP   string
	Year  int
}

func SendVerificationEmail(to, otp string) error {
	// Parse template from file
	t, err := template.ParseFiles("templates/email_verification.html")
	if err != nil {
		return err
	}

	// Prepare data
	data := EmailData{
		Email: to,
		OTP:   otp,
		Year:  time.Now().Year(),
	}

	// Execute template
	var body bytes.Buffer
	if err := t.Execute(&body, data); err != nil {
		return err
	}

	EMAIL_FROM := os.Getenv("EMAIL_FROM")

	// Create new message
	m := gomail.NewMessage()
	m.SetHeader("From", EMAIL_FROM)
	m.SetHeader("To", to)
	m.SetHeader("Subject", "Verify Your Aspire Auth Account")
	m.SetBody("text/html", body.String())

	// Parse SMTP port from env
	port, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil {
		return err
	}

	SMTP_HOST := os.Getenv("SMTP_HOST")
	EMAIL_USERNAME := os.Getenv("EMAIL_USERNAME")
	EMAIL_PASSWORD := os.Getenv("EMAIL_PASSWORD")

	// Create dialer
	d := gomail.NewDialer(
		SMTP_HOST,
		port,
		EMAIL_USERNAME,
		EMAIL_PASSWORD,
	)

	fmt.Println(SMTP_HOST, port, EMAIL_USERNAME, EMAIL_PASSWORD, EMAIL_FROM)

	// Set TLS config with InsecureSkipVerify
	d.TLSConfig = &tls.Config{
		InsecureSkipVerify: true,
	}

	// Send email
	return d.DialAndSend(m)
}
