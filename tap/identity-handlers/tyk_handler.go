package identityHandlers

import (
	"errors"
	"fmt"
	"github.com/lonelycode/tyk-auth-proxy/tap"
	"github.com/lonelycode/tyk-auth-proxy/tyk-api"
	"github.com/markbates/goth"
	"net/http"
	"time"
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
	API                  *tyk.TykAPI
	profile              tap.Profile
	dashboardUserAPICred string
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
	if conf.(tap.Profile).IdentityHandlerConfig != nil {
		t.dashboardUserAPICred = conf.(tap.Profile).IdentityHandlerConfig.(map[string]interface{})["DashboardCredential"].(string)
	}

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

func (t *TykIdentityHandler) CompleteIdentityActionForDashboard(w http.ResponseWriter, r *http.Request, i interface{}, profile tap.Profile) {
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

func (t *TykIdentityHandler) CompleteIdentityActionForPortal(w http.ResponseWriter, r *http.Request, i interface{}, profile tap.Profile) {
	// Create a nonce
	log.Info(TykAPILogTag + " Creating nonce")
	nonce, nErr := t.CreateIdentity(i)

	if nErr != nil {
		log.Error(TykAPILogTag+" Nonce creation failed: ", nErr)
		fmt.Fprintf(w, "Login failed")
		return
	}

	// Check if user exists
	userEmail := i.(goth.User).Email
	thisUser, retErr := t.API.GetDeveloper(t.dashboardUserAPICred, userEmail)
	log.Warning(TykAPILogTag+" Returned: ", thisUser)

	createUser := false
	if retErr != nil {
		log.Warning(TykAPILogTag+" API Error: ", nErr)
		log.Info(TykAPILogTag + " User not found, creating new record")
		createUser = true
	}

	// If not, create user
	if createUser {
		log.Info(TykAPILogTag + " Creating user")
		newUser := tyk.PortalDeveloper{
			Email:         i.(goth.User).Email,
			Password:      "",
			DateCreated:   time.Now(),
			OrgId:         t.profile.OrgID,
			ApiKeys:       map[string]string{},
			Subscriptions: map[string]string{},
			Fields:        map[string]string{},
			Nonce:         nonce,
		}
		createErr := t.API.CreateDeveloper(t.dashboardUserAPICred, newUser)
		if createErr != nil {
			log.Error(TykAPILogTag+" failed to create user! ", createErr)
			fmt.Fprintf(w, "Login failed")
			return
		}
	} else {
		// Set nonce value in user profile
		thisUser.Nonce = nonce
		updateErr := t.API.UpdateDeveloper(t.dashboardUserAPICred, thisUser)
		if updateErr != nil {
			log.Error("Failed to update user! ", updateErr)
			fmt.Fprintf(w, "Login failed")
			return
		}
	}

	// After login, we need to redirect this user
	log.Info(TykAPILogTag + " --> Running redirect...")
	if profile.ReturnURL != "" {
		newURL := profile.ReturnURL + "?nonce=" + nonce
		log.Info(TykAPILogTag+" --> URL With NONCE is: ", newURL)
		http.Redirect(w, r, newURL, 301)
		return
	}

	log.Warning(TykAPILogTag + "No return URL found, redirect failed.")
	fmt.Fprintf(w, "Success! (Have you set a return URL?)")
}

func (t *TykIdentityHandler) CompleteIdentityAction(w http.ResponseWriter, r *http.Request, i interface{}, profile tap.Profile) {
	if profile.ActionType == tap.GenerateOrLoginUserProfile {
		t.CompleteIdentityActionForDashboard(w, r, i, profile)
		return
	} else if profile.ActionType == tap.GenerateOrLoginDeveloperProfile {
		t.CompleteIdentityActionForPortal(w, r, i, profile)
		return
	}
}
