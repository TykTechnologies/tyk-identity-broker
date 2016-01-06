package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

var APILogTag string = "[API]"

// HandleAPIError is a generic API error handler
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

func IsAuthenticated(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != config.Secret {
			HandleAPIError(APILogTag, "Authorization failed", errors.New("Header mismatch"), 401, w, r)
			return
		}

		h.ServeHTTP(w, r)
	})
}

func HandleGetProfileList(w http.ResponseWriter, r *http.Request) {

	return
}
