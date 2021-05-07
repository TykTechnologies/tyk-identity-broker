package initializer

import (
	"github.com/TykTechnologies/tyk-identity-broker/backends"

	logger "github.com/TykTechnologies/tyk-identity-broker/log"
	"github.com/TykTechnologies/tyk-identity-broker/providers"
	"github.com/TykTechnologies/tyk-identity-broker/tap"
	"github.com/TykTechnologies/tyk-identity-broker/tothic"
	"github.com/TykTechnologies/tyk/certs"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	mgo "gopkg.in/mgo.v2"
)

var log = logger.Get()
var initializerLogger = log.WithField("prefix", "TIB INITIALIZER")

// initBackend: Get our backend to use from configs files, new back-ends must be registered here
func InitBackend(profileBackendConfiguration interface{}, identityBackendConfiguration interface{}) (tap.AuthRegisterBackend, tap.AuthRegisterBackend) {

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

func SetLogger(newLogger *logrus.Logger) {
	logger.SetLogger(newLogger)
	log = newLogger

	initializerLogger = &logrus.Entry{Logger: log}
	initializerLogger = initializerLogger.Logger.WithField("prefix", "TIB INITIALIZER")
}

func SetCertManager(cm *certs.CertificateManager) {
	providers.CertManager = cm
}

func CreateInMemoryBackend() tap.AuthRegisterBackend {
	inMemoryBackend := &backends.InMemoryBackend{}
	var config interface{}
	inMemoryBackend.Init(config)
	return inMemoryBackend
}

func CreateMongoBackend(db *mgo.Database) tap.AuthRegisterBackend {
	mongoBackend := &backends.MongoBackend{
		Db:         db,
		Collection: tap.ProfilesCollectionName,
	}
	var config interface{}
	mongoBackend.Init(config)
	return mongoBackend
}

func SetConfigHandler(backend tap.AuthRegisterBackend) {
	tothic.SetParamsStoreHandler(backend)
}
