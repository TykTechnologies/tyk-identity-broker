package backends

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"strings"
	"sync/atomic"

	"github.com/TykTechnologies/tyk-identity-broker/log"

	"strconv"
	"time"

	"github.com/TykTechnologies/tyk-identity-broker/internal/redis"
	"github.com/sirupsen/logrus"
)

var redisLoggerTag = "TIB REDIS STORE"
var redisLogger = logger.WithField("prefix", redisLoggerTag)

var singlePool atomic.Value
var singleCachePool atomic.Value
var redisUp atomic.Value
var ctx = context.Background()

type RedisConfig struct {
	MaxIdle               int
	MaxActive             int
	MasterName            string
	SentinelPassword      string
	Database              int
	Username              string
	Password              string
	EnableCluster         bool
	Hosts                 map[string]string // Deprecated: Use Addrs instead.
	Addrs                 []string
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

	var client redis.UniversalClient
	opts := &redis.UniversalOptions{
		MasterName:       r.config.MasterName,
		SentinelPassword: r.config.SentinelPassword,
		Addrs:            r.getRedisAddrs(),
		DB:               r.config.Database,
		Username:         r.config.Username,
		Password:         r.config.Password,
		ReadTimeout:      timeout,
		WriteTimeout:     timeout,
		DialTimeout:      timeout,
		TLSConfig:        tlsConfig,
	}

	if opts.MasterName != "" {
		redisLogger.Info("Creating sentinel-backed fail-over client")
		client = redis.NewFailoverClient(opts.Failover())
	} else if r.config.EnableCluster {
		redisLogger.Info("Creating cluster client")
		client = redis.NewClusterClient(opts.Cluster())
	} else {
		redisLogger.Info("Creating single-node client")
		client = redis.NewClient(opts.Simple())
	}

	return client
}

func (r *RedisBackend) getRedisAddrs() (addrs []string) {
	if len(r.config.Addrs) != 0 {
		addrs = r.config.Addrs
	} else {
		for h, p := range r.config.Hosts {
			addr := h + ":" + p
			addrs = append(addrs, addr)
		}
	}

	if len(addrs) == 0 && r.config.Port != 0 {
		addr := r.config.Host + ":" + strconv.Itoa(r.config.Port)
		addrs = append(addrs, addr)
	}

	return addrs
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

func (r *RedisBackend) SetKey(key string, orgId string, val interface{}) error {
	db := r.ensureConnection()

	if err := db.Set(ctx, r.fixKey(key), val, 0).Err(); err != nil {
		redisLogger.WithError(err).Debug("Error trying to set value")
		return err
	}

	return nil
}

func (r *RedisBackend) GetKey(key string, orgId string, val interface{}) error {
	db := r.ensureConnection()
	var err error
	result, err := db.Get(ctx, r.fixKey(key)).Result()
	if err != nil {
		return err
	}

	// if AuthConfigStore is redis adapter, then redis return string
	if err = json.Unmarshal([]byte(result), &val); err != nil {
		redisLogger.WithError(err).Error("unmarshalling redis result into interface")
	}

	return err
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

		iter := client.Scan(ctx, 0, searchStr, 0).Iterator()
		for iter.Next(ctx) {
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
			err = v.ForEachMaster(ctx, func(context context.Context, client *redis.Client) error {
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

func (r *RedisBackend) GetAll(orgId string) []interface{} {
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
				getCmds = append(getCmds, pipe.Get(ctx, key))
			}
			_, err := pipe.Exec(ctx)
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
			result, err := v.MGet(ctx, keys...).Result()
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

func (r *RedisBackend) DeleteKey(key string, orgId string) error {
	db := r.ensureConnection()
	return db.Del(ctx, r.fixKey(key)).Err()
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
