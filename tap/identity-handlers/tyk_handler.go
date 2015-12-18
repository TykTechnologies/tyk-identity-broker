package identityHandlers

import (
	"errors"
	"fmt"
	"github.com/lonelycode/tyk-auth-proxy/tap"
	"github.com/lonelycode/tyk-auth-proxy/tyk-api"
	"net/http"
)

var TykAPILogTag string = "[TYK ID HANDLER]"

type ModuleName string

const (
	SSOForDashboard ModuleName = "dashboard"
	SSOForPortal    ModuleName = "portal"
)

type SSOAccessData struct {
	ForSection ModuleName
	OrgID      string
}

type TykIdentityHandler struct {
	API     *tyk.TykAPI
	profile tap.Profile
}

func mapActionToModule(action tap.Action) (ModuleName, error) {
	switch action {
	case tap.GenerateOrLoginUserProfile:
		return SSOForDashboard, nil
	case tap.GenerateOrLoginDeveloperProfile:
		return SSOForPortal, nil
	}

	log.Error(TykAPILogTag+"Action: ", action)
	return SSOForPortal, errors.New("Action does not exist")
}

func (t *TykIdentityHandler) Init(conf interface{}) error {
	t.profile = conf.(tap.Profile)

	return nil
}

func (t *TykIdentityHandler) CreateIdentity(i interface{}) (string, error) {
	log.Info(TykAPILogTag+" Creating identity for: ", i)

	thisModule, modErr := mapActionToModule(t.profile.ActionType)
	if modErr != nil {
		log.Error(TykAPILogTag+" Failed to assign module: ", modErr)
		return "", modErr
	}

	accessRequest := SSOAccessData{
		ForSection: thisModule,
		OrgID:      t.profile.OrgID,
	}

	returnVal, retErr := t.API.CreateSSONonce(tyk.SSO, accessRequest)
	log.Warning("Returned: ", returnVal)
	asMapString := returnVal.(map[string]interface{})
	if retErr != nil {
		log.Error(TykAPILogTag+" API Response error: ", retErr)
		return "", retErr
	}
	return asMapString["Meta"].(string), nil
}

func (t *TykIdentityHandler) LoginIdentity(user string, pass string) (string, error) {
	// Not used
	return "", nil
}

func (t *TykIdentityHandler) CompleteIdentityAction(w http.ResponseWriter, r *http.Request, i interface{}, profile tap.Profile) {
	nonce, nErr := t.CreateIdentity(i)

	if nErr != nil {
		log.Error(TykAPILogTag+" Nonce creation failed: ", nErr)
		fmt.Fprintf(w, "Login failed")
		return
	}

	// After login, we need to redirect this user
	log.Debug(TykAPILogTag + " --> Running redirect...")
	if profile.ReturnURL != "" {
		newURL := profile.ReturnURL + "?nonce=" + nonce
		http.Redirect(w, r, newURL, 301)
		return
	}

	log.Warning(TykAPILogTag + "No return URL found, redirect failed.")
	fmt.Fprintf(w, "Success! (Have you set a return URL?)")
}
