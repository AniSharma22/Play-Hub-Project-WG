package utils

import (
	"fmt"
	"net/smtp"
	"project2/internal/config"
)

// sendEmail sends an email using the provided parameters
func sendEmail(to, subject, body string) error {
	from := "anish22gems@gmail.com"
	appPassword := config.APP_PASSWORD

	// SMTP server configuration
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	// Message body with Subject and Body
	message := []byte("Subject: " + subject + "\r\n\r\n" + body + "\r\n")

	// Authentication
	auth := smtp.PlainAuth("", from, appPassword, smtpHost)

	// Sending email
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{to}, message)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	fmt.Println("Email sent successfully")
	return nil
}

// SendOTPEmail sends an OTP email with the provided OTP
func SendOTPEmail(to, otp string) error {
	subject := "Your OTP Code For Play-Hub"
	body := fmt.Sprintf("Your OTP for verification is: %s. It is valid for 15 minutes.", otp)
	return sendEmail(to, subject, body)
}
