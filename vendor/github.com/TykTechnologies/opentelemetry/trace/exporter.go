package trace

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"os"

	"net"
	"net/url"
	"strings"
	"time"

	"github.com/TykTechnologies/opentelemetry/config"
	"google.golang.org/grpc/credentials"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func exporterFactory(ctx context.Context, cfg *config.OpenTelemetry) (sdktrace.SpanExporter, error) {
	var client otlptrace.Client
	var err error
	switch cfg.Exporter {
	case config.GRPCEXPORTER:
		client, err = newGRPCClient(ctx, cfg)
	case config.HTTPEXPORTER:
		client, err = newHTTPClient(ctx, cfg)
	default:
		err = fmt.Errorf("invalid exporter type: %s", cfg.Exporter)
	}

	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, time.Duration(cfg.ConnectionTimeout)*time.Second)
	defer cancel()
	// Create the trace exporter
	return otlptrace.New(ctx, client)
}

func newGRPCClient(ctx context.Context, cfg *config.OpenTelemetry) (otlptrace.Client, error) {
	clientOptions := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(cfg.Endpoint),
		otlptracegrpc.WithTimeout(time.Duration(cfg.ConnectionTimeout) * time.Second),
		otlptracegrpc.WithHeaders(cfg.Headers),
	}

	isTLSDisabled := !cfg.TLS.Enable

	if isTLSDisabled {
		clientOptions = append(clientOptions, otlptracegrpc.WithInsecure())
	} else {
		TLSConf, err := handleTLS(&cfg.TLS)
		if err != nil {
			return nil, err
		}
		clientOptions = append(clientOptions, otlptracegrpc.WithTLSCredentials(credentials.NewTLS(TLSConf)))
	}

	return otlptracegrpc.NewClient(clientOptions...), nil
}

func newHTTPClient(ctx context.Context, cfg *config.OpenTelemetry) (otlptrace.Client, error) {
	// OTel SDK does not support URL with scheme nor path, so we need to parse it
	// The scheme will be added automatically, depending on the TLSInsure setting
	endpoint := parseEndpoint(cfg)

	var clientOptions []otlptracehttp.Option
	clientOptions = append(clientOptions, otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithTimeout(time.Duration(cfg.ConnectionTimeout)*time.Second),
		otlptracehttp.WithHeaders(cfg.Headers))

	isTLSDisabled := !cfg.TLS.Enable

	if isTLSDisabled {
		clientOptions = append(clientOptions, otlptracehttp.WithInsecure())
	} else {
		TLSConf, err := handleTLS(&cfg.TLS)
		if err != nil {
			return nil, err
		}
		clientOptions = append(clientOptions, otlptracehttp.WithTLSClientConfig(TLSConf))
	}

	return otlptracehttp.NewClient(clientOptions...), nil
}

func parseEndpoint(cfg *config.OpenTelemetry) string {
	endpoint := cfg.Endpoint
	// Temporary adding scheme to get the host and port
	if !strings.Contains(endpoint, "://") {
		endpoint = "http://" + endpoint
	}

	u, err := url.Parse(endpoint)
	if err != nil {
		return cfg.Endpoint
	}

	host := u.Hostname()
	port := u.Port()

	if port == "" {
		return host
	}

	return net.JoinHostPort(host, port)
}

func handleTLS(cfg *config.TLS) (*tls.Config, error) {
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
			return nil, fmt.Errorf("failed to add CA certificate")
		}

		TLSConf.RootCAs = certPool
	}

	minVersion, maxVersion, err := handleTLSVersion(cfg)
	if err != nil {
		return nil, err
	}

	TLSConf.MinVersion = uint16(minVersion)
	TLSConf.MaxVersion = uint16(maxVersion)

	return TLSConf, nil
}

func handleTLSVersion(cfg *config.TLS) (minVersion, maxVersion int, err error) {
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
		err = errors.New("Invalid MaxVersion specified. Please specify a valid TLS version: 1.0, 1.1, 1.2, or 1.3")
		return
	}

	if cfg.MinVersion == "" {
		cfg.MinVersion = "1.2"
	}

	if _, ok := validVersions[cfg.MinVersion]; ok {
		minVersion = validVersions[cfg.MinVersion]
	} else {
		err = errors.New("Invalid MinVersion specified. Please specify a valid TLS version: 1.0, 1.1, 1.2, or 1.3")
		return
	}

	if minVersion > maxVersion {
		err = errors.New(
			"MinVersion is higher than MaxVersion. Please specify a valid MinVersion that is lower or equal to MaxVersion",
		)

		return
	}

	return
}
