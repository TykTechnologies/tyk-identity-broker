package jwe

import (
	"crypto/rsa"
	"crypto/tls"
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
