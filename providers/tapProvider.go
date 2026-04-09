package providers

import (
	"encoding/json"
	"errors"

	"github.com/TykTechnologies/tyk-identity-broker/constants"
	"github.com/TykTechnologies/tyk-identity-broker/tap"
	identityHandlers "github.com/TykTechnologies/tyk-identity-broker/tap/identity-handlers"
	"github.com/TykTechnologies/tyk-identity-broker/tyk-api"
	"github.com/sirupsen/logrus"
)

// return a provider based on the name of the provider type, add new providers here
func GetTAProvider(conf tap.Profile, handler tyk.TykAPI, identityKeyStore tap.AuthRegisterBackend) (tap.TAProvider, error) {

	var thisProvider tap.TAProvider
	switch conf.ProviderName {
	case constants.SocialProvider:
		thisProvider = &Social{}
	case constants.ADProvider:
		thisProvider = &ADProvider{}
	case constants.ProxyProvider:
		thisProvider = &ProxyProvider{}
	case constants.SAMLProvider:
		thisProvider = &SAMLProvider{}
	default:
		return nil, errors.New("invalid provider name")
	}

	thisIdentityHandler := getIdentityHandler(conf.ActionType, handler, identityKeyStore)
	log.Debugf("Initializing Identity Handler with config: %+v", conf)
	thisIdentityHandler.Init(conf)
	log.Debug("Initializing Provider")
	err := thisProvider.Init(thisIdentityHandler, conf, hackProviderConf(conf.ProviderConfig))

	return thisProvider, err
}

// Maps an identity handler from an Action type, register new Identity Handlers and methods here
func getIdentityHandler(name tap.Action, handler tyk.TykAPI, identityKeyStore tap.AuthRegisterBackend) tap.IdentityHandler {
	var thisIdentityHandler tap.IdentityHandler

	switch name {
	case tap.GenerateOrLoginDeveloperProfile, tap.GenerateOrLoginUserProfile, tap.GenerateOAuthTokenForClient, tap.GenerateTemporaryAuthToken:
		thisIdentityHandler = &identityHandlers.TykIdentityHandler{
			API:   &handler,
			Store: identityKeyStore}
	}

	return thisIdentityHandler
}

func GetTapProfile(AuthConfigStore, identityKeyStore tap.AuthRegisterBackend, id string, tykHandler tyk.TykAPI) (tap.TAProvider, tap.Profile, *tap.HttpError) {

	thisProfile := tap.Profile{}
	log.WithField("prefix", constants.HandlerLogTag).Debug("--> Looking up profile ID: ", id)
	foundProfileErr := AuthConfigStore.GetKey(id, thisProfile.OrgID, &thisProfile)

	if foundProfileErr != nil {
		errorMsg := "Profile " + id + " not found"
		return nil, thisProfile, &tap.HttpError{
			Message: errorMsg,
			Code:    404,
			Error:   foundProfileErr,
		}
	}

	thisIdentityProvider, providerErr := GetTAProvider(thisProfile, tykHandler, identityKeyStore)
	if providerErr != nil {
		log.WithError(providerErr).Error("Getting Tap Provider")
		return nil, thisProfile, &tap.HttpError{
			Message: "Could not initialise provider",
			Code:    400,
			Error:   providerErr,
		}
	}

	return thisIdentityProvider, thisProfile, nil
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
