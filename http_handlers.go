package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/lonelycode/tyk-auth-proxy/providers"
	"github.com/lonelycode/tyk-auth-proxy/tap"
	"github.com/lonelycode/tyk-auth-proxy/tap/identity-handlers"
	"net/http"
)

// APIErrorMessage is an object that defines when a generic error occurred
type APIErrorMessage struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}

var HandlerLogTag = "[AUTH HANDLERS]"

func getId(req *http.Request) (string, error) {
	id := req.URL.Query().Get("id")
	if id == "" {
		id = req.URL.Query().Get(":id")
	}
	if id == "" {
		return id, errors.New("No profile id detected")
	}
	return id, nil
}

func getIdentityHandler(name tap.Action) tap.IdentityHandler {
	var thisIdentityHandler tap.IdentityHandler

	switch name {
	case tap.GenerateOrLoginDeveloperProfile:
		thisIdentityHandler = identityHandlers.DummyIdentityHandler{} // TODO: Change These
	case tap.GenerateOrLoginUserProfile:
		thisIdentityHandler = identityHandlers.DummyIdentityHandler{} // TODO: Change These
	}

	return thisIdentityHandler
}

func getTAProvider(conf tap.Profile) tap.TAProvider {

	var thisProvider tap.TAProvider

	switch conf.ProviderName {
	case "SocialProvider":
		thisProvider = &providers.Social{}
	}

	var thisIdentityHandler tap.IdentityHandler = getIdentityHandler(conf.ActionType)

	thisProvider.Init(thisIdentityHandler, conf, []byte(conf.ProviderConfig))
	return thisProvider

}

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

func HandleAuth(w http.ResponseWriter, r *http.Request) {
	thisIdentityProvider, err := getTapProfile(w, r)
	if err != nil {
		return
	}

	thisIdentityProvider.Handle(w, r)
	return
}

func HandleAuthCallback(w http.ResponseWriter, r *http.Request) {

	thisIdentityProvider, err := getTapProfile(w, r)
	if err != nil {
		return
	}

	thisIdentityProvider.HandleCallback(w, r, HandleSuccess, HandleError)
	return
}

func HandleSuccess(w http.ResponseWriter, r *http.Request, user interface{}, profile tap.Profile) {
	var thisIdentityHandler tap.IdentityHandler = getIdentityHandler(profile.ActionType)

	log.Warning(HandlerLogTag+" --> Developer Data: ", user)

	// Lets finish things, if we have a user, log in and create the identities
	identityErr := thisIdentityHandler.CompleteIdentityAction(user)
	if identityErr != nil {
		HandleError(HandlerLogTag, "Failed to complete identity transaction", identityErr, 400, w, r)
	}

	if profile.ReturnURL != "" {
		http.Redirect(w, r, profile.ReturnURL, 301)
		return
	}

	log.Warning("No return URL found, redirect failed.")
	fmt.Fprintf(w, "Success! (Have you set a return URL?)")
}
