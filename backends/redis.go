package backends

import (
	"crypto/tls"
	"encoding/json"
	"github.com/TykTechnologies/tyk-identity-broker/log"
	"strings"
	"sync/atomic"

	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
	"strconv"
	"time"
)

var redisLoggerTag = "TIB REDIS STORE"
var redisLogger = logger.WithField("prefix", redisLoggerTag)

var singlePool atomic.Value
var singleCachePool atomic.Value
var redisUp atomic.Value

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
	config    *RedisConfig
	HashKeys  bool
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

// SetDb from existent connection
func (r *RedisBackend) SetDb(db redis.UniversalClient) {
	logger = log.Get()
	redisLogger = &logrus.Entry{Logger: logger}
	redisLogger = redisLogger.Logger.WithField("prefix", "TIB REDIS STORE")

	r.db = db
	redisLogger.Info("Set DB")
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

func (r *RedisBackend) hashKey(in string) string {
	// missing implementation
	return in
}

// GetKeys will return all keys according to the filter (filter is a prefix - e.g. tyk.keys.*)
func (r *RedisBackend) GetKeys(filter string) []string {
	db := r.db
	client := db

	searchStr := r.KeyPrefix + "*"
	logger.Debug("[STORE] Getting list by: ", searchStr)

	fnFetchKeys := func(client *redis.Client) ([]string, error) {
		values := make([]string, 0)

		iter := client.Scan(0, searchStr, 0).Iterator()
		for iter.Next() {
			values = append(values, iter.Val())
		}

		if err := iter.Err(); err != nil {
			return nil, err
		}

		return values, nil
	}

	var err error
	sessions := make([]string, 0)

	switch v := client.(type) {
	case *redis.ClusterClient:
		ch := make(chan []string)

		go func() {
			err = v.ForEachMaster(func(client *redis.Client) error {
				values, err := fnFetchKeys(client)
				if err != nil {
					return err
				}

				ch <- values
				return nil
			})
			close(ch)
		}()

		for res := range ch {
			sessions = append(sessions, res...)
		}
	case *redis.Client:
		sessions, err = fnFetchKeys(v)
	}

	if err != nil {
		logger.Error("Error while fetching keys:", err)
		return nil
	}

	for i, v := range sessions {
		sessions[i] = r.cleanKey(v)
	}

	return sessions
}

func (r *RedisBackend) GetAll() []interface{} {
	db := r.ensureConnection()
	keys := r.GetKeys(r.KeyPrefix)
	if keys == nil {
		logger.Error("Error trying to get filtered client keys")
		return nil
	}

	if len(keys) == 0 {
		return nil
	}

	for i, v := range keys {
		keys[i] = r.KeyPrefix + v
	}

	client := db
	values := make([]interface{}, 0)

	switch v := client.(type) {
	case *redis.ClusterClient:
		{
			getCmds := make([]*redis.StringCmd, 0)
			pipe := v.Pipeline()
			for _, key := range keys {
				getCmds = append(getCmds, pipe.Get(key))
			}
			_, err := pipe.Exec()
			if err != nil && err != redis.Nil {
				logger.Error("Error trying to get client keys: ", err)
				return nil
			}

			for _, cmd := range getCmds {
				values = append(values, cmd.Val())
			}
		}
	case *redis.Client:
		{
			result, err := v.MGet(keys...).Result()
			if err != nil {
				logger.Error("Error trying to get client keys: ", err)
				return nil
			}

			for _, val := range result {
				values = append(values, val)
			}
		}
	}

	return values
}

func (r *RedisBackend) cleanKey(keyName string) string {
	return strings.Replace(keyName, r.KeyPrefix, "", 1)
}

func singleton(cache bool) redis.UniversalClient {
	if cache {
		v := singleCachePool.Load()
		if v != nil {
			return v.(redis.UniversalClient)
		}
		return nil
	}
	v := singlePool.Load()
	if v != nil {
		return v.(redis.UniversalClient)
	}
	return nil
}

func (r *RedisBackend) DeleteKey(key string) error {
	db := r.ensureConnection()
	logger.Info("Trying to delete:", r.fixKey(key))
	return db.Del(r.fixKey(key)).Err()
}

func (r *RedisBackend) getDB() redis.UniversalClient {
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
