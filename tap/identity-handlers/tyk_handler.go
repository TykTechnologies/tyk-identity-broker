/* package identityHandlers provides a collection of handlers for target systems,
these handlers create accounts and sso tokens */
package identityHandlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/lonelycode/tyk-auth-proxy/tap"
	"github.com/lonelycode/tyk-auth-proxy/tyk-api"
	"github.com/markbates/goth"
	"net/http"
	"time"
)

var TykAPILogTag string = "[TYK ID HANDLER]" // log tag

type ModuleName string // To separate out target modules of the dashboard

const (
	// Enums to identify which target it being used, dashbaord or portal, they are distinct.
	SSOForDashboard ModuleName = "dashboard"
	SSOForPortal    ModuleName = "portal"
)

// SSOAccessData is the data type used for speaking to the SSO endpoint in the advanced API
type SSOAccessData struct {
	ForSection ModuleName
	OrgID      string
}

// TykIdentityHandler provides an interface for generating SSO identities on a tyk node
type TykIdentityHandler struct {
	API                   *tyk.TykAPI
	Store                 tap.AuthRegisterBackend
	profile               tap.Profile
	dashboardUserAPICred  string
	oauth                 OAuthSettings
	token                 TokenSettings
	disableOneTokenPerAPI bool
}

// OAuthSettings determine the OAuth parameters for the tap.GenerateOAuthTokenForClient action
type OAuthSettings struct {
	APIListenPath string
	RedirectURI   string
	ResponseType  string
	ClientId      string
	Secret        string
	BaseAPIID     string
}

type TokenSettings struct {
	BaseAPIID string
	Expires   int64
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

// initialise th Tyk handler, the Tyk handler *requires* initialisation with the TykAPI handler global set
// up in main
func (t *TykIdentityHandler) Init(conf interface{}) error {
	t.profile = conf.(tap.Profile)
	if conf.(tap.Profile).IdentityHandlerConfig != nil {
		theseConfs := conf.(tap.Profile).IdentityHandlerConfig.(map[string]interface{})
		if theseConfs["DashboardCredential"] != nil {
			t.dashboardUserAPICred = theseConfs["DashboardCredential"].(string)
		}

		if theseConfs["DisableOneTokenPerAPI"] != nil {
			t.disableOneTokenPerAPI = theseConfs["DisableOneTokenPerAPI"].(bool)
		}

		oauthSettings, ok := theseConfs["OAuth"]
		if ok {
			log.Debug(TykAPILogTag + "Found Oauth configuration, loading...")
			t.oauth = OAuthSettings{}
			t.oauth.APIListenPath = oauthSettings.(map[string]interface{})["APIListenPath"].(string)
			t.oauth.RedirectURI = oauthSettings.(map[string]interface{})["RedirectURI"].(string)
			t.oauth.ResponseType = oauthSettings.(map[string]interface{})["ResponseType"].(string)
			t.oauth.ClientId = oauthSettings.(map[string]interface{})["ClientId"].(string)
			t.oauth.Secret = oauthSettings.(map[string]interface{})["Secret"].(string)
			t.oauth.BaseAPIID = oauthSettings.(map[string]interface{})["BaseAPIID"].(string)
		}

		tokenSettings, tokenOk := theseConfs["TokenAuth"]
		if tokenOk {
			if tokenSettings.(map[string]interface{})["BaseAPIID"] == nil {
				log.Error(TykAPILogTag + " Base API is empty!")
				return errors.New("Base API cannot be empty")
			}
			t.token.BaseAPIID = tokenSettings.(map[string]interface{})["BaseAPIID"].(string)

			if tokenSettings.(map[string]interface{})["ExpirySeconds"] == nil {
				log.Warning(TykAPILogTag + " No expiry found - defaulting to 3600 seconds")
				t.token.Expires = 3600
			} else {
				t.token.Expires = int64(tokenSettings.(map[string]interface{})["ExpirySeconds"].(float64))
			}

		}
	}

	return nil
}

// CreateIdentity will generate an SSO token that can be used with the tyk SSO endpoints for dash or portal.
// Identity is assumed to be a goth.User object as this is what we arestnadardiseing on.
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
	asMapString := returnVal.(map[string]interface{})
	if retErr != nil {
		log.Error(TykAPILogTag+" API Response error: ", retErr)
		return "", retErr
	}
	return asMapString["Meta"].(string), nil
}

// CompleteIdentityActionForDashboard handles a dashboard identity. No ise is created, only an SSO login session
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

// CompleteIdentityActionForPortal will generate an identity for a portal user based, so it will AddOrUpdate that
// user depnding on if they exist or not and validate the login using a one-time nonce.
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
	sso_key := tap.GenerateSSOKey(i.(goth.User))
	thisUser, retErr := t.API.GetDeveloperBySSOKey(t.dashboardUserAPICred, sso_key)
	log.Warning(TykAPILogTag+" Returned: ", thisUser)

	createUser := false
	if retErr != nil {
		log.Warning(TykAPILogTag+" API Error: ", nErr)
		log.Info(TykAPILogTag + " User not found, creating new record")
		createUser = true
	}

	// If not, create user
	if createUser {
		if thisUser.Email == "" {
			thisUser.Email = sso_key
		}

		log.Info(TykAPILogTag + " Creating user")
		newUser := tyk.PortalDeveloper{
			Email:         thisUser.Email,
			Password:      "",
			DateCreated:   time.Now(),
			OrgId:         t.profile.OrgID,
			ApiKeys:       map[string]string{},
			Subscriptions: map[string]string{},
			Fields:        map[string]string{},
			Nonce:         nonce,
			SSOKey:        sso_key,
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

func (t *TykIdentityHandler) CompleteIdentityActionForOAuth(w http.ResponseWriter, r *http.Request, i interface{}, profile tap.Profile) {
	log.Info(TykAPILogTag + " Starting OAuth Flow...")

	// Generate identity key match ID
	sso_key := tap.GenerateSSOKey(i.(goth.User))
	id_with_profile := t.profile.ID + "-" + sso_key
	// Check if key already exists

	value := ""
	log.Debug("Store is: ", t.Store)
	log.Debug("ID IS: ", id_with_profile)

	if !t.disableOneTokenPerAPI {
		fErr := t.Store.GetKey(id_with_profile, &value)
		if fErr == nil {
			// Key found
			log.Warning(TykAPILogTag + " --> Token exists, invalidating")
			iErr := t.API.InvalidateToken(t.dashboardUserAPICred, t.oauth.BaseAPIID, value)
			if iErr != nil {
				log.Error(TykAPILogTag+" ----> Token Invalidation failed: ", iErr)
			}
		}
	}

	// Generate OAuth
	resp, oErr := t.API.RequestOAuthToken(t.oauth.APIListenPath,
		t.oauth.RedirectURI,
		t.oauth.ResponseType,
		t.oauth.ClientId,
		t.oauth.Secret,
		t.profile.OrgID,
		t.profile.MatchedPolicyID,
		t.oauth.BaseAPIID,
		i)

	// Redirect request
	if oErr != nil {
		log.Error("Failed to generate OAuth token ", oErr)
		fmt.Fprintf(w, "OAuth token generation failed")
		return
	}

	if resp == nil {
		log.Error(TykAPILogTag + " --> Login failure. Request not allowed")
		fmt.Fprintf(w, "Login failed")
		return
	}

	if resp.AccessToken != "" {
		log.Warning(TykAPILogTag + " --> Storing token reference")
		t.Store.SetKey(id_with_profile, resp.AccessToken)
	}

	// After login, we need to redirect this user
	log.Info(TykAPILogTag + " --> Running oauth redirect...")
	if resp.RedirectTo != "" {
		log.Debug(TykAPILogTag+" --> URL is: ", resp.RedirectTo)
		http.Redirect(w, r, resp.RedirectTo, 301)
		return
	}
}

func (t *TykIdentityHandler) CompleteIdentityActionForTokenAuth(w http.ResponseWriter, r *http.Request, i interface{}, profile tap.Profile) {
	log.Info(TykAPILogTag + " Starting Token Flow...")

	// Generate identity key match ID
	sso_key := tap.GenerateSSOKey(i.(goth.User))
	id_with_profile := t.profile.ID + "-" + sso_key
	// Check if key already exists

	value := ""
	log.Debug("Store is: ", t.Store)
	log.Debug("ID IS: ", id_with_profile)

	if !t.disableOneTokenPerAPI {
		fErr := t.Store.GetKey(id_with_profile, &value)
		if fErr == nil {
			// Key found
			log.Warning(TykAPILogTag + " --> Token exists, invalidating")
			iErr := t.API.InvalidateToken(t.dashboardUserAPICred, t.token.BaseAPIID, value)
			if iErr != nil {
				log.Error(TykAPILogTag+" ----> Token Invalidation failed: ", iErr)
			}
		}
	}

	// Generate Token
	resp, tErr := t.API.RequestStandardToken(t.profile.OrgID,
		t.profile.MatchedPolicyID,
		t.token.BaseAPIID,
		t.dashboardUserAPICred,
		t.token.Expires,
		i)

	if tErr != nil {
		log.Error("Failed to generate Auth token ", tErr)
		fmt.Fprintf(w, "Auth token generation failed")
		return
	}

	if resp == nil {
		log.Error(TykAPILogTag + " --> Login failure. Request not allowed")
		fmt.Fprintf(w, "Login failed")
		return
	}

	if resp.KeyID != "" {
		log.Warning(TykAPILogTag + " --> Storing token reference")
		t.Store.SetKey(id_with_profile, resp.KeyID)
	}

	// After login, we need to redirect this user
	if t.profile.ReturnURL != "" {
		log.Info(TykAPILogTag + " --> Running auth redirect...")
		cleanURL := t.profile.ReturnURL + "#token=" + resp.KeyID
		log.Debug(TykAPILogTag+" --> URL is: ", cleanURL)
		http.Redirect(w, r, cleanURL, 301)
		return
	}

	asJson, jErr := json.Marshal(resp)
	if jErr != nil {
		log.Error(TykAPILogTag+" --> Marshalling failure: ", jErr)
		fmt.Fprintf(w, "Data Failure")
	}

	log.Info(TykAPILogTag + " --> No redirect, returning token...")
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, string(asJson))
	return
}

// CompleteIdentityAction will log a user into Tyk dashbaord or Tyk portal
func (t *TykIdentityHandler) CompleteIdentityAction(w http.ResponseWriter, r *http.Request, i interface{}, profile tap.Profile) {
	if profile.ActionType == tap.GenerateOrLoginUserProfile {
		t.CompleteIdentityActionForDashboard(w, r, i, profile)
		return
	} else if profile.ActionType == tap.GenerateOrLoginDeveloperProfile {
		t.CompleteIdentityActionForPortal(w, r, i, profile)
		return
	} else if profile.ActionType == tap.GenerateOAuthTokenForClient {
		t.CompleteIdentityActionForOAuth(w, r, i, profile)
		return
	} else if profile.ActionType == tap.GenerateTemporaryAuthToken {
		t.CompleteIdentityActionForTokenAuth(w, r, i, profile)
		return
	}
}
