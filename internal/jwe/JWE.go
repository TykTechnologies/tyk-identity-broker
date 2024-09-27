package jwe

import (
	"crypto/rsa"
	"crypto/tls"
	"encoding/json"
	"fmt"
	jose "gopkg.in/square/go-jose.v2"
)

type Config struct {
	Enabled            bool             `json:"enabled"`
	PrivateKeyLocation string           `json:"private_key_location"`
	Key                *tls.Certificate `json:"-"`
}

func Encrypt(token string) string {
	// Convert payload to JSON
	payloadBytes, err := json.Marshal(token)
	if err != nil {
		return ""
	}

	// Create an encrypter
	encrypter, err := jose.NewEncrypter(
		jose.A256GCM, // Content encryption algorithm
		jose.Recipient{
			Algorithm: jose.RSA_OAEP_256, // Key encryption algorithm
			Key:       getPublicKey(),
		},
		(&jose.EncrypterOptions{}).WithType("JWT"), // Optional: set the "typ" header to "JWT"
	)
	if err != nil {
		return ""
	}

	// Encrypt the payload
	jwe, err := encrypter.Encrypt(payloadBytes)
	if err != nil {
		fmt.Printf("Error encrypting payload: %v\n", err)
		return ""
	}

	// Serialize the encrypted token
	serialized, err := jwe.CompactSerialize()
	if err != nil {
		fmt.Printf("Error serializing JWE: %v\n", err)
		return ""
	}

	fmt.Printf("Encrypted token: %s\n", serialized)
	return serialized
}

func Decrypt(tokenString string, privateKey *rsa.PrivateKey) ([]byte, error) {
	// Parse the serialized token
	jwe, err := jose.ParseEncrypted(tokenString)
	if err != nil {
		return nil, fmt.Errorf("error parsing JWE: %v", err)
	}

	// Decrypt the token
	decrypted, err := jwe.Decrypt(privateKey)
	if err != nil {
		return nil, fmt.Errorf("error decrypting JWE: %v", err)
	}

	return decrypted, nil
}
