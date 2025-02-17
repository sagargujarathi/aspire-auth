package helpers

import (
	"aspire-auth/internal/config"
	"bytes"
	"crypto/tls"
	"html/template"
	"time"

	"gopkg.in/gomail.v2"
)

type EmailData struct {
	Email string
	OTP   string
	Year  int
}

func SendVerificationEmail(to, otp string, config *config.Config) error {
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

	// Create new message
	m := gomail.NewMessage()
	m.SetHeader("From", config.Email.From)
	m.SetHeader("To", to)
	m.SetHeader("Subject", "Verify Your Aspire Auth Account")
	m.SetBody("text/html", body.String())

	// Create dialer
	d := gomail.NewDialer(
		config.Email.Host,
		config.Email.Port,
		config.Email.Username,
		config.Email.Password,
	)

	// Set TLS config with InsecureSkipVerify
	d.TLSConfig = &tls.Config{
		InsecureSkipVerify: true,
	}

	// Send email
	return d.DialAndSend(m)
}
