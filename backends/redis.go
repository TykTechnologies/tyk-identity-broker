package backends

import (
	"crypto/tls"
	"encoding/json"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis"
)

var redisLogger = log.WithField("prefix", "REDIS STORE")

type RedisConfig struct {
	MaxIdle               int
	MaxActive             int
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

	seedRedis := []string{}
	if len(r.config.Hosts) > 0 {
		for h, p := range r.config.Hosts {
			addr := h + ":" + p
			seedRedis = append(seedRedis, addr)
		}
	} else {
		addr := r.config.Host + ":" + strconv.Itoa(r.config.Port)
		seedRedis = append(seedRedis, addr)
	}

	if !r.config.EnableCluster {
		seedRedis = seedRedis[:1]
	}

	thisInstance := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:        seedRedis,
		DB:           r.config.Database,
		Password:     r.config.Password,
		IdleTimeout:  240 * time.Second,
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
		DialTimeout:  timeout,
		TLSConfig:    tlsConfig,
	})

	return thisInstance
}

func (r *RedisBackend) Connect() bool {
	r.dbMu.Lock()
	defer r.dbMu.Unlock()

	r.db = r.newRedisClusterPool()
	return true
}

func (r *RedisBackend) cleanKey(keyName string) string {
	return strings.Replace(keyName, r.KeyPrefix, "", 1)
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
	r.db = db
	redisLogger.Info("Set DB")
}

// SetDbMu set mutex from existent connection
func (r *RedisBackend) SetDbMu(mu sync.RWMutex) {
	r.dbMu = mu
	redisLogger.Info("Set db mutex")
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
