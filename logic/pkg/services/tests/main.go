package main

import (
	"log"
	"pdm-logic-server/pkg/services"
)

/**
Used for email testing purposes

*/

func main() {
	// Example usage
	from := "hi@demomailtrap.com"   // Replace with your alias or custom domain alias
	to := "18604713262my@gmail.com" // Replace with recipient's email
	subject := "Test Email from Go"
	body := "This is a test email sent via SimpleLogin relay from a Go server."
	//apiKey := os.Getenv("EMAIL_API_KEY")

	if err := services.SendEmail(from, to, subject, body, "33234b37fe214106c3fa53b49953eb02"); err != nil {
		log.Fatalf("Error sending email: %v", err)
	}

	log.Println("Email sent successfully!")
}
