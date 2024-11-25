package auth

import (
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type AuthService struct {
	PrivateKey ed25519.PrivateKey
	PublicKey  ed25519.PublicKey
}

// LoadKeys loads Ed25519 keys from environment variables
func LoadKeys() (ed25519.PrivateKey, ed25519.PublicKey, error) {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		return nil, nil, fmt.Errorf("error loading .env file: %w", err)
	}

	// Get encoded keys from environment
	privKeyStr := os.Getenv("JWT_PRIVATE_KEY")
	pubKeyStr := os.Getenv("JWT_PUBLIC_KEY")

	if privKeyStr == "" || pubKeyStr == "" {
		return nil, nil, fmt.Errorf("JWT keys not found in environment")
	}

	// Decode private key
	privKey, err := base64.StdEncoding.DecodeString(privKeyStr)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode private key: %w", err)
	}

	// Decode public key
	pubKey, err := base64.StdEncoding.DecodeString(pubKeyStr)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode public key: %w", err)
	}

	// Verify key sizes
	if len(privKey) != ed25519.PrivateKeySize {
		return nil, nil, fmt.Errorf("invalid private key size")
	}
	if len(pubKey) != ed25519.PublicKeySize {
		return nil, nil, fmt.Errorf("invalid public key size")
	}

	return ed25519.PrivateKey(privKey), ed25519.PublicKey(pubKey), nil
}

func NewAuthService(privateKey ed25519.PrivateKey, publicKey ed25519.PublicKey) *AuthService {
	if privateKey == nil || publicKey == nil {
		panic("Ed25519 keys must be set")
	}
	return &AuthService{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
	}
}

// GenerateToken generates a new JWT token with provided claims using Ed25519.
func (a *AuthService) GenerateToken(email string) (string, int64, error) {
	token := jwt.New(jwt.SigningMethodEdDSA) // Use Ed25519 for signing
	claims := token.Claims.(jwt.MapClaims)
	expiration := time.Now().Add(24 * time.Hour).Unix() // Adjust token validity duration as needed

	claims["email"] = email
	claims["exp"] = expiration
	claims["iat"] = time.Now().Unix()

	// Sign the token with the private key
	t, err := token.SignedString(a.PrivateKey)
	if err != nil {
		return "", 0, fmt.Errorf("failed to sign token: %w", err)
	}

	return t, expiration, nil
}

// ValidateToken validates the provided JWT token using Ed25519.
func (a *AuthService) ValidateToken(tokenStr string) (*jwt.Token, error) {
	return jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		// Ensure the token uses Ed25519 signing method
		if _, ok := token.Method.(*jwt.SigningMethodEd25519); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return a.PublicKey, nil
	})
}

// HealthCheck performs a sanity check of the JWT system by generating and validating a test token
func (a *AuthService) HealthCheck() error {
	testEmail := "checker@check.checker"
	log.Printf("Running JWT system health check with test email: %s", testEmail)

	// Try to generate a token
	token, exp, err := a.GenerateToken(testEmail)
	if err != nil {
		return fmt.Errorf("health check failed - token generation error: %w", err)
	}
	log.Printf("Test token generated successfully, expires at: %d", exp)

	// Try to validate the token
	parsedToken, err := a.ValidateToken(token)
	if err != nil {
		return fmt.Errorf("health check failed - token validation error: %w", err)
	}

	// Verify claims
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return fmt.Errorf("health check failed - could not parse claims")
	}

	email, ok := claims["email"].(string)
	if !ok || email != testEmail {
		return fmt.Errorf("health check failed - email claim mismatch: got %v, want %s", claims["email"], testEmail)
	}

	log.Printf("JWT system health check passed - token generation and validation working correctly")
	return nil
}
