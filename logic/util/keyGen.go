package util

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

func KeyGen() {
	// Generate Ed25519 key pair
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic(fmt.Sprintf("failed to generate keys: %v", err))
	}

	// Print keys in base64 format for storage
	fmt.Println("Public Key:", base64.StdEncoding.EncodeToString(publicKey))
	fmt.Println("Private Key:", base64.StdEncoding.EncodeToString(privateKey))
}
