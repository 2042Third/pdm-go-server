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
	"log"
	"os"
	"strings"
)

type KeyPair struct {
	PrivateKey string
	PublicKey  string
	Algorithm  string
}

type AlgorithmInfo struct {
	Name        string
	Description string
	KeySize     string
}

var supportedAlgos = []AlgorithmInfo{
	{"ed25519", "Edwards-curve Digital Signature Algorithm", "32 bytes"},
	{"rsa256", "RSA with SHA-256", "2048 bits"},
	{"rsa384", "RSA with SHA-384", "3072 bits"},
	{"rsa512", "RSA with SHA-512", "4096 bits"},
	{"es256", "ECDSA with P-256 curve", "256 bits"},
	{"es384", "ECDSA with P-384 curve", "384 bits"},
	{"es512", "ECDSA with P-521 curve", "521 bits"},
}

func showHelp() {
	fmt.Println("JWT Key Pair Generator")
	fmt.Println("\nUsage:")
	fmt.Println("  keygen [-w] -algo <algorithm>")
	fmt.Println("\nFlags:")
	fmt.Println("  -w\t\tWrite keys to .env.example file")
	fmt.Println("  -algo string\tSpecify the algorithm to use")
	fmt.Println("\nSupported Algorithms:")
	fmt.Println("----------------------------------------")
	fmt.Printf("%-10s %-30s %s\n", "NAME", "DESCRIPTION", "KEY SIZE")
	fmt.Println("----------------------------------------")
	for _, algo := range supportedAlgos {
		fmt.Printf("%-10s %-30s %s\n", algo.Name, algo.Description, algo.KeySize)
	}
	fmt.Println("----------------------------------------")
	fmt.Println("\nExamples:")
	fmt.Println("  keygen -algo ed25519")
	fmt.Println("  keygen -w -algo rsa256")
	fmt.Println("  keygen -algo es384 -w")
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
	fmt.Println("\nGenerated Key Pair:")
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
		fmt.Println("\nKeys have been written to .env.example")
	}

	return nil
}

func isValidAlgorithm(algo string) bool {
	for _, supported := range supportedAlgos {
		if supported.Name == strings.ToLower(algo) {
			return true
		}
	}
	return false
}

func main() {
	// Define command line flags
	writeFlag := flag.Bool("w", false, "Write keys to .env.example file")
	algoFlag := flag.String("algo", "", "Algorithm to use (ed25519, rsa256, rsa384, rsa512, es256, es384, es512)")

	flag.Parse()

	// Show help if no algorithm is specified
	if *algoFlag == "" {
		showHelp()
		os.Exit(0)
	}

	// Validate algorithm choice
	if !isValidAlgorithm(*algoFlag) {
		fmt.Printf("Error: Unsupported algorithm '%s'\n\n", *algoFlag)
		showHelp()
		os.Exit(1)
	}

	var keyPair *KeyPair
	var err error

	// Generate keys based on selected algorithm
	switch strings.ToLower(*algoFlag) {
	case "ed25519":
		keyPair, err = generateED25519Keys()
	case "rsa256":
		keyPair, err = generateRSAKeys(2048)
	case "rsa384":
		keyPair, err = generateRSAKeys(3072)
	case "rsa512":
		keyPair, err = generateRSAKeys(4096)
	case "es256":
		keyPair, err = generateECDSAKeys(elliptic.P256(), "256")
	case "es384":
		keyPair, err = generateECDSAKeys(elliptic.P384(), "384")
	case "es512":
		keyPair, err = generateECDSAKeys(elliptic.P521(), "512")
	}

	if err != nil {
		log.Fatalf("Failed to generate keys: %v", err)
	}

	if err := writeToEnvFile(keyPair, *writeFlag); err != nil {
		log.Fatalf("Failed to write keys: %v", err)
	}
}
