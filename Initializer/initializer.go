package Initializer

import (
	"github.com/TykTechnologies/tyk-identity-broker/backends"
	logger "github.com/TykTechnologies/tyk-identity-broker/log"
	"github.com/TykTechnologies/tyk-identity-broker/tap"
	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"

)

var log = logger.Get()
var initializerLogger = log.WithField("prefix", "INITIALIZER")

// initBackend: Get our backend to use from configs files, new backends must be registered here
func InitBackend(profileBackendConfiguration interface{}, identityBackendConfiguration interface{})(tap.AuthRegisterBackend,tap.AuthRegisterBackend) {

	AuthConfigStore := &backends.InMemoryBackend{}
	IdentityKeyStore := &backends.RedisBackend{KeyPrefix: "identity-cache."}

	initializerLogger.Info("Initialising Profile Configuration Store")
	AuthConfigStore.Init(profileBackendConfiguration)
	initializerLogger.Info("Initialising Identity Cache")
	IdentityKeyStore.Init(identityBackendConfiguration)

	return AuthConfigStore, IdentityKeyStore
}

// CreateBackendFromRedisConn: creates a redis backend from an existent redis Connection
func CreateBackendFromRedisConn(db redis.UniversalClient, keyPrefix string) tap.AuthRegisterBackend {
	redisBackend := &backends.RedisBackend{KeyPrefix: keyPrefix}

	initializerLogger.Info("Initializing Identity Cache")
	redisBackend.SetDb(db)

	return redisBackend
}

func SetLogger(newLogger *logrus.Logger){
	logger.SetLogger(newLogger)
	log = newLogger
	initializerLogger = &logrus.Entry{Logger:log}
}

func CreateInMemoryBackend() tap.AuthRegisterBackend  {
	inMemoryBackend := &backends.InMemoryBackend{}
	var config interface{}
	inMemoryBackend.Init(config)
	return inMemoryBackend
}