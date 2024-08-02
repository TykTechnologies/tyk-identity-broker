package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/TykTechnologies/tyk-identity-broker/initializer"

	tykerror "github.com/TykTechnologies/tyk-identity-broker/error"

	"github.com/TykTechnologies/tyk-identity-broker/tap"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

var APILogTag string = "API"

type APIOKMessage struct {
	Status string
	ID     string
	Data   interface{}
}

func HandleAPIOK(data interface{}, id string, code int, w http.ResponseWriter, r *http.Request) {
	okObj := APIOKMessage{
		Status: "ok",
		ID:     id,
		Data:   data,
	}

	responseMsg, err := json.Marshal(&okObj)

	if err != nil {
		log.WithFields(logrus.Fields{
			"prefix": APILogTag,
			"error":  err,
		}).Error("[OK Handler] Couldn't marshal message stats")
		fmt.Fprintf(w, "System Error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	fmt.Fprintf(w, string(responseMsg))
}

func HandleAPIError(tag string, errorMsg string, rawErr error, code int, w http.ResponseWriter, r *http.Request) {
	log.WithFields(logrus.Fields{
		"prefix": tag,
		"error":  errorMsg,
	}).Error(rawErr)

	errorObj := tykerror.APIErrorMessage{Status: "error", Error: errorMsg}
	responseMsg, err := json.Marshal(&errorObj)

	if err != nil {
		log.WithFields(logrus.Fields{
			"prefix": tag,
			"error":  err,
		}).Error("[Error Handler] Couldn't marshal error stats")
		fmt.Fprintf(w, "System Error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	fmt.Fprintf(w, string(responseMsg))
}

// ------ Middleware methods -------
func IsAuthenticated(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != config.Secret {
			HandleAPIError(APILogTag, "Authorization failed", errors.New("Header mismatch"), 401, w, r)
			return
		}

		h.ServeHTTP(w, r)
	})
}

// ------ End Middleware methods -------

func HandleGetProfileList(w http.ResponseWriter, r *http.Request) {
	profiles := initializer.AuthConfigStore.GetAll("")

	HandleAPIOK(profiles, "", 200, w, r)
}

func HandleGetProfile(w http.ResponseWriter, r *http.Request) {
	key := mux.Vars(r)["id"]
	thisProfile := tap.Profile{}

	keyErr := initializer.AuthConfigStore.GetKey(key, thisProfile.OrgID, &thisProfile)
	if keyErr != nil {
		HandleAPIError(APILogTag, "Profile not found", keyErr, 404, w, r)
		return
	}

	HandleAPIOK(thisProfile, key, 200, w, r)
}

func HandleAddProfile(w http.ResponseWriter, r *http.Request) {
	key := mux.Vars(r)["id"]

	profileData, err := ioutil.ReadAll(r.Body)
	if err != nil {
		HandleAPIError(APILogTag, "Invalid request data", err, 400, w, r)
		return
	}

	thisProfile := tap.Profile{}
	decodeErr := json.Unmarshal(profileData, &thisProfile)
	if decodeErr != nil {
		HandleAPIError(APILogTag, "Failed to decode body data", decodeErr, 400, w, r)
		return
	}

	if thisProfile.ID != key {
		HandleAPIError(APILogTag, "Object ID and URI resource ID do not match", errors.New("ID Mismatch"), 400, w, r)
		return
	}

	httpErr := tap.AddProfile(thisProfile, initializer.AuthConfigStore, GlobalDataLoader.Flush)
	if httpErr != nil {
		HandleAPIError(APILogTag, httpErr.Message, httpErr.Error, httpErr.Code, w, r)
		return
	}

	HandleAPIOK(thisProfile, key, 201, w, r)
}

func HandleUpdateProfile(w http.ResponseWriter, r *http.Request) {
	key := mux.Vars(r)["id"]

	profileData, err := ioutil.ReadAll(r.Body)
	if err != nil {
		HandleAPIError(APILogTag, "Invalid request data", err, 400, w, r)
		return
	}

	thisProfile := tap.Profile{}
	decodeErr := json.Unmarshal(profileData, &thisProfile)
	if decodeErr != nil {
		HandleAPIError(APILogTag, "Failed to decode body data", decodeErr, 400, w, r)
		return
	}

	updateErr := tap.UpdateProfile(key, thisProfile, initializer.AuthConfigStore, GlobalDataLoader.Flush)
	if updateErr != nil {
		HandleAPIError(APILogTag, updateErr.Message, updateErr.Error, updateErr.Code, w, r)
		return
	}

	HandleAPIOK(thisProfile, key, 200, w, r)
}

func HandleDeleteProfile(w http.ResponseWriter, r *http.Request) {
	key := mux.Vars(r)["id"]
	err := tap.DeleteProfile(key, "", initializer.AuthConfigStore, GlobalDataLoader.Flush)
	if err != nil {
		HandleAPIError(APILogTag, err.Message, err.Error, err.Code, w, r)
		return
	}

	data := make(map[string]string)
	HandleAPIOK(data, key, 200, w, r)
}
