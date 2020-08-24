package tap

import (
	"errors"
	logger "github.com/TykTechnologies/tyk-identity-broker/log"
	"net/http"
)

var log = logger.Get()

type HttpError struct{
	Message string
	Code int
	Error error
}

func AddProfile(profile Profile, AuthConfigStore AuthRegisterBackend, flush func(backend AuthRegisterBackend) error) *HttpError {
	dumpProfile := Profile{}
	keyErr := AuthConfigStore.GetKey(profile.ID,profile.OrgID, &dumpProfile)
	if keyErr == nil {
		return &HttpError{
			Message: "Object ID already exists",
			Code:    http.StatusBadRequest,
			Error:   keyErr,
		}
	}

	saveErr := AuthConfigStore.SetKey(profile.ID,profile.OrgID, &profile)
	if saveErr != nil {
		return &HttpError{
			Message: "Update failed",
			Code:    http.StatusInternalServerError,
			Error:   saveErr,
		}
	}

	fErr := flush(AuthConfigStore)
	if fErr != nil {
		return &HttpError{
			Message: "flush failed",
			Code:    http.StatusBadRequest,
			Error:   fErr,
		}
	}

	return nil
}

func UpdateProfile(key string, profile Profile, AuthConfigStore AuthRegisterBackend,flush func(backend AuthRegisterBackend) error) *HttpError {

	// Shenanigans
	if profile.ID != key {
		return &HttpError{
			Message: "Object ID and URI resource ID do not match",
			Code:    http.StatusBadRequest,
			Error:   errors.New("ID Mismatch"),
		}
	}

	dumpProfile := Profile{}
	keyErr := AuthConfigStore.GetKey(key,profile.OrgID, &dumpProfile)
	if keyErr != nil {
		return &HttpError{
			Message: "Object ID does not exist, operation not permitted",
			Code:    http.StatusNotFound,
			Error:   keyErr,
		}
	}

	saveErr := AuthConfigStore.SetKey(key,profile.OrgID, &profile)
	if saveErr != nil {
		return &HttpError{
			Message: "Update failed",
			Code:    http.StatusInternalServerError,
			Error:   saveErr,
		}
	}

	fErr := flush(AuthConfigStore)
	if fErr != nil {
		return &HttpError{
			Message: "flush failed",
			Code:    http.StatusBadRequest,
			Error:   fErr,
		}
	}

	return nil
}

func DeleteProfile(key, orgID string,AuthConfigStore AuthRegisterBackend, flush func(backend AuthRegisterBackend) error) *HttpError {

	dumpProfile := Profile{}
	keyErr := AuthConfigStore.GetKey(key,orgID, &dumpProfile)
	if keyErr != nil {
		return &HttpError{
			Message: "Object ID does not exist",
			Code:    http.StatusNotFound,
			Error:  keyErr,
		}
	}

	delErr := AuthConfigStore.DeleteKey(key, orgID)
	if delErr != nil {
		return &HttpError{
			Message: "Delete failed",
			Code:    http.StatusInternalServerError,
			Error:   delErr,
		}
	}

	fErr := flush(AuthConfigStore)
	if fErr != nil {
		return &HttpError{
			Message: "flush failed",
			Code:    http.StatusBadRequest,
			Error:   fErr,
		}
	}
	return nil
}


