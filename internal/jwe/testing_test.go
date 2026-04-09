package jwe

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
)

func TestGenerateMockPrivateKey(t *testing.T) {
	cert, err := GenerateMockPrivateKey()

	// Assert that there is no error
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Assert that the certificate is not nil
	if cert == nil {
		t.Fatal("Expected certificate to be non-nil")
	}

	// Assert that the PrivateKey is of the correct type
	if _, ok := cert.PrivateKey.(*rsa.PrivateKey); !ok {
		t.Fatalf("Expected PrivateKey to be of type *rsa.PrivateKey, got %T", cert.PrivateKey)
	}

	// Additional check (optional): Verify the size of the generated private key
	if cert.PrivateKey.(*rsa.PrivateKey).Size() != 256 { // 2048 bits = 256 bytes
		t.Fatalf("Expected private key size to be 256 bytes, got %d", cert.PrivateKey.(*rsa.PrivateKey).Size())
	}
}

func TestCreateJWE(t *testing.T) {
	// Generate a new RSA key pair for testing
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key pair: %v", err)
	}
	publicKey := &privKey.PublicKey

	// Define a sample payload
	payload := []byte("test payload")

	// Call the CreateJWE function
	jwe, err := CreateJWE(payload, publicKey)

	// Assert that there is no error
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Assert that the JWE string is not empty
	if jwe == "" {
		t.Fatal("Expected JWE to be non-empty")
	}

}
