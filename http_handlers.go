package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/TykTechnologies/tyk-identity-broker/constants"
	"github.com/TykTechnologies/tyk-identity-broker/providers"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

// APIErrorMessage is an object that defines when a generic error occurred
type APIErrorMessage struct {
	Status string
	Error  string
}

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

// HandleAuth is the main entrypoint handler for any profile (i.e. /auth/:profile-id/:provider)
func HandleAuth(w http.ResponseWriter, r *http.Request) {

	thisId, idErr := getId(r)
	if idErr != nil {
		HandleError(constants.HandlerLogTag, "Could not retrieve ID", idErr, 400, w, r)
		return
	}

	thisIdentityProvider, err := providers.GetTapProfile( AuthConfigStore, IdentityKeyStore, thisId, TykAPIHandler)
	if err != nil {
		return
	}

	thisIdentityProvider.Handle(w, r)
	return
}

// HandleAuthCallback Is a callback URL passed to OAuth providers such as Social, handles completing an auth request
func HandleAuthCallback(w http.ResponseWriter, r *http.Request) {

	thisId, idErr := getId(r)
	if idErr != nil {
		HandleError(constants.HandlerLogTag, "Could not retrieve ID", idErr, 400, w, r)
		return
	}

	thisIdentityProvider, err := providers.GetTapProfile( AuthConfigStore, IdentityKeyStore, thisId, TykAPIHandler)
	if err != nil {
		HandleError(constants.HandlerLogTag, err.Message, err.Error, err.Code, w, r)
		return
	}

	thisIdentityProvider.HandleCallback(w, r, HandleError)
	return
}

func HandleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

