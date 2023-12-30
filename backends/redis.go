package backends

import (
	"context"
	"encoding/json"
	"github.com/TykTechnologies/storage/temporal/connector"
	temporal "github.com/TykTechnologies/storage/temporal/keyvalue"
	"github.com/TykTechnologies/storage/temporal/model"
	"strings"
	"sync/atomic"

	"github.com/TykTechnologies/tyk-identity-broker/log"

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
	kv        temporal.KeyValue
	config    *RedisConfig
	HashKeys  bool
	KeyPrefix string
}

type KeyError struct{}

func (e KeyError) Error() string {
	return "Key not found"
}

func (r *RedisBackend) Connect() (temporal.KeyValue, error) {
	redisLogger.Info("Creating new Redis connection pool")

	conf := r.config
	// ToDo: why we do not have TLS conf?
	optsR := model.RedisOptions{
		Username:         conf.Username,
		Password:         conf.Password,
		Host:             conf.Host,
		Port:             conf.Port,
		Timeout:          conf.Timeout,
		Hosts:            conf.Hosts,
		Addrs:            conf.Addrs,
		MasterName:       conf.MasterName,
		SentinelPassword: conf.SentinelPassword,
		Database:         conf.Database,
		MaxActive:        conf.MaxActive,
		EnableCluster:    conf.EnableCluster,
	}

	tls := model.TLS{
		Enable:             conf.UseSSL,
		InsecureSkipVerify: conf.SSLInsecureSkipVerify,
	}

	connector, err := connector.NewConnector(model.RedisV9Type, model.WithRedisConfig(&optsR), model.WithTLS(&tls))
	if err != nil {
		redisLogger.WithError(err).Error("creating redis connector")
		return nil, err
	}
	r.kv, err = temporal.NewKeyValue(connector)
	if err != nil {
		redisLogger.WithError(err).Error("creating KV store")
		return nil, err
	}

	return r.kv, nil
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
func (r *RedisBackend) SetDb(kv temporal.KeyValue) {
	logger = log.Get()
	redisLogger = &logrus.Entry{Logger: logger}
	redisLogger = redisLogger.Logger.WithField("prefix", "TIB REDIS STORE")

	r.kv = kv
	redisLogger.Info("Set KV store")
}

func (r *RedisBackend) SetKey(key string, orgId string, val interface{}) error {

	if err := r.kv.Set(ctx, r.fixKey(key), val.(string), 0); err != nil {
		redisLogger.WithError(err).Debug("Error trying to set value")
		return err
	}

	return nil
}

func (r *RedisBackend) GetKey(key string, orgId string, val interface{}) error {
	result, err := r.kv.Get(ctx, r.fixKey(key))
	if err != nil {
		return err
	}

	// if AuthConfigStore is redis adapter, then redis return string
	if err = json.Unmarshal([]byte(result), &val); err != nil {
		redisLogger.WithError(err).Error("unmarshalling redis result into interface")
	}

	return err
}

// GetKeys will return all keys according to the filter (filter is a prefix - e.g. tyk.keys.*)
func (r *RedisBackend) GetKeys(filter string) []string {
	keys, err := r.kv.Keys(ctx, filter)
	if err != nil {
		redisLogger.WithError(err).Error("getting keys")
	}

	for k, v := range keys {
		keys[k] = r.cleanKey(v)
	}
	return keys
}

func (r *RedisBackend) GetAll(orgId string) []interface{} {

	keys, err := r.kv.Keys(ctx, r.KeyPrefix)
	if err != nil {
		redisLogger.WithError(err).Error("retrieving keys from redis")
		return nil
	}

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

	var values []interface{} = make([]interface{}, len(keys))
	for i, s := range keys {
		values[i] = s
	}
	return values
}

func (r *RedisBackend) cleanKey(keyName string) string {
	return strings.Replace(keyName, r.KeyPrefix, "", 1)
}

func (r *RedisBackend) hashKey(in string) string {
	// missing implementation
	return in
}

func (r *RedisBackend) DeleteKey(key string, orgId string) error {
	return r.kv.Delete(ctx, r.fixKey(key))
}

func (r *RedisBackend) fixKey(keyName string) string {
	return r.KeyPrefix + keyName
}
