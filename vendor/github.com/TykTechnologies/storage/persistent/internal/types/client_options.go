package types

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"os"
	"time"

	"github.com/TykTechnologies/storage/persistent/internal/helper"
)

const (
	DEFAULT_CONN_TIMEOUT = 10 * time.Second
)

type ClientOpts struct {
	// ConnectionString is the expression used to connect to a storage db server.
	// It contains parameters such as username, hostname, password and port
	ConnectionString string
	// UseSSL is SSL connection is required to connect
	UseSSL bool
	// This setting allows the use of self-signed certificates when connecting to an encrypted storage database.
	SSLInsecureSkipVerify bool
	// Ignore hostname check when it differs from the original (for example with SSH tunneling).
	// The rest of the TLS verification will still be performed
	SSLAllowInvalidHostnames bool
	// Path to the PEM file with trusted root certificates
	SSLCAFile string
	// Path to the PEM file which contains both client certificate and private key. This is required for Mutual TLS.
	SSLPEMKeyfile string
	// Sets the session consistency for the storage connection
	SessionConsistency string
	// Sets the connection timeout to the database. Defaults to 10s.
	ConnectionTimeout int
	// DirectConnection informs whether to establish connections only with the specified seed servers,
	// or to obtain information for the whole cluster and establish connections with further servers too.
	// If true, the client will only connect to the host provided in the ConnectionString
	// and won't attempt to discover other hosts in the cluster. Useful when network restrictions
	// prevent discovery, such as with SSH tunneling. Default is false.
	DirectConnection bool
	// type of database/driver
	Type string
}

// GetTLSConfig returns the TLS config given the configuration specified in ClientOpts. It loads certificates if necessary.
func (opts *ClientOpts) GetTLSConfig() (*tls.Config, error) {
	tlsConfig := &tls.Config{}

	if !opts.UseSSL {
		return tlsConfig, errors.New("error getting tls config when ssl is disabled")
	}

	if opts.SSLInsecureSkipVerify {
		tlsConfig.InsecureSkipVerify = true
	}

	if opts.SSLCAFile != "" {
		if err := opts.loadCACertificates(tlsConfig); err != nil {
			return tlsConfig, err
		}
	}

	if opts.SSLAllowInvalidHostnames {
		tlsConfig.InsecureSkipVerify = true
		opts.verifyPeerCertificate(tlsConfig)
	}

	if opts.SSLPEMKeyfile != "" {
		if err := opts.loadClientCertificates(tlsConfig); err != nil {
			return tlsConfig, err
		}
	}

	return tlsConfig, nil
}

func (opts *ClientOpts) loadCACertificates(tlsConfig *tls.Config) error {
	caCert, err := os.ReadFile(opts.SSLCAFile)
	if err != nil {
		return errors.New("can't load CA certificates:" + err.Error())
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	tlsConfig.RootCAs = caCertPool

	return nil
}

func (opts *ClientOpts) verifyPeerCertificate(tlsConfig *tls.Config) {
	tlsConfig.VerifyPeerCertificate = func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
		certs := make([]*x509.Certificate, len(rawCerts))

		for i, asn1Data := range rawCerts {
			cert, err := x509.ParseCertificate(asn1Data)
			if err != nil {
				return err
			}

			certs[i] = cert
		}

		opts := x509.VerifyOptions{
			Roots:         tlsConfig.RootCAs,
			CurrentTime:   time.Now(),
			DNSName:       "",
			Intermediates: x509.NewCertPool(),
		}

		for i, cert := range certs {
			if i == 0 {
				continue
			}

			opts.Intermediates.AddCert(cert)
		}

		_, err := certs[0].Verify(opts)

		return err
	}
}

func (opts *ClientOpts) loadClientCertificates(tlsConfig *tls.Config) error {
	cert, err := helper.LoadCertificateAndKeyFromFile(opts.SSLPEMKeyfile)
	if err != nil {
		return err
	}

	tlsConfig.Certificates = []tls.Certificate{*cert}

	return nil
}
