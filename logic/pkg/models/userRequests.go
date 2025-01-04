package models

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type SignupRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
	Name     string `json:"name"`
	Product  string `json:"product"`
}

type SignupInternalResponse struct {
	UserId           uint   `json:"userId"`
	VerificationCode string `json:"verificationCode"`
}

type VerificationRequest struct {
	Email            string `json:"email" validate:"required"`
	VerificationCode string `json:"code" validate:"required"`
}
