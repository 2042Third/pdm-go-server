package services

var users = map[string]string{
	"user@example.com": "$2a$12$KIXQv6hsEz0BrO6Uihz7fuIwQF4ztR.B21vQSYPlXKrP2Yc5cG5iW", // Password: "password"
}

func ValidateUser(email, password string) bool {
	hashedPassword, ok := users[email]
	if !ok {
		return false
	}
	return password == hashedPassword
}
