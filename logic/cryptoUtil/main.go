package main

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"flag"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"math/big"
	"os"
	"strings"
	"time"
)

const usage = `Usage:
  Generate keys:    %s genkey -alg=<algorithm> -out=<prefix> [-format=std|ssh|tls|armor]
  Sign message:     %s sign -alg=<algorithm> -key=<private-key-file> -msg=<message>
  Verify signature: %s verify -alg=<algorithm> -key=<public-key-file> -msg=<message> -sig=<signature>

Supported algorithms:
  ed25519    Edwards-curve Digital Signature Algorithm (32 bytes)
  rsa256     RSA with SHA-256 (2048 bits)
  rsa384     RSA with SHA-384 (3072 bits)
  rsa512     RSA with SHA-512 (4096 bits)
  es256      ECDSA with P-256 curve
  es384      ECDSA with P-384 curve
  es512      ECDSA with P-521 curve

Formats:
  std        Standard PKCS#8/PKIX PEM format
  ssh        OpenSSH key format
  tls        TLS certificate and key format
  armor      ASCII-armored format (suitable for websites)
`

func main() {
	if len(os.Args) < 2 {
		fmt.Printf(usage, os.Args[0], os.Args[0], os.Args[0])
		os.Exit(1)
	}

	genCmd := flag.NewFlagSet("genkey", flag.ExitOnError)
	genAlg := genCmd.String("alg", "", "Algorithm to use")
	genOut := genCmd.String("out", "key", "Output file prefix")
	genFormat := genCmd.String("format", "std", "Key format: std, ssh, or tls")

	signCmd := flag.NewFlagSet("sign", flag.ExitOnError)
	signAlg := signCmd.String("alg", "", "Algorithm to use")
	signKey := signCmd.String("key", "", "Private key file")
	signMsg := signCmd.String("msg", "", "Message to sign")

	verifyCmd := flag.NewFlagSet("verify", flag.ExitOnError)
	verifyAlg := verifyCmd.String("alg", "", "Algorithm to use")
	verifyKey := verifyCmd.String("key", "", "Public key file")
	verifyMsg := verifyCmd.String("msg", "", "Original message")
	verifySig := verifyCmd.String("sig", "", "Signature to verify (base64)")

	switch os.Args[1] {
	case "genkey":
		genCmd.Parse(os.Args[2:])
		generateKeys(*genAlg, *genOut, *genFormat)
	case "sign":
		signCmd.Parse(os.Args[2:])
		sign(*signAlg, *signKey, *signMsg)
	case "verify":
		verifyCmd.Parse(os.Args[2:])
		verify(*verifyAlg, *verifyKey, *verifyMsg, *verifySig)
	default:
		fmt.Printf(usage, os.Args[0], os.Args[0], os.Args[0])
		os.Exit(1)
	}
}

func generateKeys(alg, outFile, format string) {
	var privKey interface{}
	var pubKey interface{}
	var _ error

	switch strings.ToLower(alg) {
	case "ed25519":
		pub, priv, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			fmt.Printf("Error generating Ed25519 keys: %v\n", err)
			os.Exit(1)
		}
		privKey = priv
		pubKey = pub

	case "rsa256", "rsa384", "rsa512":
		bits := map[string]int{
			"rsa256": 2048,
			"rsa384": 3072,
			"rsa512": 4096,
		}[alg]

		key, err := rsa.GenerateKey(rand.Reader, bits)
		if err != nil {
			fmt.Printf("Error generating RSA keys: %v\n", err)
			os.Exit(1)
		}
		privKey = key
		pubKey = &key.PublicKey

	case "es256", "es384", "es512":
		var curve elliptic.Curve
		switch alg {
		case "es256":
			curve = elliptic.P256()
		case "es384":
			curve = elliptic.P384()
		case "es512":
			curve = elliptic.P521()
		}

		key, err := ecdsa.GenerateKey(curve, rand.Reader)
		if err != nil {
			fmt.Printf("Error generating ECDSA keys: %v\n", err)
			os.Exit(1)
		}
		privKey = key
		pubKey = &key.PublicKey
	}

	switch format {
	case "std":
		saveStandardKeys(privKey, pubKey, outFile)
	case "ssh":
		saveSSHKeys(privKey, pubKey, outFile)
	case "tls":
		saveTLSKeys(privKey, pubKey, outFile)
	case "armor":
		err := saveArmoredKey(privKey, pubKey, alg, outFile)
		if err != nil {
			fmt.Printf("Error saving ASCII-armored keys: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Generated ASCII-armored public key: %s.asc\n", outFile)
		fmt.Printf("Private key backup saved as: %s.private.pem\n", outFile)
	default:
		fmt.Printf("Unsupported format: %s\n", format)
		os.Exit(1)
	}
}

func generateASCIIArmor(pubKey interface{}, alg string) (string, error) {
	// Marshal the public key
	pubBytes, err := x509.MarshalPKIXPublicKey(pubKey)
	if err != nil {
		return "", fmt.Errorf("error encoding public key: %v", err)
	}

	// Create a buffer for the ASCII-armored output
	var buf bytes.Buffer

	// Write the header
	buf.WriteString("-----BEGIN PUBLIC KEY BLOCK-----\n")
	buf.WriteString(fmt.Sprintf("Version: KeyGen v1.0 (%s)\n", alg))
	buf.WriteString(fmt.Sprintf("Comment: Generated on %s\n\n", time.Now().Format(time.RFC3339)))

	// Encode the key data in base64 with line wrapping
	//encoder := base64.NewEncoder(base64.StdEncoding, &bytes.Buffer{})
	b64Data := make([]byte, base64.StdEncoding.EncodedLen(len(pubBytes)))
	base64.StdEncoding.Encode(b64Data, pubBytes)

	// Wrap lines at 64 characters
	for i := 0; i < len(b64Data); i += 64 {
		end := i + 64
		if end > len(b64Data) {
			end = len(b64Data)
		}
		buf.Write(b64Data[i:end])
		buf.WriteByte('\n')
	}

	// Write the footer
	buf.WriteString("-----END PUBLIC KEY BLOCK-----\n")

	return buf.String(), nil
}

func saveArmoredKey(privKey, pubKey interface{}, alg, outFile string) error {
	// Generate ASCII-armored public key
	armoredPub, err := generateASCIIArmor(pubKey, alg)
	if err != nil {
		return err
	}

	// Save public key in ASCII-armored format
	err = ioutil.WriteFile(outFile+".asc", []byte(armoredPub), 0644)
	if err != nil {
		return fmt.Errorf("error saving ASCII-armored public key: %v", err)
	}

	// Save private key in standard PKCS#8 format (for backup)
	privBytes, err := x509.MarshalPKCS8PrivateKey(privKey)
	if err != nil {
		return fmt.Errorf("error encoding private key: %v", err)
	}

	savePEMFile(privBytes, "PRIVATE KEY", outFile+".private.pem")

	return nil
}

func saveStandardKeys(privKey, pubKey interface{}, outFile string) {
	// Save private key in PKCS#8 format
	privBytes, err := x509.MarshalPKCS8PrivateKey(privKey)
	if err != nil {
		fmt.Printf("Error encoding private key: %v\n", err)
		os.Exit(1)
	}
	savePEMFile(privBytes, "PRIVATE KEY", outFile+".private.pem")

	// Save public key in PKIX format
	pubBytes, err := x509.MarshalPKIXPublicKey(pubKey)
	if err != nil {
		fmt.Printf("Error encoding public key: %v\n", err)
		os.Exit(1)
	}
	savePEMFile(pubBytes, "PUBLIC KEY", outFile+".public.pem")
}

func saveSSHKeys(privKey, pubKey interface{}, outFile string) {
	// Convert to SSH format
	sshPubKey, err := ssh.NewPublicKey(pubKey)
	if err != nil {
		fmt.Printf("Error creating SSH public key: %v\n", err)
		os.Exit(1)
	}

	// Save public key in SSH format
	pubBytes := ssh.MarshalAuthorizedKey(sshPubKey)
	err = ioutil.WriteFile(outFile+".pub", pubBytes, 0644)
	if err != nil {
		fmt.Printf("Error saving SSH public key: %v\n", err)
		os.Exit(1)
	}

	// Save private key in SSH format
	var privBytes []byte
	switch k := privKey.(type) {
	case *rsa.PrivateKey:
		privBytes = pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(k),
		})
	case *ecdsa.PrivateKey:
		privBytes, err = x509.MarshalECPrivateKey(k)
		if err != nil {
			fmt.Printf("Error encoding ECDSA private key: %v\n", err)
			os.Exit(1)
		}
		privBytes = pem.EncodeToMemory(&pem.Block{
			Type:  "EC PRIVATE KEY",
			Bytes: privBytes,
		})
	case ed25519.PrivateKey:
		privBytes = pem.EncodeToMemory(&pem.Block{
			Type:  "OPENSSH PRIVATE KEY",
			Bytes: k,
		})
	}
	err = ioutil.WriteFile(outFile, privBytes, 0600)
	if err != nil {
		fmt.Printf("Error saving SSH private key: %v\n", err)
		os.Exit(1)
	}
}

func saveTLSKeys(privKey, pubKey interface{}, outFile string) {
	// Generate self-signed certificate
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Self-signed certificate"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, template, template, pubKey, privKey)
	if err != nil {
		fmt.Printf("Error creating certificate: %v\n", err)
		os.Exit(1)
	}

	// Save certificate
	certOut, err := os.Create(outFile + ".crt")
	if err != nil {
		fmt.Printf("Error creating certificate file: %v\n", err)
		os.Exit(1)
	}
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	certOut.Close()

	// Save private key in PKCS#8 format
	privBytes, err := x509.MarshalPKCS8PrivateKey(privKey)
	if err != nil {
		fmt.Printf("Error encoding private key: %v\n", err)
		os.Exit(1)
	}
	savePEMFile(privBytes, "PRIVATE KEY", outFile+".key")
}

func savePEMFile(bytes []byte, keyType string, filename string) {
	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Error creating key file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	err = pem.Encode(file, &pem.Block{
		Type:  keyType,
		Bytes: bytes,
	})
	if err != nil {
		fmt.Printf("Error writing key file: %v\n", err)
		os.Exit(1)
	}
}

func generateSignatureArmor(signature []byte, alg string, msg string) string {
	var buf bytes.Buffer

	// Write the header
	buf.WriteString("-----BEGIN SIGNATURE-----\n")
	buf.WriteString(fmt.Sprintf("Version: KeyGen v1.0 (%s)\n", alg))
	buf.WriteString(fmt.Sprintf("Comment: Generated on %s\n", time.Now().Format(time.RFC3339)))

	// Add message hash to prevent tampering
	msgHash := sha256.Sum256([]byte(msg))
	buf.WriteString(fmt.Sprintf("Hash: SHA256:%x\n\n", msgHash))

	// Encode signature in base64 with line wrapping
	b64Data := make([]byte, base64.StdEncoding.EncodedLen(len(signature)))
	base64.StdEncoding.Encode(b64Data, signature)

	// Wrap lines at 64 characters
	for i := 0; i < len(b64Data); i += 64 {
		end := i + 64
		if end > len(b64Data) {
			end = len(b64Data)
		}
		buf.Write(b64Data[i:end])
		buf.WriteByte('\n')
	}

	buf.WriteString("-----END SIGNATURE-----\n")
	return buf.String()
}

func parseSignatureArmor(armoredSig string) ([]byte, error) {
	// Split into lines
	lines := strings.Split(armoredSig, "\n")

	// Find the boundaries of the base64 data
	var b64Lines []string
	inData := false

	for _, line := range lines {
		if line == "" {
			if inData {
				continue
			}
			inData = true
			continue
		}
		if strings.HasPrefix(line, "-----BEGIN ") {
			continue
		}
		if strings.HasPrefix(line, "-----END ") {
			break
		}
		if strings.Contains(line, ":") && !inData {
			continue
		}
		if inData {
			b64Lines = append(b64Lines, line)
		}
	}

	// Concatenate and decode base64
	b64Data := strings.Join(b64Lines, "")
	return base64.StdEncoding.DecodeString(b64Data)
}

func sign(alg, keyFile, message string) {
	privKey := loadPrivateKey(alg, keyFile)
	msgBytes := []byte(message)
	var signature []byte
	var err error

	switch strings.ToLower(alg) {
	case "ed25519":
		key := privKey.(ed25519.PrivateKey)
		signature = ed25519.Sign(key, msgBytes)

	case "rsa256", "rsa384", "rsa512":
		key := privKey.(*rsa.PrivateKey)
		hash := getHash(alg)
		h := hash.New()
		h.Write(msgBytes)
		signature, err = rsa.SignPKCS1v15(rand.Reader, key, hash, h.Sum(nil))
		if err != nil {
			fmt.Printf("Error signing message: %v\n", err)
			os.Exit(1)
		}

	case "es256", "es384", "es512":
		key := privKey.(*ecdsa.PrivateKey)
		hash := getHash(alg)
		h := hash.New()
		h.Write(msgBytes)
		r, s, err := ecdsa.Sign(rand.Reader, key, h.Sum(nil))
		if err != nil {
			fmt.Printf("Error signing message: %v\n", err)
			os.Exit(1)
		}
		signature = append(r.Bytes(), s.Bytes()...)

	default:
		fmt.Printf("Unsupported algorithm: %s\n", alg)
		os.Exit(1)
	}

	// Generate ASCII armored signature
	armoredSig := generateSignatureArmor(signature, alg, message)
	fmt.Println(armoredSig)
}

func verify(alg, keyFile, message, armoredSig string) {
	pubKey := loadPublicKey(alg, keyFile)
	msgBytes := []byte(message)

	// Parse ASCII armored signature
	signature, err := parseSignatureArmor(armoredSig)
	if err != nil {
		fmt.Printf("Error decoding signature: %v\n", err)
		os.Exit(1)
	}

	var valid bool
	switch strings.ToLower(alg) {
	case "ed25519":
		key := pubKey.(ed25519.PublicKey)
		valid = ed25519.Verify(key, msgBytes, signature)

	case "rsa256", "rsa384", "rsa512":
		key := pubKey.(*rsa.PublicKey)
		hash := getHash(alg)
		h := hash.New()
		h.Write(msgBytes)
		err = rsa.VerifyPKCS1v15(key, hash, h.Sum(nil), signature)
		valid = err == nil

	case "es256", "es384", "es512":
		key := pubKey.(*ecdsa.PublicKey)
		hash := getHash(alg)
		h := hash.New()
		h.Write(msgBytes)
		sigLen := len(signature)
		r := new(big.Int).SetBytes(signature[:sigLen/2])
		s := new(big.Int).SetBytes(signature[sigLen/2:])
		valid = ecdsa.Verify(key, h.Sum(nil), r, s)

	default:
		fmt.Printf("Unsupported algorithm: %s\n", alg)
		os.Exit(1)
	}

	if valid {
		fmt.Println("Signature is valid")
	} else {
		fmt.Println("Signature is invalid")
		os.Exit(1)
	}
}

func saveKey(keyBytes []byte, keyType string, filename string) {
	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Error creating key file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	err = pem.Encode(file, &pem.Block{
		Type:  keyType,
		Bytes: keyBytes,
	})
	if err != nil {
		fmt.Printf("Error writing key file: %v\n", err)
		os.Exit(1)
	}
}

func loadPrivateKey(alg, filename string) interface{} {
	pemBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading private key file: %v\n", err)
		os.Exit(1)
	}

	block, _ := pem.Decode(pemBytes)
	if block == nil {
		fmt.Println("Error decoding PEM block")
		os.Exit(1)
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		fmt.Printf("Error parsing private key: %v\n", err)
		os.Exit(1)
	}

	return key
}

func loadPublicKey(alg, filename string) interface{} {
	pemBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading public key file: %v\n", err)
		os.Exit(1)
	}

	block, _ := pem.Decode(pemBytes)
	if block == nil {
		fmt.Println("Error decoding PEM block")
		os.Exit(1)
	}

	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		fmt.Printf("Error parsing public key: %v\n", err)
		os.Exit(1)
	}

	return key
}

func getHash(alg string) crypto.Hash {
	switch strings.ToLower(alg) {
	case "rsa256", "es256":
		return crypto.SHA256
	case "rsa384", "es384":
		return crypto.SHA384
	case "rsa512", "es512":
		return crypto.SHA512
	default:
		return crypto.Hash(0)
	}
}
