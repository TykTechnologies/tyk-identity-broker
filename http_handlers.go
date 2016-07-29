package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"tyk-identity-broker/providers"
	"tyk-identity-broker/tap"
	"tyk-identity-broker/tap/identity-handlers"
	"net/http"
)

// APIErrorMessage is an object that defines when a generic error occurred
type APIErrorMessage struct {
	Status string
	Error  string
}

// HandlerLogTag is a tag we are using to identify log messages from the handler
var HandlerLogTag = "[AUTH HANDLERS] "

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

// A hack to marshal a provider conf from map[string]interface{} into a type without type checking, ugly, but effective
func hackProviderConf(conf interface{}) []byte {
	thisConf, err := json.Marshal(conf)
	if err != nil {
		log.Warning(HandlerLogTag + "Failure in JSON conversion")
		return []byte{}
	}
	return thisConf
}

// return a provider based on the name of the provider type, add new providers here
func getTAProvider(conf tap.Profile) providers.TAProvider {

	var thisProvider providers.TAProvider

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

// Generic Error Handler which logs and then outputs HTTP Error Response
func HandleError(logTag string, errorMsg string, rawErr error, code int, w http.ResponseWriter, r *http.Request) {
	log.Error(logTag, errorMsg, rawErr)
	RespondWithError(errorMsg, code, w, r)
	return
}

// Outputs an HTTP Error Response
func RespondWithError(errorMsg string, code int, w http.ResponseWriter, r *http.Request) {
	errorObj := APIErrorMessage{"error", errorMsg}
	responseMsg, err := json.Marshal(&errorObj)

	if err != nil {
		log.Error(HandlerLogTag, "Couldn't marshal error stats: ", err)
		responseMsg = []byte("{\"Status\":\"error\",\"Error\": \"System Error\"}")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	fmt.Fprintf(w, string(responseMsg))
	return
}

func getTapProfile(w http.ResponseWriter, r *http.Request) (providers.TAProvider, error) {
	thisId, idErr := getId(r)

	if idErr != nil {
		HandleError(HandlerLogTag, "Could not retrieve ID", idErr, 400, w, r)
		return nil, idErr
	}

	thisProfile := tap.Profile{}
	log.Debug(HandlerLogTag+"--> Looking up profile ID:", thisId)
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
		HandleError(HandlerLogTag, "Failed to get Provider Profile", err, 500, w, r)
		return
	}

	thisUser, handleErr := thisIdentityProvider.Handle(w, r)
	if handleErr != nil {
		HandleError(thisIdentityProvider.GetLogTag(), handleErr.Error(), handleErr, 401, w, r)
	}

	// Create Session in Tyk
	handler := thisIdentityProvider.GetHandler()
	asJson, actionErr := handler.CompleteIdentityAction(w, r, thisUser, thisIdentityProvider.GetProfile())

	if actionErr != nil {
		HandleError(HandlerLogTag, "Could not retrieve ID", actionErr, 500, w, r)
		return
	}

	cors := thisIdentityProvider.GetCORS()

	if cors {
		corsOrigins := thisIdentityProvider.GetCORSOrigin()
		w.Header().Set("Access-Control-Allow-Origin", corsOrigins)
	}
	w.WriteHeader(200)
	fmt.Fprintf(w, string(asJson))
	return
}

// HandleAuthCallback Is a callback URL passed to OAuth providers such as Social, handles completing an auth request
func HandleAuthCallback(w http.ResponseWriter, r *http.Request) {

	thisIdentityProvider, err := getTapProfile(w, r)
	if err != nil {
		HandleError(HandlerLogTag, "Failed to get Provider Profile", err, 500, w, r)
		return
	}

	thisIdentityProvider.HandleCallback(w, r, HandleError)
	return
}

func HandleCORS(w http.ResponseWriter, r *http.Request) {
	thisIdentityProvider, err := getTapProfile(w, r)
	if err != nil {
		HandleError(HandlerLogTag, "Failed to get Provider Profile", err, 500, w, r)
		return
	}

	cors := thisIdentityProvider.GetCORS()

	if cors {

		corsOrigins := thisIdentityProvider.GetCORSOrigin()
		corsHeaders := thisIdentityProvider.GetCORSHeaders()
		corsMaxAge := thisIdentityProvider.GetCORSMaxAge()

		w.Header().Set("Access-Control-Allow-Origin", corsOrigins)
		w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, POST")
		w.Header().Set("Access-Control-Allow-Headers", corsHeaders)
		w.Header().Set("Access-Control-Max-Age", corsMaxAge)

		w.Header().Set("Allow", "OPTIONS, POST")
		w.WriteHeader(200)
	} else {
		w.WriteHeader(500)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, "{\"Status\":\"error\",\"Error\": \"CORS is not enabled for this profile\"}")
	}
}
