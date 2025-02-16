package helpers

import (
	"bytes"
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

	// Create new message
	m := gomail.NewMessage()
	m.SetHeader("From", os.Getenv("EMAIL_FROM"))
	m.SetHeader("To", to)
	m.SetHeader("Subject", "Verify Your Aspire Auth Account")
	m.SetBody("text/html", body.String())

	// Optional: Add logo attachment if needed
	// m.Embed("path/to/logo.png")

	// Parse SMTP port from env
	port, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil {
		return err
	}

	// Create dialer
	d := gomail.NewDialer(
		os.Getenv("SMTP_HOST"),
		port,
		os.Getenv("EMAIL_FROM"),
		os.Getenv("EMAIL_PASSWORD"),
	)

	// Send email
	return d.DialAndSend(m)
}
