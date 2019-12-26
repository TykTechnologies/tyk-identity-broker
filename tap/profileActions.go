package tap

import (
	"encoding/json"
	"errors"
	"github.com/Sirupsen/logrus"
	"github.com/TykTechnologies/tyk-identity-broker/constants"
	logger "github.com/TykTechnologies/tyk-identity-broker/log"
	"github.com/TykTechnologies/tyk-identity-broker/providers"
	identityHandlers "github.com/TykTechnologies/tyk-identity-broker/tap/identity-handlers"
	"github.com/TykTechnologies/tyk-identity-broker/tyk-api"
	"net/http"
)

var log = logger.Get()

type HttpError struct{
	Message string
	Code int
	Error error
}

func AddProfile(profile Profile, AuthConfigStore AuthRegisterBackend, flush func(backend AuthRegisterBackend) error) *HttpError {
	dumpProfile := Profile{}
	keyErr := AuthConfigStore.GetKey(profile.ID, &dumpProfile)
	if keyErr != nil {
		return &HttpError{
			Message: "Object ID already exists",
			Code:    400,
			Error:   keyErr,
		}
	}

	saveErr := AuthConfigStore.SetKey(profile.ID, &profile)
	if saveErr != nil {
		return &HttpError{
			Message: "Update failed",
			Code:    500,
			Error:   saveErr,
		}
	}

	fErr := flush(AuthConfigStore)
	if fErr != nil {
		return &HttpError{
			Message: "flush failed",
			Code:    400,
			Error:   fErr,
		}
	}

	return nil
}

func UpdateProfile(key string, profile Profile, AuthConfigStore AuthRegisterBackend,flush func(backend AuthRegisterBackend) error) *HttpError {

	// Shenanigans
	if profile.ID != key {
		return &HttpError{
			Message: "Object ID and URI resource ID do not match",
			Code:    400,
			Error:   errors.New("ID Mismatch"),
		}
	}

	dumpProfile := Profile{}
	keyErr := AuthConfigStore.GetKey(key, &dumpProfile)
	if keyErr != nil {
		return &HttpError{
			Message: "Object ID does not exist, operation not permitted",
			Code:    400,
			Error:   keyErr,
		}
	}

	saveErr := AuthConfigStore.SetKey(key, &profile)
	if saveErr != nil {
		return &HttpError{
			Message: "Update failed",
			Code:    500,
			Error:   saveErr,
		}
	}

	return nil
}

func DeleteProfile(key string,AuthConfigStore AuthRegisterBackend, flush func(backend AuthRegisterBackend) error) *HttpError {

	dumpProfile := Profile{}
	keyErr := AuthConfigStore.GetKey(key, &dumpProfile)
	if keyErr != nil {
		return &HttpError{
			Message: "Object ID does not exist",
			Code:    400,
			Error:  keyErr,
		}
	}

	delErr := AuthConfigStore.DeleteKey(key)
	if delErr != nil {
		return &HttpError{
			Message: "Delete failed",
			Code:    500,
			Error:   delErr,
		}
	}

	fErr := flush(AuthConfigStore)
	if fErr != nil {
		return &HttpError{
			Message: "flush failed",
			Code:    400,
			Error:   fErr,
		}
	}
	return nil
}

func GetTapProfile(w http.ResponseWriter, r *http.Request, AuthConfigStore, identityKeyStore AuthRegisterBackend, id string,tykHandler tyk.TykAPI) (TAProvider, *HttpError) {

	thisProfile := Profile{}
	log.WithField("prefix", constants.HandlerLogTag).Debug("--> Looking up profile ID: ", id)
	foundProfileErr := AuthConfigStore.GetKey(id, &thisProfile)

	if foundProfileErr != nil {
		errorMsg := "Profile " + id + " not found"
		return nil, &HttpError{
			Message: errorMsg,
			Code:    404,
			Error:   foundProfileErr,
		}
	}

	thisIdentityProvider, providerErr := GetTAProvider(thisProfile,tykHandler,identityKeyStore)
	if providerErr != nil {
		return  nil, &HttpError{
			Message: "Could not initialise provider",
			Code:    400,
			Error:   providerErr,
		}
	}

	return thisIdentityProvider, nil
}

// return a provider based on the name of the provider type, add new providers here
func GetTAProvider(conf Profile,handler tyk.TykAPI, identityKeyStore AuthRegisterBackend) (TAProvider, error) {

	var thisProvider TAProvider

	switch conf.ProviderName {
	case constants.SocialProvider:
		thisProvider = &providers.Social{}
	case constants.ADProvider:
		thisProvider = &providers.ADProvider{}
	case constants.ProxyProvider:
		thisProvider = &providers.ProxyProvider{}
	default:
		return nil, errors.New("invalid provider name")
	}

	thisIdentityHandler := getIdentityHandler(conf.ActionType, handler, identityKeyStore)
	thisIdentityHandler.Init(conf)
	err := thisProvider.Init(thisIdentityHandler, conf, hackProviderConf(conf.ProviderConfig))

	return thisProvider, err
}

// Maps an identity handler from an Action type, register new Identity Handlers and methods here
func getIdentityHandler(name Action,handler tyk.TykAPI, identityKeyStore AuthRegisterBackend) IdentityHandler {
	var thisIdentityHandler IdentityHandler

	switch name {
	case GenerateOrLoginDeveloperProfile, GenerateOrLoginUserProfile, GenerateOAuthTokenForClient, GenerateTemporaryAuthToken:
		thisIdentityHandler = &identityHandlers.TykIdentityHandler{
			API:   &handler,
			Store: identityKeyStore}
	}

	return thisIdentityHandler
}

// A hack to marshal a provider conf from map[string]interface{} into a type without type checking, ugly, but effective
func hackProviderConf(conf interface{}) []byte {
	thisConf, err := json.Marshal(conf)
	if err != nil {
		log.WithFields(logrus.Fields{
			"prefix": constants.HandlerLogTag,
			"error":  err,
		}).Warning("Failure in JSON conversion")
		return []byte{}
	}
	return thisConf
}
