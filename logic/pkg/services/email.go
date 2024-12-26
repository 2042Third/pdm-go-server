package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"pdm-logic-server/pkg/models"
)

func SendEmail(from, to, subject, body, apiKey string) error {
	url := "https://send.api.mailtrap.io/api/send"

	emailData := models.EmailCall{
		To: []models.EmailAddress{
			{Email: "18604713262my@gmail.com", Name: "Test Name"},
		},
		From: models.EmailAddress{
			Email: "hi@demomailtrap.com",
			Name:  "Test Email",
		},
		Subject:  "Testing Email Sending for PDM Notes",
		Html:     "<p>This is a testing email for PDM Notes.  <strong>PDM Notes</strong>.</p>",
		Text:     "Testing Email",
		Category: "API Test",
	}

	jsonData, err := json.Marshal(emailData)
	if err != nil {
		log.Printf("Failed to marshal payload: %v", err)
		return err
	}

	fmt.Println(string(jsonData))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Api-Token", apiKey)

	fmt.Printf("api key: %s\n ", apiKey)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	// Change this line to use a different variable name
	responseBody, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	fmt.Println(res)
	fmt.Println(string(responseBody)) // Convert byte slice to string for printing

	return nil
}
