package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/TykTechnologies/tyk-identity-broker/providers"
	"github.com/TykTechnologies/tyk-identity-broker/tap"
	"github.com/TykTechnologies/tyk-identity-broker/tap/identity-handlers"
	"github.com/gorilla/mux"
)

// APIErrorMessage is an object that defines when a generic error occurred
type APIErrorMessage struct {
	Status string
	Error  string
}

// HandlerLogTag is a tag we are uing to identify log messages from the handler
var HandlerLogTag = "AUTH HANDLERS"

// Returns a profile ID
func getId(req *http.Request) (string, error) {
	id := mux.Vars(req)["id"]
	if id == "" {
		id = mux.Vars(req)[":id"]
	}
	if id == "" {
		return id, errors.New("No profile id detected")
	}
	return id, nil
}

// Maps an identity handler from an Action type, register new Identity Handlers and methods here
func getIdentityHandler(name tap.Action) tap.IdentityHandler {
	var thisIdentityHandler tap.IdentityHandler

	switch name {
	case tap.GenerateOrLoginDeveloperProfile, tap.GenerateOrLoginUserProfile, tap.GenerateOAuthTokenForClient, tap.GenerateTemporaryAuthToken:
		thisIdentityHandler = &identityHandlers.TykIdentityHandler{
			API:   &TykAPIHandler,
			Store: IdentityKeyStore}
	}

	return thisIdentityHandler
}

// A hack to marshal a provider conf from map[string]interface{} intoa type without type checking, ugly, but effective
func hackProviderConf(conf interface{}) []byte {
	thisConf, err := json.Marshal(conf)
	if err != nil {
		log.WithFields(logrus.Fields{
			"prefix": HandlerLogTag,
			"error":  err,
		}).Warning("Failure in JSON conversion")
		return []byte{}
	}
	return thisConf
}

// return a provider based on the name of the provider type, add new providers here
func getTAProvider(conf tap.Profile) (tap.TAProvider, error) {

	var thisProvider tap.TAProvider

	switch conf.ProviderName {
	case "SocialProvider":
		thisProvider = &providers.Social{}
	case "ADProvider":
		thisProvider = &providers.ADProvider{}
	case "ProxyProvider":
		thisProvider = &providers.ProxyProvider{}
	}

	thisIdentityHandler := getIdentityHandler(conf.ActionType)
	fmt.Printf("%+v", thisIdentityHandler)
	thisIdentityHandler.Init(conf)
	err := thisProvider.Init(thisIdentityHandler, conf, hackProviderConf(conf.ProviderConfig))

	return thisProvider, err
}

// HandleError is a generic error handler
func HandleError(tag string, errorMsg string, rawErr error, code int, w http.ResponseWriter, r *http.Request) {
	log.WithFields(logrus.Fields{
		"prefix":   tag,
		"errorMsg": errorMsg,
	}).Error(rawErr)

	errorObj := APIErrorMessage{"error", errorMsg}
	responseMsg, err := json.Marshal(&errorObj)

	if err != nil {
		log.WithField("prefix", tag).Error("[Error Handler] Couldn't marshal error stats: ", err)
		fmt.Fprintf(w, "System Error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	fmt.Fprintf(w, string(responseMsg))
}

func getTapProfile(w http.ResponseWriter, r *http.Request) (tap.TAProvider, error) {
	thisId, idErr := getId(r)
	if idErr != nil {
		HandleError(HandlerLogTag, "Could not retrieve ID", idErr, 400, w, r)
		return nil, idErr
	}

	thisProfile := tap.Profile{}
	log.WithField("prefix", HandlerLogTag).Debug("--> Looking up profile ID: ", thisId)
	foundProfileErr := AuthConfigStore.GetKey(thisId, &thisProfile)

	if foundProfileErr != nil {
		errorMsg := "Profile " + thisId + " not found"
		HandleError(HandlerLogTag, errorMsg, foundProfileErr, 404, w, r)
		return nil, foundProfileErr
	}

	thisIdentityProvider, providerErr := getTAProvider(thisProfile)
	if providerErr != nil {
		HandleError(HandlerLogTag, "Could not initialise provider", providerErr, 400, w, r)
		return nil, providerErr
	}
	return thisIdentityProvider, nil
}

// HandleAuth is the main entrypoint handler for any profile (i.e. /auth/:profile-id/:provider)
func HandleAuth(w http.ResponseWriter, r *http.Request) {
	thisIdentityProvider, err := getTapProfile(w, r)
	if err != nil {
		return
	}

	thisIdentityProvider.Handle(w, r)
	return
}

// HandleAuthCallback Is a callback URL passed to OAuth providers such as Social, handles completing an auth request
func HandleAuthCallback(w http.ResponseWriter, r *http.Request) {

	thisIdentityProvider, err := getTapProfile(w, r)
	if err != nil {
		return
	}

	thisIdentityProvider.HandleCallback(w, r, HandleError)
	return
}

func HandleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
