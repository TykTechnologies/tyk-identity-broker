package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/lonelycode/tyk-auth-proxy/providers"
	"github.com/lonelycode/tyk-auth-proxy/tap"
	"github.com/lonelycode/tyk-auth-proxy/tap/identity-handlers"
	"net/http"
)

// APIErrorMessage is an object that defines when a generic error occurred
type APIErrorMessage struct {
	Status string
	Error  string
}

// HandlerLogTag is a tag we are uing to identify log messages from the handler
var HandlerLogTag = "[AUTH HANDLERS]"

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
	case tap.GenerateOrLoginDeveloperProfile:
		thisIdentityHandler = &identityHandlers.TykIdentityHandler{
			API:   &TykAPIHandler,
			Store: IdentityKeyStore}
	case tap.GenerateOrLoginUserProfile:
		thisIdentityHandler = &identityHandlers.TykIdentityHandler{
			API:   &TykAPIHandler,
			Store: IdentityKeyStore}
	case tap.GenerateOAuthTokenForClient:
		thisIdentityHandler = &identityHandlers.TykIdentityHandler{
			API:   &TykAPIHandler,
			Store: IdentityKeyStore}
	case tap.GenerateTemporaryAuthToken:
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
		log.Warning("Failure in JSON conversion")
		return []byte{}
	}
	return thisConf
}

// return a provider based on the name of the provider type, add new providers here
func getTAProvider(conf tap.Profile) tap.TAProvider {

	var thisProvider tap.TAProvider

	switch conf.ProviderName {
	case "SocialProvider":
		thisProvider = &providers.Social{}
	case "ADProvider":
		thisProvider = &providers.ADProvider{}
	case "ProxyProvider":
		thisProvider = &providers.ProxyProvider{}
	}

	var thisIdentityHandler tap.IdentityHandler = getIdentityHandler(conf.ActionType)
	thisIdentityHandler.Init(conf)
	thisProvider.Init(thisIdentityHandler, conf, hackProviderConf(conf.ProviderConfig))
	return thisProvider

}

// HandleError is a generic error handler
func HandleError(tag string, errorMsg string, rawErr error, code int, w http.ResponseWriter, r *http.Request) {
	log.Error(tag+" "+errorMsg+": ", rawErr)

	errorObj := APIErrorMessage{"error", errorMsg}
	responseMsg, err := json.Marshal(&errorObj)

	if err != nil {
		log.Error("[Error Handler] Couldn't marshal error stats: ", err)
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
	log.Debug(HandlerLogTag+" --> Looking up profile ID:", thisId)
	foundProfileErr := AuthConfigStore.GetKey(thisId, &thisProfile)

	if foundProfileErr != nil {
		HandleError(HandlerLogTag, "Profile not found", foundProfileErr, 404, w, r)
		return nil, foundProfileErr
	}

	thisIdentityProvider := getTAProvider(thisProfile)
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
