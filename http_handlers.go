package main

import (
	"errors"
	"net/http"

	"github.com/TykTechnologies/tyk-identity-broker/constants"
	"github.com/TykTechnologies/tyk-identity-broker/providers"

	tykerrors "github.com/TykTechnologies/tyk-identity-broker/error"
	"github.com/gorilla/mux"
)

// Returns a profile ID
func getId(req *http.Request) (string, error) {
	id := mux.Vars(req)["id"]
	if id == "" {
		id = mux.Vars(req)[":id"]
	}
	if id == "" {
		return id, errors.New("no profile id detected")
	}
	return id, nil

}

func HandleAuthSAMLLogon(w http.ResponseWriter, r *http.Request) {
	//thisId, idErr := getId(r)
	//if idErr != nil {
	//	tykerrors.HandleError(constants.HandlerLogTag, "Could not retrieve ID", idErr, 400, w, r)
	//	return
	//}
	//
	//thisIdentityProvider, err := providers.GetTapProfile(AuthConfigStore, IdentityKeyStore, thisId, TykAPIHandler)
	//if err != nil {
	//	tykerrors.HandleError(constants.HandlerLogTag, err.Message, err.Error, err.Code, w, r)
	//	return
	//}

	return
}

//does nothing - extend tap interface
func HandleSAMLMetadata(w http.ResponseWriter, r *http.Request) {
	thisId, idErr := getId(r)
	if idErr != nil {
		tykerrors.HandleError(constants.HandlerLogTag, "Could not retrieve ID", idErr, 400, w, r)
		return
	}

	thisIdentityProvider, err := providers.GetTapProfile(AuthConfigStore, IdentityKeyStore, thisId, TykAPIHandler)
	if err != nil {
		tykerrors.HandleError(constants.HandlerLogTag, err.Message, err.Error, err.Code, w, r)
		return
	}
	switch thisIdentityProvider.Name() {
	case "SAMLProvider":
	default:
		return
	}

}

// HandleAuth is the main entry point handler for any profile (i.e. /auth/:profile-id/:provider)
func HandleAuth(w http.ResponseWriter, r *http.Request) {

	thisId, idErr := getId(r)
	if idErr != nil {
		tykerrors.HandleError(constants.HandlerLogTag, "Could not retrieve ID", idErr, 400, w, r)
		return
	}

	thisIdentityProvider, err := providers.GetTapProfile(AuthConfigStore, IdentityKeyStore, thisId, TykAPIHandler)
	if err != nil {
		return
	}

	pathParams := mux.Vars(r)
	thisIdentityProvider.Handle(w, r, pathParams)
	return
}

// HandleAuthCallback Is a callback URL passed to OAuth providers such as Social, handles completing an auth request
func HandleAuthCallback(w http.ResponseWriter, r *http.Request) {

	thisId, idErr := getId(r)
	if idErr != nil {
		tykerrors.HandleError(constants.HandlerLogTag, "Could not retrieve ID", idErr, 400, w, r)
		return
	}

	thisIdentityProvider, err := providers.GetTapProfile(AuthConfigStore, IdentityKeyStore, thisId, TykAPIHandler)
	if err != nil {
		tykerrors.HandleError(constants.HandlerLogTag, err.Message, err.Error, err.Code, w, r)
		return
	}
	thisIdentityProvider.HandleCallback(w, r, tykerrors.HandleError)
	return
}

func HandleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
