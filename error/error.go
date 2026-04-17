package error

import (
	"encoding/json"
	"net/http"

	logger "github.com/TykTechnologies/tyk-identity-broker/log"
	"github.com/sirupsen/logrus"
)

var log = logger.Get()

// APIErrorMessage is an object that defines when a generic error occurred
type APIErrorMessage struct {
	Status string
	Error  string
}

// HandleError is a generic error handler
func HandleError(tag string, errorMsg string, rawErr error, code int, w http.ResponseWriter, _ *http.Request) {
	log.WithFields(logrus.Fields{
		"prefix":   tag,
		"errorMsg": errorMsg,
	}).Error(rawErr)

	errorObj := APIErrorMessage{"error", errorMsg}
	responseMsg, err := json.Marshal(&errorObj)

	if err != nil {
		log.WithField("prefix", tag).Error("[Error Handler] Couldn't marshal error stats: ", err)
		w.Write([]byte("System Error")) //nolint:errcheck
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(responseMsg) //nolint:errcheck
}
