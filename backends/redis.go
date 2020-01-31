package backends

import (
	"crypto/tls"
	"encoding/json"
	"strconv"
	"sync"
	"time"

	"github.com/go-redis/redis"
)

var redisLogger = log.WithField("prefix", "REDIS STORE")

type RedisConfig struct {
	MaxIdle               int
	MaxActive             int
	MasterName            string
	Database              int
	Password              string
	EnableCluster         bool
	Hosts                 map[string]string
	UseSSL                bool
	SSLInsecureSkipVerify bool
	Timeout               int
	Port                  int
	Host                  string
}

type RedisBackend struct {
	db        redis.UniversalClient
	dbMu      sync.RWMutex
	config    *RedisConfig
	KeyPrefix string
}

type KeyError struct{}

func (e KeyError) Error() string {
	return "Key not found"
}

func (r *RedisBackend) newRedisClusterPool() redis.UniversalClient {
	redisLogger.Info("Creating new Redis connection pool")

	timeout := 5 * time.Second

	if r.config.Timeout > 0 {
		timeout = time.Duration(r.config.Timeout) * time.Second
	}

	var tlsConfig *tls.Config
	if r.config.UseSSL {
		tlsConfig = &tls.Config{
			InsecureSkipVerify: r.config.SSLInsecureSkipVerify,
		}
	}

	var address []string
	if len(r.config.Hosts) > 0 {
		for h, p := range r.config.Hosts {
			addr := h + ":" + p
			address = append(address, addr)
		}
	} else {
		addr := r.config.Host + ":" + strconv.Itoa(r.config.Port)
		address = append(address, addr)
	}

	if !r.config.EnableCluster {
		address = address[:1]
	}

	var client redis.UniversalClient
	opts := &RedisOpts{
		MasterName:   r.config.MasterName,
		Addrs:        address,
		DB:           r.config.Database,
		Password:     r.config.Password,
		IdleTimeout:  240 * time.Second,
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
		DialTimeout:  timeout,
		TLSConfig:    tlsConfig,
	}

	if opts.MasterName != "" {
		redisLogger.Info("Creating sentinel-backed fail-over client")
		client = redis.NewFailoverClient(opts.failover())
	} else if r.config.EnableCluster {
		redisLogger.Info("Creating cluster client")
		client = redis.NewClusterClient(opts.cluster())
	} else {
		redisLogger.Info("Creating single-node client")
		client = redis.NewClient(opts.simple())
	}

	return client
}

func (r *RedisBackend) Connect() bool {
	r.dbMu.Lock()
	defer r.dbMu.Unlock()

	r.db = r.newRedisClusterPool()
	return true
}

// Init will create the initial in-memory store structures
func (r *RedisBackend) Init(config interface{}) {
	asJ, _ := json.Marshal(config)
	fixedConf := RedisConfig{}
	json.Unmarshal(asJ, &fixedConf)
	r.config = &fixedConf
	r.Connect()
	redisLogger.Info("Initialized")
}

func (r *RedisBackend) SetKey(key string, val interface{}) error {
	db := r.ensureConnection()

	redisLogger.Debug("Setting key=", key)
	if err := db.Set(r.fixKey(key), val, 0).Err(); err != nil {
		redisLogger.WithError(err).Debug("Error trying to set value")
		return err
	}

	return nil
}

func (r *RedisBackend) GetKey(key string, val interface{}) error {
	db := r.ensureConnection()
	var err error
	val, err = db.Get(r.fixKey(key)).Result()
	if err != nil {
		return err
	}
	return nil
}

func (r *RedisBackend) GetAll() []interface{} {
	target := make([]interface{}, 0)
	redisLogger.Warning("GetAll() Not implemented")
	return target
}

func (r *RedisBackend) DeleteKey(key string) error {
	db := r.ensureConnection()
	return db.Del(r.fixKey(key)).Err()
}

func (r *RedisBackend) getDB() redis.UniversalClient {
	r.dbMu.RLock()
	defer r.dbMu.RUnlock()

	return r.db
}

func (r *RedisBackend) ensureConnection() redis.UniversalClient {
	if db := r.getDB(); db != nil {
		// already connected
		return db
	}
	redisLogger.Info("Connection dropped, reconnecting...")
	for {
		r.Connect()
		if db := r.getDB(); db != nil {
			// reconnection worked
			return db
		}
		redisLogger.Info("Reconnecting again...")
	}
}

func (r *RedisBackend) fixKey(keyName string) string {
	return r.KeyPrefix + keyName
}


// RedisOpts is the overriden type of redis.UniversalOptions. simple() and cluster() functions are not public
// in redis library. Therefore, they are redefined in here to use in creation of new redis cluster logic.
// We don't want to use redis.NewUniversalClient() logic.
type RedisOpts redis.UniversalOptions

func (o *RedisOpts) failover() *redis.FailoverOptions {
	if len(o.Addrs) == 0 {
		o.Addrs = []string{"127.0.0.1:6379"}
	}

	return &redis.FailoverOptions{
		SentinelAddrs: o.Addrs,
		MasterName:    o.MasterName,
		OnConnect:     o.OnConnect,

		DB:       o.DB,
		Password: o.Password,

		MaxRetries:      o.MaxRetries,
		MinRetryBackoff: o.MinRetryBackoff,
		MaxRetryBackoff: o.MaxRetryBackoff,

		DialTimeout:  o.DialTimeout,
		ReadTimeout:  o.ReadTimeout,
		WriteTimeout: o.WriteTimeout,

		PoolSize:           o.PoolSize,
		MinIdleConns:       o.MinIdleConns,
		MaxConnAge:         o.MaxConnAge,
		PoolTimeout:        o.PoolTimeout,
		IdleTimeout:        o.IdleTimeout,
		IdleCheckFrequency: o.IdleCheckFrequency,

		TLSConfig: o.TLSConfig,
	}
}

func (o *RedisOpts) cluster() *redis.ClusterOptions {
	if len(o.Addrs) == 0 {
		o.Addrs = []string{"127.0.0.1:6379"}
	}

	return &redis.ClusterOptions{
		Addrs:     o.Addrs,
		OnConnect: o.OnConnect,

		Password: o.Password,

		MaxRedirects:   o.MaxRedirects,
		ReadOnly:       o.ReadOnly,
		RouteByLatency: o.RouteByLatency,
		RouteRandomly:  o.RouteRandomly,

		MaxRetries:      o.MaxRetries,
		MinRetryBackoff: o.MinRetryBackoff,
		MaxRetryBackoff: o.MaxRetryBackoff,

		DialTimeout:        o.DialTimeout,
		ReadTimeout:        o.ReadTimeout,
		WriteTimeout:       o.WriteTimeout,
		PoolSize:           o.PoolSize,
		MinIdleConns:       o.MinIdleConns,
		MaxConnAge:         o.MaxConnAge,
		PoolTimeout:        o.PoolTimeout,
		IdleTimeout:        o.IdleTimeout,
		IdleCheckFrequency: o.IdleCheckFrequency,

		TLSConfig: o.TLSConfig,
	}
}

func (o *RedisOpts) simple() *redis.Options {
	addr := "127.0.0.1:6379"
	if len(o.Addrs) > 0 {
		addr = o.Addrs[0]
	}

	return &redis.Options{
		Addr:      addr,
		OnConnect: o.OnConnect,

		DB:       o.DB,
		Password: o.Password,

		MaxRetries:      o.MaxRetries,
		MinRetryBackoff: o.MinRetryBackoff,
		MaxRetryBackoff: o.MaxRetryBackoff,

		DialTimeout:  o.DialTimeout,
		ReadTimeout:  o.ReadTimeout,
		WriteTimeout: o.WriteTimeout,

		PoolSize:           o.PoolSize,
		MinIdleConns:       o.MinIdleConns,
		MaxConnAge:         o.MaxConnAge,
		PoolTimeout:        o.PoolTimeout,
		IdleTimeout:        o.IdleTimeout,
		IdleCheckFrequency: o.IdleCheckFrequency,

		TLSConfig: o.TLSConfig,
	}
}
