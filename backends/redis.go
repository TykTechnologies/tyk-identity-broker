/* package backends provides different storage back ends for the configuration of a
TAP node. Backends ned only be k/v stores. The in-memory provider is given as an example and usefule for testing
*/
package backends

import (
	"encoding/json"
	"github.com/garyburd/redigo/redis"
	"github.com/lonelycode/redigocluster/rediscluster"
	"time"
)

var redisClusterSingleton *rediscluster.RedisCluster

// RedisBackend implements tap.AuthRegisterBackend to store profile configs in memory
type RedisBackend struct {
	db        *rediscluster.RedisCluster
	config    *RedisConfig
	KeyPrefix string
}

type RedisConfig struct {
	MaxIdle       int
	MaxActive     int
	Database      int
	Password      string
	EnableCluster bool
	Hosts         map[string]string
}

func newRedisClusterPool(forceReconnect bool, config *RedisConfig) *rediscluster.RedisCluster {
	if !forceReconnect {
		if redisClusterSingleton != nil {
			log.Debug("[REDIS STORE] Redis pool already INITIALISED")
			return redisClusterSingleton
		}
	} else {
		if redisClusterSingleton != nil {
			redisClusterSingleton.CloseConnection()
		}
	}

	log.Debug("[REDIS STORE] Creating new Redis connection pool")

	maxIdle := 100
	if config.MaxIdle > 0 {
		maxIdle = config.MaxIdle
	}

	maxActive := 500
	if config.MaxActive > 0 {
		maxActive = config.MaxActive
	}

	if config.EnableCluster {
		log.Info("[REDIS STORE] --> Using clustered mode")
	}

	thisPoolConf := rediscluster.PoolConfig{
		MaxIdle:     maxIdle,
		MaxActive:   maxActive,
		IdleTimeout: 240 * time.Second,
		Database:    config.Database,
		Password:    config.Password,
		IsCluster:   config.EnableCluster,
	}

	seed_redii := []map[string]string{}

	if len(config.Hosts) > 0 {
		for h, p := range config.Hosts {
			seed_redii = append(seed_redii, map[string]string{h: p})
		}
	} else {
		log.Fatal("[REDIS STORE] No Redis hosts set!")
	}

	thisInstance := rediscluster.NewRedisCluster(seed_redii, thisPoolConf, false)

	redisClusterSingleton = &thisInstance

	return &thisInstance
}

func (r *RedisBackend) fixKey(keyName string) string {
	return r.KeyPrefix + keyName
}

func (r *RedisBackend) connect() {
	if r.db == nil {
		log.Debug("[REDIS STORE] Connecting to redis")
		r.db = newRedisClusterPool(false, r.config)
	}

	log.Debug("[REDIS STORE] Storage Engine already initialised...")
	log.Debug("[REDIS STORE] Redis handles: ", len(r.db.Handles))

	// Reset it just in case
	r.db = redisClusterSingleton
}

// Init will create the initial in-memory store structures
func (r *RedisBackend) Init(config interface{}) {
	asJ, _ := json.Marshal(config)
	fixedConf := RedisConfig{}
	json.Unmarshal(asJ, &fixedConf)
	r.config = &fixedConf
	r.connect()
	log.Info("[REDIS STORE] Initialised")
}

// SetKey will set the value of a key in the map
func (r *RedisBackend) SetKey(key string, val interface{}) error {
	log.Debug("[REDIS STORE] SET Raw key is: ", key)
	log.Debug("[REDIS STORE] Setting key: ", r.fixKey(key))

	if r.db == nil {
		log.Info("[REDIS STORE] Connection dropped, connecting..")
		r.connect()
		return r.SetKey(key, val)
	} else {
		asByte, encErr := json.Marshal(val)
		if encErr != nil {
			return encErr
		}

		_, err := r.db.Do("SET", r.fixKey(key), string(asByte))
		if err != nil {
			log.Error("[REDIS STORE] Error trying to set value: ", err)
			return err
		}
	}

	return nil
}

// SetKey will set the value of a key in the map
func (r *RedisBackend) DeleteKey(key string) error {
	if r.db == nil {
		log.Info("[REDIS STORE] Connection dropped, connecting..")
		r.connect()
		return r.DeleteKey(key)
	}

	log.Debug("[REDIS STORE] DEL Key was: ", key)
	log.Debug("[REDIS STORE] DEL Key became: ", r.fixKey(key))
	_, err := r.db.Do("DEL", r.fixKey(key))
	if err != nil {
		log.Error("[REDIS STORE] Error trying to delete key:", err)
		return err
	}

	return nil
}

// GetKey will retuyrn the value of a key as an interface
func (r *RedisBackend) GetKey(key string, target interface{}) error {
	if r.db == nil {
		log.Info("[REDIS STORE] Connection dropped, connecting..")
		r.connect()
		return r.GetKey(key, target)
	}
	log.Debug("[REDIS STORE] Getting WAS: ", key)
	log.Debug("[REDIS STORE] Getting: ", r.fixKey(key))
	val, err := redis.String(r.db.Do("GET", r.fixKey(key)))

	decErr := json.Unmarshal([]byte(val), target)
	if decErr != nil {
		return decErr
	}

	if err != nil {
		log.Debug("[REDIS STORE] Error trying to get value:", err)
		return err
	}

	return nil
}

func (r *RedisBackend) GetAll() []interface{} {
	target := make([]interface{}, 0)
	log.Warning("[REDIS STORE] GetAll(): Not implemented")
	return target
}
