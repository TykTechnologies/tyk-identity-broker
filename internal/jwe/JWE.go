package jwe

import (
	"crypto/rsa"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/markbates/goth/providers/openidConnect"
	jose "gopkg.in/square/go-jose.v2"
)

type Handler struct {
	Enabled            bool             `json:"enabled"`
	PrivateKeyLocation string           `json:"private_key_location"`
	Key                *tls.Certificate `json:"-"`
}

func Encrypt(token string) (string, error) {
	// Convert payload to JSON
	payloadBytes, err := json.Marshal(token)
	if err != nil {
		return "", err
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
		return "", err
	}

	// Encrypt the payload
	jwe, err := encrypter.Encrypt(payloadBytes)
	if err != nil {
		return "", err
	}

	// Serialize the encrypted token
	serialized, err := jwe.CompactSerialize()
	if err != nil {
		return "", err
	}

	return serialized, nil
}

func (handler *Handler) Decrypt(token string) (string, error) {
	if !handler.Enabled {
		return token, nil
	}

	if handler.Key == nil {
		return "", errors.New("JWE Private Key not loaded")
	}

	privateKey := handler.Key.PrivateKey.(*rsa.PrivateKey)

	// Parse the serialized token
	jwe, err := jose.ParseEncrypted(token)
	if err != nil {
		return "", fmt.Errorf("error parsing JWE: %v", err)
	}

	// Decrypt the token
	decrypted, err := jwe.Decrypt(privateKey)
	if err != nil {
		return "", fmt.Errorf("error decrypting JWE: %v", err)
	}

	return string(decrypted), nil
}

func DecryptIDToken(jweHandler *Handler, JWTSession *openidConnect.Session) error {
	decryptedIDToken, err := jweHandler.Decrypt(JWTSession.IDToken)
	if err != nil {
		return err
	}
	JWTSession.IDToken = decryptedIDToken
	return nil
}
