package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/lonelycode/tyk-auth-proxy/tap"
	"io/ioutil"
	"net/http"
)

var APILogTag string = "[API]"

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
		log.Error("[OK Handler] Couldn't marshal message stats: ", err)
		fmt.Fprintf(w, "System Error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	fmt.Fprintf(w, string(responseMsg))
}

func HandleAPIError(tag string, errorMsg string, rawErr error, code int, w http.ResponseWriter, r *http.Request) {
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
	profiles := AuthConfigStore.GetAll()

	HandleAPIOK(profiles, "", 200, w, r)
}

func HandleGetProfile(w http.ResponseWriter, r *http.Request) {
	key := mux.Vars(r)["id"]
	thisProfile := tap.Profile{}

	keyErr := AuthConfigStore.GetKey(key, &thisProfile)
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

	dumpProfile := tap.Profile{}
	keyErr := AuthConfigStore.GetKey(key, &dumpProfile)
	if keyErr == nil {
		HandleAPIError(APILogTag, "Object ID already exists", keyErr, 400, w, r)
		return
	}

	saveErr := AuthConfigStore.SetKey(key, &thisProfile)
	if saveErr != nil {
		HandleAPIError(APILogTag, "Update failed", saveErr, 500, w, r)
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

	if thisProfile.ID != key {
		HandleAPIError(APILogTag, "Object ID and URI resource ID do not match", errors.New("ID Mismatch"), 400, w, r)
		return
	}

	dumpProfile := tap.Profile{}
	keyErr := AuthConfigStore.GetKey(key, &dumpProfile)
	if keyErr != nil {
		HandleAPIError(APILogTag, "Object ID does not exist, operation not permnitted", keyErr, 400, w, r)
		return
	}

	saveErr := AuthConfigStore.SetKey(key, &thisProfile)
	if saveErr != nil {
		HandleAPIError(APILogTag, "Update failed", saveErr, 500, w, r)
		return
	}

	HandleAPIOK(thisProfile, key, 200, w, r)
}

func HandleDeleteProfile(w http.ResponseWriter, r *http.Request) {
	key := mux.Vars(r)["id"]

	dumpProfile := tap.Profile{}
	keyErr := AuthConfigStore.GetKey(key, &dumpProfile)
	if keyErr != nil {
		HandleAPIError(APILogTag, "Object ID does not exist", keyErr, 400, w, r)
		return
	}

	delErr := AuthConfigStore.DeleteKey(key)
	if delErr != nil {
		HandleAPIError(APILogTag, "Delete failed", delErr, 500, w, r)
		return
	}

	data := make(map[string]string)
	HandleAPIOK(data, key, 200, w, r)
}
