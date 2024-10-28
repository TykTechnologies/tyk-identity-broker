package jwe

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"github.com/go-jose/go-jose/v3"
)

func GenerateMockPrivateKey() (*tls.Certificate, error) {
	// Generate a new RSA private key for testing
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	cert := &tls.Certificate{
		PrivateKey: privKey,
	}
	return cert, nil
}

func CreateJWE(payload []byte, recipient *rsa.PublicKey) (string, error) {

	encrypter, err := jose.NewEncrypter(
		jose.A256GCM,
		jose.Recipient{
			Algorithm: jose.RSA_OAEP_256,
			Key:       recipient,
		},
		(&jose.EncrypterOptions{}).WithType("JWT"))
	if err != nil {
		return "", err
	}
	jwe, err := encrypter.Encrypt(payload)
	if err != nil {
		return "", err
	}
	return jwe.CompactSerialize()
}
