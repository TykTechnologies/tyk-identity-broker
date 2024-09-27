package model

import (
	"context"
	"time"
)

type BaseConfig struct {
	RedisConfig *RedisOptions
	RetryConfig *RetryOptions
	OnConnect   func(context.Context) error
	TLS         *TLS
}

// RedisOptions contains options specific to Redis storage.
type RedisOptions struct {
	// Connection username
	Username string `json:"username"`
	// Connection password
	Password string `json:"password"`
	// Connection host. For example: "localhost"
	Host string `json:"host"`
	// Connection port. For example: 6379
	Port int `json:"port"`
	// Set a custom timeout for Redis network operations. Default value 5 seconds.
	Timeout int               `json:"timeout"`
	Hosts   map[string]string `json:"hosts"` // Deprecated: Addrs instead.
	// If you have multi-node setup, you should use this field instead. For example: ["host1:port1", "host2:port2"].
	Addrs []string `json:"addrs"`
	// Redis sentinel master name
	MasterName string `json:"master_name"`
	// Redis sentinel password
	SentinelPassword string `json:"sentinel_password"`
	// Redis database
	Database int `json:"database"`
	// Set the number of maximum connections in the Redis connection pool, which defaults to 500
	// Set to a higher value if you are expecting more traffic.
	MaxActive int `json:"optimisation_max_active"`
	// Enable Redis Cluster support
	EnableCluster bool `json:"enable_cluster"`
}

type RetryOptions struct {
	// Maximum number of retries before error.
	MaxRetries int
	// Minimum backoff between each retry.
	MinRetryBackoff time.Duration
	// Maximum backoff between each retry.
	MaxRetryBackoff time.Duration
}

type TLS struct {
	// Flag that can be used to enable TLS. Defaults to false (disabled).
	Enable bool `json:"enable"`
	// Flag that can be used to skip TLS verification if TLS is enabled.
	// Defaults to false.
	InsecureSkipVerify bool `json:"insecure_skip_verify"`
	// Path to the CA file.
	CAFile string `json:"ca_file"`
	// Path to the cert file.
	CertFile string `json:"cert_file"`
	// Path to the key file.
	KeyFile string `json:"key_file"`
	// Maximum TLS version that is supported.
	// Options: ["1.0", "1.1", "1.2", "1.3"].
	// Defaults to "1.3".
	MaxVersion string `json:"max_version"`
	// Minimum TLS version that is supported.
	// Options: ["1.0", "1.1", "1.2", "1.3"].
	// Defaults to "1.2".
	MinVersion string `json:"min_version"`
}
