package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"pdm-logic-server/pkg/models"
	"pdm-logic-server/templates"
)

func SendEmail(from, to, subject, body, verificationCode, apiKey string) error {
	url := "https://send.api.mailtrap.io/api/send"

	tmpl, err := template.New("verification").Parse(templates.EmailTemplate)
	if err != nil {
		return err
	}

	var htmlBuffer bytes.Buffer
	data := models.EmailVerificationTemplateData{
		Code: verificationCode,
	}

	err = tmpl.Execute(&htmlBuffer, data)
	if err != nil {
		return err
	}

	emailData := models.EmailCall{
		To: []models.EmailAddress{
			{Email: to, Name: ""},
		},
		From: models.EmailAddress{
			Email: from,
			Name:  "PDM Notes",
		},
		Subject:  subject,
		Html:     htmlBuffer.String(),
		Text:     fmt.Sprintf("Your PDM Notes verification code is: %s", data.Code),
		Category: "PDM Notes Email Verification",
		// Add these headers to improve deliverability
		Headers: &models.Headers{
			XMessageSource:        "pdm.pw", // Your domain
			XMailer:               "PDM Notes",
			Precedence:            "transactional",
			XAutoResponseSuppress: "OOF, AutoReply",
			ListUnsubscribe:       fmt.Sprintf("<mailto:unsubscribe@%s>", "pdm.pw"),
		},
		CustomVariables: &models.CustomVariables{
			App:       "PDM Notes",
			EmailType: "Verification",
		},
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

	//fmt.Printf("api key: %s\n ", apiKey)

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
