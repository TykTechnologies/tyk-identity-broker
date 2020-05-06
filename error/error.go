package error

import (
	"encoding/json"
	"fmt"
	logger "github.com/TykTechnologies/tyk-identity-broker/log"
	"github.com/sirupsen/logrus"
	"net/http"
)

var log = logger.Get()

// APIErrorMessage is an object that defines when a generic error occurred
type APIErrorMessage struct {
	Status string
	Error  string
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

