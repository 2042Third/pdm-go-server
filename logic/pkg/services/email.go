package services

import (
"fmt"
"log"
"net/smtp"
)

func sendEmail(from, to, subject, body string) error {
	// SMTP server configuration
	smtpHost := "smtp.simplelogin.io"
	smtpPort := "587"
	username := "your-simplelogin-email@example.com" // Replace with your SimpleLogin email
	password := "your-simplelogin-api-key"          // Replace with your API key

	// Construct the email headers and body
	message := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s", from, to, subject, body)

	// Connect to the SMTP server
	auth := smtp.PlainAuth("", username, password, smtpHost)

	// Send the email
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{to}, []byte(message))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func main() {
	// Example usage
	from := "alias@simplelogin.io"        // Replace with your alias or custom domain alias
	to := "recipient@example.com"        // Replace with recipient's email
	subject := "Test Email from Go"
	body := "This is a test email sent via SimpleLogin relay from a Go server."

	if err := sendEmail(from, to, subject, body); err != nil {
		log.Fatalf("Error sending email: %v", err)
	}

	log.Println("Email sent successfully!")
}
