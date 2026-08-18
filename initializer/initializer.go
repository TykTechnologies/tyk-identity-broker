package initializer

import (
	"errors"

	temporal "github.com/TykTechnologies/storage/temporal/keyvalue"
	"github.com/TykTechnologies/tyk-identity-broker/backends"
	tykerror "github.com/TykTechnologies/tyk-identity-broker/error"

	logger "github.com/TykTechnologies/tyk-identity-broker/log"
	"github.com/TykTechnologies/tyk-identity-broker/providers"
	"github.com/TykTechnologies/tyk-identity-broker/tap"
	"github.com/TykTechnologies/tyk-identity-broker/tothic"
	"github.com/TykTechnologies/tyk/certs"
	"github.com/sirupsen/logrus"
)

var log = logger.Get()
var initializerLogger = log.WithField("prefix", "TIB INITIALIZER")

// IdentityKeyStore keeps a record of identities tied to tokens (if needed)
var IdentityKeyStore tap.AuthRegisterBackend

// AuthConfigStore Is the back end we are storing our configuration files to
var AuthConfigStore tap.AuthRegisterBackend

// initBackend: Get our backend to use from configs files, new back-ends must be registered here
func InitBackend(profileBackendConfiguration interface{}, identityBackendConfiguration interface{}) (tap.AuthRegisterBackend, tap.AuthRegisterBackend) {

	AuthConfigStore = &backends.InMemoryBackend{}
	IdentityKeyStore = &backends.RedisBackend{KeyPrefix: "identity-cache."}

	initializerLogger.Info("Initialising Profile Configuration Store")
	AuthConfigStore.Init(profileBackendConfiguration)
	initializerLogger.Info("Initialising Identity Cache")
	IdentityKeyStore.Init(identityBackendConfiguration)

	return AuthConfigStore, IdentityKeyStore
}

// CreateBackendFromRedisConn: creates a redis backend from an existent redis Connection
func createBackendFromRedisConn(kv temporal.KeyValue, keyPrefix string) tap.AuthRegisterBackend {
	redisBackend := &backends.RedisBackend{KeyPrefix: keyPrefix}
	redisBackend.SetDb(kv)
	return redisBackend
}

func setLogger(newLogger *logrus.Logger) {
	logger.SetLogger(newLogger)
	log = newLogger

	initializerLogger = &logrus.Entry{Logger: log}
	initializerLogger = initializerLogger.Logger.WithField("prefix", "TIB INITIALIZER")
}

func SetCertManager(cm certs.CertificateManager) {
	providers.CertManager = cm
}

func SetConfigHandler(backend tap.AuthRegisterBackend) {
	tothic.SetParamsStoreHandler(backend)
}

func setKVStorage(kv temporal.KeyValue) {
	configHandler := createBackendFromRedisConn(kv, "tib-provider-config-")

	initializerLogger.Info("Initializing Config handler")
	tothic.SetParamsStoreHandler(configHandler)

	initializerLogger.Info("Initializing Identity Cache")
	IdentityKeyStore = createBackendFromRedisConn(kv, "identity-cache")
}

type TIB struct {
	Logger       *logrus.Logger
	KV           temporal.KeyValue
	CertManager  certs.CertificateManager
	CookieSecret string
}

func (tib *TIB) Start() error {
	if tib.Logger == nil {
		tib.Logger = logrus.New()
	}
	setLogger(tib.Logger)

	if tib.KV == nil {
		return errors.New("kv store cannot be nil")
	}
	setKVStorage(tib.KV)

	if tib.CertManager == nil {
		return errors.New("certificate manager cannot be nil")
	}
	SetCertManager(tib.CertManager)
	tothic.TothErrorHandler = tykerror.HandleError
	if tib.CookieSecret != "" {
		tothic.SetupSessionStore(tib.CookieSecret)
	} else {
		// then it will read it from env
		tothic.SetupSessionStore()
	}
	return nil
}
