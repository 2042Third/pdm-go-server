package models

type EmailVerificationTemplateData struct {
	Code  string
	Email string
}

type EmailAddress struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type Attachment struct {
	Content     string `json:"content"`
	Filename    string `json:"filename"`
	Type        string `json:"type"`
	Disposition string `json:"disposition"`
}

type CustomVariables struct {
	App       string `json:"app,omitempty"`
	EmailType string `json:"email_type,omitempty"`
	UserId    string `json:"user_id,omitempty"`
	BatchId   string `json:"batch_id,omitempty"`
}

type Headers struct {
	XMessageSource        string `json:"X-Message-Source,omitempty"`
	XMailer               string `json:"X-Mailer,omitempty"`
	Precedence            string `json:"Precedence,omitempty"`
	XAutoResponseSuppress string `json:"X-Auto-Response-Suppress,omitempty"`
	ListUnsubscribe       string `json:"List-Unsubscribe,omitempty"`
}

type EmailCall struct {
	To              []EmailAddress   `json:"to"`
	Cc              []EmailAddress   `json:"cc,omitempty"`
	Bcc             []EmailAddress   `json:"bcc,omitempty"`
	From            EmailAddress     `json:"from"`
	ReplyTo         *EmailAddress    `json:"reply_to,omitempty"`
	Attachments     []Attachment     `json:"attachments,omitempty"`
	CustomVariables *CustomVariables `json:"custom_variables,omitempty"`
	Headers         *Headers         `json:"headers,omitempty"`
	Subject         string           `json:"subject"`
	Html            string           `json:"html"`
	Text            string           `json:"text"`
	Category        string           `json:"category"`
}
