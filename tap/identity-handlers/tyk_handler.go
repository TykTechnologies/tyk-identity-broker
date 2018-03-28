/* package identityHandlers provides a collection of handlers for target systems,
these handlers create accounts and sso tokens */
package identityHandlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/markbates/goth"
	"github.com/satori/go.uuid"

	"github.com/TykTechnologies/tyk-identity-broker/tap"
	"github.com/TykTechnologies/tyk-identity-broker/tyk-api"

	"github.com/Sirupsen/logrus"
)

var TykAPILogTag string = "[TYK ID HANDLER]" // log tag

type ModuleName string // To separate out target modules of the dashboard

const (
	// Enums to identify which target it being used, dashbaord or portal, they are distinct.
	SSOForDashboard ModuleName = "dashboard"
	SSOForPortal    ModuleName = "portal"
	InvalidModule   ModuleName = ""
)

// SSOAccessData is the data type used for speaking to the SSO endpoint in the advanced API
type SSOAccessData struct {
	ForSection   ModuleName
	OrgID        string
	EmailAddress string
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
	NoRedirect    bool
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

	log.Error(TykAPILogTag+" Action: ", action)
	return InvalidModule, errors.New("Action does not exist")
}

// initialise th Tyk handler, the Tyk handler *requires* initialisation with the TykAPI handler global set
// up in main
func (t *TykIdentityHandler) Init(conf interface{}) error {
	log.Level = logrus.DebugLevel
	var ok bool
	if t.profile, ok = conf.(tap.Profile); !ok {
		return errors.New("No profile found ")
	}

	if t.profile.IdentityHandlerConfig == nil {
		return errors.New("No IdentityHandlerConfig found in profile.")
	}

	var theseConfs map[string]interface{}
	if theseConfs, ok = t.profile.IdentityHandlerConfig.(map[string]interface{}); !ok {
		msg := fmt.Sprintf("Profile '%s': `IdentityHandlerConfig` should be a map.", t.profile.ID)
		return errors.New(msg)
	}

	if theseConfs["DashboardCredential"] != nil {
		if dashboardUserAPICred, ok := theseConfs["DashboardCredential"].(string); ok {
			t.dashboardUserAPICred = dashboardUserAPICred
		} else {
			msg := fmt.Sprintf("Profile '%s': `dashboardUserAPICred` should be bool.", t.profile.ID)
			return errors.New(msg)
		}
	} else {
		log.Warningf("%s Profile '%s': Not recommended for `DashboardCredential` to be nil",
			TykAPILogTag, t.profile.ID)
	}

	if theseConfs["DisableOneTokenPerAPI"] != nil {
		if disableOneTokenPerAPI, ok := theseConfs["DisableOneTokenPerAPI"].(bool); ok {
			t.disableOneTokenPerAPI = disableOneTokenPerAPI
		} else {
			msg := fmt.Sprintf("Profile '%s': `DisableOneTokenPerAPI` should be boolean.", t.profile.ID)
			return errors.New(msg)
		}
	} else {
		log.Warningf("%s Profile '%s': Not recommended for `DisableOneTokenPerAPI` to be nil",
			TykAPILogTag, t.profile.ID)
	}

	if err := t.getOauthSettings(theseConfs) ; err == nil {
		return err
	}

	if err := t.getTokenAuthSettings(theseConfs) ; err == nil {
		return err
	}

	return nil
}

func (t *TykIdentityHandler) getOauthSettings (theseConfs map[string]interface{}) error {
	if oauthSettings, ok := theseConfs["OAuth"] ; ok {
		log.Debug(TykAPILogTag + " Found Oauth configuration, loading...")

		if oauthSettingsMap, ok := oauthSettings.(map[string]interface{}); ok {
			t.oauth = OAuthSettings{}
			//Todo: check all the 'ok's
			t.oauth.APIListenPath, ok = oauthSettingsMap["APIListenPath"].(string)
			t.oauth.RedirectURI, ok = oauthSettingsMap["RedirectURI"].(string)
			t.oauth.ResponseType, ok = oauthSettingsMap["ResponseType"].(string)
			t.oauth.ClientId, ok = oauthSettingsMap["ClientId"].(string)
			t.oauth.Secret, ok = oauthSettingsMap["Secret"].(string)
			t.oauth.BaseAPIID, ok = oauthSettingsMap["BaseAPIID"].(string)
			t.oauth.NoRedirect, ok = oauthSettingsMap["NoRedirect"].(bool)

			return nil
		}
		msg := fmt.Sprintf("Profile '%s': `OAuth` should be a map.", t.profile.ID)
		log.Errorf("%s %s", TykAPILogTag, msg)
		return errors.New(msg)
	}
	return nil
}

func (t *TykIdentityHandler) getTokenAuthSettings (theseConfs map[string]interface{}) error {
	if theseConfs["TokenAuth"] == nil {
		msg := fmt.Sprintf("Profile '%s': `TokenAuth` should not be nil.", t.profile.ID)
		log.Debug(TykAPILogTag + msg)
		return errors.New(msg)
	}

	var ok bool = false
	var tokenSettingsMap map[string]interface{}
	if tokenSettingsMap, ok = theseConfs["TokenAuth"].(map[string]interface{}); !ok {
		msg := fmt.Sprintf("Profile '%s': `TokenAuth` should be a map.", t.profile.ID)
		return errors.New(msg)
	}

	if tokenSettingsMap["BaseAPIID"] == nil {
		log.Error(TykAPILogTag + " Base API is empty!")
		return errors.New(" Base API cannot be empty")
	}

	if t.token.BaseAPIID, ok = tokenSettingsMap["BaseAPIID"].(string); !ok {
			msg := fmt.Sprintf("Profile '%s': field `BaseAPIID` should be string .", t.profile.ID)
			return errors.New(msg)
	}

	if tokenSettingsMap["Expires"] == nil {
		log.Warningf( " %s Profile '%s': No 'Expires' field - defaulting to 3600 seconds", TykAPILogTag, t.profile.ID)
		t.token.Expires = 3600
	} else {
		if expire, ok := tokenSettingsMap["Expires"].(float64); ok {
			t.token.Expires = int64(expire)
		} else {
			msg := fmt.Sprintf("Profile '%s': `Expires` field should be float", t.profile.ID)
			return errors.New(msg)
		}
	}
	return nil
}

// CreateIdentity will generate an SSO token that can be used with the tyk SSO endpoints for dash or portal.
// Identity is assumed to be a goth.User object as this is what we are stnadardiseing on.
func (t *TykIdentityHandler) CreateIdentity(i interface{}) (string, error) {
	log.Debugf("%s  Creating identity for user: %#v", TykAPILogTag, i.(goth.User))

	thisModule, modErr := mapActionToModule(t.profile.ActionType)
	if modErr != nil {
		log.Error(TykAPILogTag+" Failed to assign module: ", modErr)
		return "", modErr
	}

	accessRequest := SSOAccessData{
		ForSection: thisModule,
		OrgID:      t.profile.OrgID,
		EmailAddress: "ssoSession@ssoSession.com",
	}

	returnVal, retErr := t.API.CreateSSONonce(tyk.SSO, accessRequest)

	if retErr != nil {
		log.Error(TykAPILogTag+" API Response error: ", retErr)
		return "", retErr
	}

	asMapString := returnVal.(map[string]interface{})

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
		log.Infoln(TykAPILogTag + " --> redirecting to URL: " + newURL)
		http.Redirect(w, r, newURL, 301)
		return
	}

	log.Error(TykAPILogTag + "No return URL found, cannot redirect. (Check why no URL redirect on the profile) ")
	fmt.Fprintf(w, "Check with your admin why there's no URI defined")
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
			Password:      uuid.NewV4().String(),
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
		thisUser.Email = sso_key
		if thisUser.Password == "" {
			thisUser.Password = uuid.NewV4().String()
		}
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

	if t.oauth.NoRedirect {
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
