package models

type LoginRequest struct {
	Email          string `json:"email" validate:"required,email"`
	Password       string `json:"password" validate:"required"`
	TurnstileToken string `json:"turnstileToken" validate:"required"`
}

type SignupRequest struct {
	Email          string `json:"email" validate:"required,email"`
	Password       string `json:"password" validate:"required"`
	TurnstileToken string `json:"turnstileToken" validate:"required"`
	Name           string `json:"name"`
	Product        string `json:"product"`
}

type SignupInternalResponse struct {
	UserId           string `json:"userId"`
	VerificationCode string `json:"verificationCode"`
}

type VerificationRequest struct {
	Email            string `json:"email" validate:"required"`
	VerificationCode string `json:"code" validate:"required"`
}

type TurnstileResponse struct {
	Success     bool     `json:"success"`
	ChallengeTS string   `json:"challenge_ts"`
	Hostname    string   `json:"hostname"`
	ErrorCodes  []string `json:"error-codes"`
	Action      string   `json:"action"`
	CData       string   `json:"cdata"`
}
