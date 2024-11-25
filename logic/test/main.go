package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"log"
	"os"
	"pdm-go-server/internal/auth"
	"time"
)

func genKey() {
	// Generate new key pair
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		log.Fatalf("Failed to generate keys: %v", err)
	}

	// Encode keys to base64
	privBase64 := base64.StdEncoding.EncodeToString(priv)
	pubBase64 := base64.StdEncoding.EncodeToString(pub)

	// Print in .env format
	fmt.Println("Add these lines to your .env file:")
	fmt.Println("----------------------------------------")
	fmt.Printf("JWT_PRIVATE_KEY=%s\n", privBase64)
	fmt.Printf("JWT_PUBLIC_KEY=%s\n", pubBase64)
	fmt.Println("----------------------------------------")

	// Optionally write directly to .env file
	envContent := fmt.Sprintf("JWT_PRIVATE_KEY=%s\nJWT_PUBLIC_KEY=%s\n", privBase64, pubBase64)
	err = os.WriteFile(".env.example", []byte(envContent), 0644)
	if err != nil {
		log.Printf("Warning: Could not write to .env.example: %v", err)
	} else {
		fmt.Println("Keys have been written to .env.example")
	}
}

func main() {
	//genKey()
	// Generate a new key pair
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		log.Fatalf("Failed to generate keys: %v", err)
	}

	// Create auth service
	authService := auth.NewAuthService(priv, pub)
	if err != nil {
		log.Fatalf("Failed to create auth service: %v", err)
	}

	// Generate token
	token, exp, err := authService.GenerateToken("user@example.com")
	if err != nil {
		log.Fatalf("Failed to generate token: %v", err)
	}

	// Convert expiration timestamp to time.Time
	expirationTime := time.Unix(exp, 0)

	// Calculate duration until expiration (for Redis TTL)
	ttl := time.Until(expirationTime)

	log.Printf("Token: %s\n", token)
	log.Printf("Expires at: %s\n", expirationTime.Format(time.RFC3339))
	log.Printf("TTL for Redis: %s\n", ttl)

	// Example of how you would use this with Redis
	/*
	   err = rdb.Set(ctx, "user:token", token, ttl).Err()
	   if err != nil {
	       log.Fatalf("Failed to set token in Redis: %v", err)
	   }
	*/

	// Validate token
	parsedToken, err := authService.ValidateToken(token)
	if err != nil {
		log.Fatalf("Failed to validate token: %v", err)
	}

	claims := parsedToken.Claims.(jwt.MapClaims)

	// Print parsed claims
	parsedExp := time.Unix(int64(claims["exp"].(float64)), 0)
	fmt.Printf("Token expires at: %s\n", parsedExp.Format(time.RFC3339))

	fmt.Printf("Token is valid for email: %s\n", claims["email"])
}
