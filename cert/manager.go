package cert

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"strings"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"

	"github.com/TykTechnologies/tyk-identity-broker/providers"
)

const (
	cacheDefaultTTL    = 300 // 5 minutes.
	cacheCleanInterval = 600 // 10 minutes.
)

type (
	// CertificateManager interface defines public certificate manager functions.
	CertificateManager interface {
		ListAny(ids []string) []*tls.Certificate
	}

	// manager implements CertificateManager interface.
	manager struct {
		store  providers.FileLoader
		logger *logrus.Entry
		cache  *cache.Cache
	}
)

var (
	logger *logrus.Logger
)

// NewCertificateManager func creates and returns default Tyk Identity Broker certificate manager,
// with FileLoader storage.
func NewCertificateManager() CertificateManager {
	if logger == nil {
		logger = logrus.New()
	}

	expirationInterval := time.Duration(cacheDefaultTTL) * time.Second
	cleanupInterval := time.Duration(cacheCleanInterval) * time.Second

	m := manager{
		store:  providers.FileLoader{},
		logger: logger.WithFields(logrus.Fields{"prefix": "cert_storage"}),
		cache:  cache.New(expirationInterval, cleanupInterval),
	}

	return &m
}

// ListAny func returns list of all requested certificates of any kind.
func (m *manager) ListAny(ids []string) []*tls.Certificate {
	certs := make([]*tls.Certificate, 0)

	for _, id := range ids {
		// Read certificate from cache.
		if cert, found := m.cache.Get(id); found {
			certs = append(certs, cert.(*tls.Certificate))

			continue
		}

		// Read certificate from FileLoader.
		val, err := m.store.GetKey("raw-" + id)

		cert, err := parsePEMCertificate([]byte(val))
		if err != nil {
			m.logger.Error("error while parsing certificate: ", id, " ", err)
			m.logger.Debug("failed certificate: ", val)

			certs = append(certs, nil)

			continue
		}

		// Write certificate to cache.
		m.cache.Set(id, cert, cache.DefaultExpiration)

		certs = append(certs, cert)
	}

	return certs
}

func parsePEM(data []byte) ([]*pem.Block, error) {
	var pemBlocks []*pem.Block

	for {
		var block *pem.Block

		block, data = pem.Decode(data)
		if block == nil {
			break
		}

		pemBlocks = append(pemBlocks, block)
	}

	return pemBlocks, nil
}

func parsePEMCertificate(data []byte) (*tls.Certificate, error) {
	var cert tls.Certificate

	blocks, err := parsePEM(data)
	if err != nil {
		return nil, err
	}

	var certID string

	for _, block := range blocks {
		if block.Type == "CERTIFICATE" {
			certID = hexSHA256(block.Bytes)

			cert.Certificate = append(cert.Certificate, block.Bytes)

			continue
		}

		if strings.HasSuffix(block.Type, "PRIVATE KEY") {
			cert.PrivateKey, err = parsePrivateKey(block.Bytes)
			if err != nil {
				return nil, err
			}
			continue
		}

		if block.Type == "PUBLIC KEY" {
			cert.Certificate = append(cert.Certificate, block.Bytes)
			cert.Leaf = &x509.Certificate{
				Subject: pkix.Name{
					CommonName: "Public Key: " + hexSHA256(block.Bytes),
				},
			}
		}
	}

	if len(cert.Certificate) == 0 {
		return nil, errors.New("can't find CERTIFICATE block")
	}

	if cert.Leaf == nil {
		cert.Leaf, err = x509.ParseCertificate(cert.Certificate[0])
		if err != nil {
			return nil, err
		}
	}

	// Cache certificate fingerprint.
	ext := pkix.Extension{
		Value: []byte(certID),
	}

	cert.Leaf.Extensions = append([]pkix.Extension{ext}, cert.Leaf.Extensions...)

	return &cert, nil
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
			return nil, errors.New("tls: found unknown private key type in PKCS#8 wrapping")
		}
	}

	if key, err := x509.ParseECPrivateKey(der); err == nil {
		return key, nil
	}

	return nil, errors.New("tls: failed to parse private key")
}

func hexSHA256(cert []byte) string {
	certSHA := sha256.Sum256(cert)

	return hex.EncodeToString(certSHA[:])
}
