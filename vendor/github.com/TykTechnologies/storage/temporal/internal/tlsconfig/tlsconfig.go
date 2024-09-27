package tlsconfig

import (
	"crypto/tls"
	"crypto/x509"
	"os"

	"github.com/TykTechnologies/storage/temporal/model"
	"github.com/TykTechnologies/storage/temporal/temperr"
)

func HandleTLS(cfg *model.TLS) (*tls.Config, error) {
	TLSConf := &tls.Config{
		InsecureSkipVerify: cfg.InsecureSkipVerify,
	}

	if cfg.CertFile != "" && cfg.KeyFile != "" {
		cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
		if err != nil {
			return nil, err
		}

		TLSConf.Certificates = []tls.Certificate{cert}
	}

	if cfg.CAFile != "" {
		caPem, err := os.ReadFile(cfg.CAFile)
		if err != nil {
			return nil, err
		}

		certPool := x509.NewCertPool()
		if !certPool.AppendCertsFromPEM(caPem) {
			return nil, temperr.AppendCertsFromPEM
		}

		TLSConf.RootCAs = certPool
	}

	minVersion, maxVersion, err := HandleTLSVersion(cfg)
	if err != nil {
		return nil, err
	}

	TLSConf.MinVersion = uint16(minVersion)
	TLSConf.MaxVersion = uint16(maxVersion)

	return TLSConf, nil
}

func HandleTLSVersion(cfg *model.TLS) (minVersion, maxVersion int, err error) {
	validVersions := map[string]int{
		"1.0": tls.VersionTLS10,
		"1.1": tls.VersionTLS11,
		"1.2": tls.VersionTLS12,
		"1.3": tls.VersionTLS13,
	}

	if cfg.MaxVersion == "" {
		cfg.MaxVersion = "1.3"
	}

	if _, ok := validVersions[cfg.MaxVersion]; ok {
		maxVersion = validVersions[cfg.MaxVersion]
	} else {
		err = temperr.InvalidTLSMaxVersion
		return
	}

	if cfg.MinVersion == "" {
		cfg.MinVersion = "1.2"
	}

	if _, ok := validVersions[cfg.MinVersion]; ok {
		minVersion = validVersions[cfg.MinVersion]
	} else {
		err = temperr.InvalidTLSMinVersion
		return
	}

	if minVersion > maxVersion {
		err = temperr.InvalidTLSVersion

		return
	}

	return
}
