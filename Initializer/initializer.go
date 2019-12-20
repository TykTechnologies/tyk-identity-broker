package Initializer

import (
	"github.com/TykTechnologies/tyk-identity-broker/backends"
	logger "github.com/TykTechnologies/tyk-identity-broker/log"
	"github.com/TykTechnologies/tyk-identity-broker/tap"
)

var log = logger.Get()
var initializerLogger = log.WithField("prefix", "INITIALIZER")

// initBackend: Get our backend to use, new backends must be registered here
func InitBackend(profileBackendConfiguration interface{}, identityBackendConfiguration interface{})(tap.AuthRegisterBackend,tap.AuthRegisterBackend) {

	AuthConfigStore := &backends.InMemoryBackend{}
	IdentityKeyStore := &backends.RedisBackend{KeyPrefix: "identity-cache."}

	initializerLogger.Info("Initialising Profile Configuration Store")
	AuthConfigStore.Init(profileBackendConfiguration)
	initializerLogger.Info("Initialising Identity Cache")
	IdentityKeyStore.Init(identityBackendConfiguration)

	return AuthConfigStore, IdentityKeyStore
}
