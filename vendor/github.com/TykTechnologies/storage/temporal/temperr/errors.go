package temperr

import "errors"

var (
	// Connection related errors
	InvalidConnector     = errors.New("invalid connector")
	InvalidOptionsType   = errors.New("invalid configuration options type")
	InvalidHandlerType   = errors.New("invalid handler type")
	InvalidConfiguration = errors.New("invalid configuration")
	ClosedConnection     = errors.New("connection closed")

	// Key related errors
	KeyNotFound = errors.New("key not found")
	KeyEmpty    = errors.New("key cannot be empty")
	KeyMisstype = errors.New("invalid operation for key type")

	// Redis related errors
	InvalidRedisClient = errors.New("invalid redis client")

	// TLS related errors
	// TLS related errors
	InvalidTLSMaxVersion = errors.New(
		"invalid MaxVersion specified. Please specify a valid TLS version: " +
			"1.0, 1.1, 1.2, or 1.3",
	)
	InvalidTLSMinVersion = errors.New(
		"invalid MinVersion specified. Please specify a valid TLS version: " +
			"1.0, 1.1, 1.2, or 1.3",
	)
	InvalidTLSVersion = errors.New(
		"MinVersion is higher than MaxVersion. Please specify a valid " +
			"MinVersion that is lower or equal to MaxVersion",
	)
	AppendCertsFromPEM = errors.New("failed to add CA certificate")

	// Others
	UnknownMessageType = errors.New("unknown message type")
)
