package jwe

import (
	"crypto/rsa"
	"crypto/tls"
	"errors"
	"fmt"

	"github.com/go-jose/go-jose/v4"

	"github.com/markbates/goth/providers/openidConnect"
)

// keyAlgorithms are the RSA key management algorithms accepted for encrypted ID tokens.
// The handler decrypts with an RSA private key, so these are all the algorithms it can
// support. RSA1_5 is kept so IdPs that still use it continue to work.
var keyAlgorithms = []jose.KeyAlgorithm{
	jose.RSA_OAEP,
	jose.RSA_OAEP_256,
	jose.RSA1_5,
}

// contentEncryptions are the content encryption algorithms accepted for encrypted ID tokens.
var contentEncryptions = []jose.ContentEncryption{
	jose.A128GCM,
	jose.A192GCM,
	jose.A256GCM,
	jose.A128CBC_HS256,
	jose.A192CBC_HS384,
	jose.A256CBC_HS512,
}

type Handler struct {
	Enabled            bool             `json:"Enabled"`
	PrivateKeyLocation string           `json:"PrivateKeyLocation"`
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
	jwe, err := jose.ParseEncrypted(token, keyAlgorithms, contentEncryptions)
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
