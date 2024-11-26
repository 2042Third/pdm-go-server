package main

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"flag"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"log"
	"os"
	"pdm-go-server/internal/auth"
	"time"
)

type KeyPair struct {
	PrivateKey string
	PublicKey  string
	Algorithm  string
}

func generateED25519Keys() (*KeyPair, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ED25519 keys: %v", err)
	}

	return &KeyPair{
		PrivateKey: base64.StdEncoding.EncodeToString(priv),
		PublicKey:  base64.StdEncoding.EncodeToString(pub),
		Algorithm:  "EdDSA",
	}, nil
}

func generateRSAKeys(bits int) (*KeyPair, error) {
	priv, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA keys: %v", err)
	}

	privBytes := x509.MarshalPKCS1PrivateKey(priv)
	pubBytes := x509.MarshalPKCS1PublicKey(&priv.PublicKey)

	return &KeyPair{
		PrivateKey: base64.StdEncoding.EncodeToString(privBytes),
		PublicKey:  base64.StdEncoding.EncodeToString(pubBytes),
		Algorithm:  fmt.Sprintf("RS%d", bits),
	}, nil
}

func generateECDSAKeys(curve elliptic.Curve, name string) (*KeyPair, error) {
	priv, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ECDSA keys: %v", err)
	}

	privBytes, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal ECDSA private key: %v", err)
	}

	pubBytes, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal ECDSA public key: %v", err)
	}

	return &KeyPair{
		PrivateKey: base64.StdEncoding.EncodeToString(privBytes),
		PublicKey:  base64.StdEncoding.EncodeToString(pubBytes),
		Algorithm:  fmt.Sprintf("ES%s", name),
	}, nil
}

func writeToEnvFile(keyPair *KeyPair, writeFile bool) error {
	// Print in .env format
	fmt.Println("Add these lines to your .env file:")
	fmt.Println("----------------------------------------")
	fmt.Printf("JWT_ALGORITHM=%s\n", keyPair.Algorithm)
	fmt.Printf("JWT_PRIVATE_KEY=%s\n", keyPair.PrivateKey)
	fmt.Printf("JWT_PUBLIC_KEY=%s\n", keyPair.PublicKey)
	fmt.Println("----------------------------------------")

	if writeFile {
		envContent := fmt.Sprintf("JWT_ALGORITHM=%s\nJWT_PRIVATE_KEY=%s\nJWT_PUBLIC_KEY=%s\n",
			keyPair.Algorithm, keyPair.PrivateKey, keyPair.PublicKey)

		err := os.WriteFile(".env.example", []byte(envContent), 0644)
		if err != nil {
			return fmt.Errorf("could not write to .env.example: %v", err)
		}
		fmt.Println("Keys have been written to .env.example")
	}

	return nil
}

func main() {
	// Define command line flags
	writeFlag := flag.Bool("w", false, "Write keys to .env.example file")
	algoFlag := flag.String("algo", "ed25519", "Algorithm to use (ed25519, rsa256, rsa384, rsa512, es256, es384, es512)")
	flag.Parse()

	var keyPair *KeyPair
	var err error

	// Generate keys based on selected algorithm
	switch *algoFlag {
	case "ed25519":
		keyPair, err = generateED25519Keys()
	case "rsa256":
		keyPair, err = generateRSAKeys(2048) // RSA-256 typically uses 2048-bit keys
	case "rsa384":
		keyPair, err = generateRSAKeys(3072) // RSA-384 typically uses 3072-bit keys
	case "rsa512":
		keyPair, err = generateRSAKeys(4096) // RSA-512 typically uses 4096-bit keys
	case "es256":
		keyPair, err = generateECDSAKeys(elliptic.P256(), "256")
	case "es384":
		keyPair, err = generateECDSAKeys(elliptic.P384(), "384")
	case "es512":
		keyPair, err = generateECDSAKeys(elliptic.P521(), "512")
	default:
		log.Fatalf("Unsupported algorithm: %s", *algoFlag)
	}

	if err != nil {
		log.Fatalf("Failed to generate keys: %v", err)
	}

	if err := writeToEnvFile(keyPair, *writeFlag); err != nil {
		log.Fatalf("Failed to write keys: %v", err)
	}
}

func testing() {
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
