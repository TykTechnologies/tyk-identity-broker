package helper

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
)

func LoadCertificateAndKey(data []byte) (*tls.Certificate, error) {
	var cert tls.Certificate
	var err error
	for {
		block, rest := pem.Decode(data)
		if block == nil {
			break
		}
		if block.Type == "CERTIFICATE" {
			cert.Certificate = append(cert.Certificate, block.Bytes)
		} else {
			cert.PrivateKey, err = parsePrivateKey(block.Bytes)
			if err != nil {
				return nil, err
			}
		}

		data = rest
	}

	if len(cert.Certificate) == 0 {
		return nil, fmt.Errorf("no certificate found")
	} else if cert.PrivateKey == nil {
		return nil, fmt.Errorf("no private key found")
	}

	return &cert, nil
}

func LoadCertificateAndKeyFromFile(path string) (*tls.Certificate, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.New("failure reading certificate file: " + err.Error())
	}

	return LoadCertificateAndKey(raw)
}

func parsePrivateKey(der []byte) (crypto.PrivateKey, error) {
	if key, err := x509.ParsePKCS1PrivateKey(der); err == nil {
		return key, nil
	}
	if key, err := x509.ParsePKCS8PrivateKey(der); err == nil {
		switch key := key.(type) {
		case *rsa.PrivateKey, *ecdsa.PrivateKey:
			return key, nil
		default:
			return nil, fmt.Errorf("found unknown private key type in PKCS#8 wrapping")
		}
	}
	if key, err := x509.ParseECPrivateKey(der); err == nil {
		return key, nil
	}
	return nil, fmt.Errorf("failed to parse private key")
}
