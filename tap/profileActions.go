package tap

import (
	"errors"
	logger "github.com/TykTechnologies/tyk-identity-broker/log"
)

var log = logger.Get()

type HttpError struct{
	Message string
	Code int
	Error error
}

func AddProfile(profile Profile, AuthConfigStore AuthRegisterBackend, flush func(backend AuthRegisterBackend) error) *HttpError {
	dumpProfile := Profile{}
	keyErr := AuthConfigStore.GetKey(profile.ID, &dumpProfile)
	if keyErr != nil {
		return &HttpError{
			Message: "Object ID already exists",
			Code:    400,
			Error:   keyErr,
		}
	}

	saveErr := AuthConfigStore.SetKey(profile.ID, &profile)
	if saveErr != nil {
		return &HttpError{
			Message: "Update failed",
			Code:    500,
			Error:   saveErr,
		}
	}

	fErr := flush(AuthConfigStore)
	if fErr != nil {
		return &HttpError{
			Message: "flush failed",
			Code:    400,
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
			Code:    400,
			Error:   errors.New("ID Mismatch"),
		}
	}

	dumpProfile := Profile{}
	keyErr := AuthConfigStore.GetKey(key, &dumpProfile)
	if keyErr != nil {
		return &HttpError{
			Message: "Object ID does not exist, operation not permitted",
			Code:    400,
			Error:   keyErr,
		}
	}

	saveErr := AuthConfigStore.SetKey(key, &profile)
	if saveErr != nil {
		return &HttpError{
			Message: "Update failed",
			Code:    500,
			Error:   saveErr,
		}
	}

	return nil
}

func DeleteProfile(key string,AuthConfigStore AuthRegisterBackend, flush func(backend AuthRegisterBackend) error) *HttpError {

	dumpProfile := Profile{}
	keyErr := AuthConfigStore.GetKey(key, &dumpProfile)
	if keyErr != nil {
		return &HttpError{
			Message: "Object ID does not exist",
			Code:    400,
			Error:  keyErr,
		}
	}

	delErr := AuthConfigStore.DeleteKey(key)
	if delErr != nil {
		return &HttpError{
			Message: "Delete failed",
			Code:    500,
			Error:   delErr,
		}
	}

	fErr := flush(AuthConfigStore)
	if fErr != nil {
		return &HttpError{
			Message: "flush failed",
			Code:    400,
			Error:   fErr,
		}
	}
	return nil
}


