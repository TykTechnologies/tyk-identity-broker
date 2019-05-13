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
)

var TykAPILogTag string = "[TYK ID HANDLER]" // log tag

type ModuleName string // To separate out target modules of the dashboard

const (
	// Enums to identify which target it being used, dashbaord or portal, they are distinct.
	SSOForDashboard ModuleName = "dashboard"
	SSOForPortal    ModuleName = "portal"
	InvalidModule   ModuleName = ""
	DefaultSSOEmail string     = "ssoSession@ssoSession.com"
)

// SSOAccessData is the data type used for speaking to the SSO endpoint in the advanced API
type SSOAccessData struct {
	ForSection   ModuleName
	OrgID        string
	EmailAddress string
	DisplayName  string
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

	logger.Error(TykAPILogTag+"Action: ", action)
	return InvalidModule, errors.New("Action does not exist")
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
			logger.Debug(TykAPILogTag + "Found Oauth configuration, loading...")
			t.oauth = OAuthSettings{}
			t.oauth.APIListenPath = oauthSettings.(map[string]interface{})["APIListenPath"].(string)
			t.oauth.RedirectURI = oauthSettings.(map[string]interface{})["RedirectURI"].(string)
			t.oauth.ResponseType = oauthSettings.(map[string]interface{})["ResponseType"].(string)
			t.oauth.ClientId = oauthSettings.(map[string]interface{})["ClientId"].(string)
			t.oauth.Secret = oauthSettings.(map[string]interface{})["Secret"].(string)
			t.oauth.BaseAPIID = oauthSettings.(map[string]interface{})["BaseAPIID"].(string)
			t.oauth.NoRedirect = oauthSettings.(map[string]interface{})["NoRedirect"].(bool)
		}

		tokenSettings, tokenOk := theseConfs["TokenAuth"]
		if tokenOk {
			if tokenSettings.(map[string]interface{})["BaseAPIID"] == nil {
				logger.Error(TykAPILogTag + " Base API is empty!")
				return errors.New("Base API cannot be empty")
			}
			t.token.BaseAPIID = tokenSettings.(map[string]interface{})["BaseAPIID"].(string)

			if tokenSettings.(map[string]interface{})["Expires"] == nil {
				logger.Warning(TykAPILogTag + " No expiry found - defaulting to 3600 seconds")
				t.token.Expires = 3600
			} else {
				t.token.Expires = int64(tokenSettings.(map[string]interface{})["Expires"].(float64))
			}

		}
	}

	return nil
}

// CreateIdentity will generate an SSO token that can be used with the tyk SSO endpoints for dash or portal.
// Identity is assumed to be a goth.User object as this is what we are stnadardiseing on.
func (t *TykIdentityHandler) CreateIdentity(i interface{}) (string, error) {

	logger.Debugf("%s  Creating identity for user: %#v", TykAPILogTag, i.(goth.User))

	thisModule, modErr := mapActionToModule(t.profile.ActionType)
	if modErr != nil {
		logger.Error(TykAPILogTag+" Failed to assign module: ", modErr)
		return "", modErr
	}

	gUser, ok := i.(goth.User)
	email := ""
	displayName := ""
	if ok {
		if t.profile.CustomEmailField != "" {
			if gUser.RawData[t.profile.CustomEmailField] != nil {
				email = gUser.RawData[t.profile.CustomEmailField].(string)
			}
		}
		if email == "" && gUser.Email != "" {
			email = gUser.Email
		} else {
			email = DefaultSSOEmail
		}

		if gUser.FirstName != "" {
			displayName = gUser.FirstName
		}
		if gUser.LastName != "" {
			if displayName != "" { //i.e. it already contains FirstName, adding space so it'll be "FirstName LastName"
				displayName += " "
			}
			displayName += gUser.LastName
		}
		if displayName == "" {
			displayName = email
		}
	}

	accessRequest := SSOAccessData{
		ForSection:   thisModule,
		OrgID:        t.profile.OrgID,
		EmailAddress: email,
		DisplayName:  displayName,
	}

	returnVal, retErr := t.API.CreateSSONonce(tyk.SSO, accessRequest)

	logger.WithField("return_value", returnVal).Debug("Returned from /admin/sso endpoint.")
	if retErr != nil {
		logger.WithField("return_value", returnVal).Error(TykAPILogTag+" API Response error: ", retErr)
		return "", retErr
	}

	asMapString := returnVal.(map[string]interface{})

	return asMapString["Meta"].(string), nil
}

// CompleteIdentityActionForDashboard handles a dashboard identity. No ise is created, only an SSO login session
func (t *TykIdentityHandler) CompleteIdentityActionForDashboard(w http.ResponseWriter, r *http.Request, i interface{}, profile tap.Profile) {
	nonce, nErr := t.CreateIdentity(i)

	if nErr != nil {
		logger.Error(TykAPILogTag+" Nonce creation failed: ", nErr)
		fmt.Fprintf(w, "Login failed")
		return
	}

	// After login, we need to redirect this user
	logger.Debug(TykAPILogTag + " --> Running redirect...")
	if profile.ReturnURL != "" {
		newURL := profile.ReturnURL + "?nonce=" + nonce
		logger.Infoln(TykAPILogTag + " --> redirecting to URL: " + newURL)
		http.Redirect(w, r, newURL, 301)
		return
	}

	logger.Error(TykAPILogTag + "No return URL found, cannot redirect. (Check why no URL redirect on the profile) ")
	fmt.Fprintf(w, "Check with your admin why there's no URI defined")
}

// CompleteIdentityActionForPortal will generate an identity for a portal user based, so it will AddOrUpdate that
// user depnding on if they exist or not and validate the login using a one-time nonce.
func (t *TykIdentityHandler) CompleteIdentityActionForPortal(w http.ResponseWriter, r *http.Request, i interface{}, profile tap.Profile) {
	// Create a nonce
	logger.Info(TykAPILogTag + " Creating nonce")
	nonce, nErr := t.CreateIdentity(i)

	if nErr != nil {
		logger.Error(TykAPILogTag+" Nonce creation failed: ", nErr)
		fmt.Fprintf(w, "Login failed")
		return
	}

	user := i.(goth.User)
	// Check if user exists
	sso_key := tap.GenerateSSOKey(user)
	logger.Debug("sso_key=", sso_key)

	thisUser, retErr, isAuthorised := t.API.GetDeveloperBySSOKey(t.dashboardUserAPICred, sso_key)
	if !isAuthorised {
		logger.WithField("returned_error", retErr).Error(TykAPILogTag + " User is unauthorized.")
		fmt.Fprintf(w, "Login failed")
		return
	}
	if retErr != nil {
		logger.WithField("returned_error", retErr).Info(TykAPILogTag + " User not found, creating new record.")

		// If not, create user
		logger.Info(TykAPILogTag + " Creating user")
		logger.WithField("user_name", user.Email).Debug(TykAPILogTag)

		newUser := tyk.PortalDeveloper{
			Email:         user.Email,
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
			logger.Error(TykAPILogTag+" failed to create user! ", createErr)
			fmt.Fprintf(w, "Login failed")
			return
		}
	} else {
		logger.Debug(TykAPILogTag+" Returned: ", thisUser)

		if thisUser.Email == "" {
			thisUser.Email = user.Email
		}
		// Set nonce value in user profile
		thisUser.Nonce = nonce
		if thisUser.Password == "" {
			thisUser.Password = uuid.NewV4().String()
		}
		updateErr := t.API.UpdateDeveloper(t.dashboardUserAPICred, thisUser)
		if updateErr != nil {
			logger.Error("Failed to update user! ", updateErr)
			fmt.Fprintf(w, "Login failed")
			return
		}
	}

	// After login, we need to redirect this user
	logger.Info(TykAPILogTag + " --> Running redirect...")
	if profile.ReturnURL != "" {
		newURL := profile.ReturnURL + "?nonce=" + nonce
		logger.Info(TykAPILogTag+" --> URL With NONCE is: ", newURL)
		http.Redirect(w, r, newURL, 301)
		return
	}

	logger.Warning(TykAPILogTag + "No return URL found, redirect failed.")
	fmt.Fprintf(w, "Success! (Have you set a return URL?)")
}

func (t *TykIdentityHandler) CompleteIdentityActionForOAuth(w http.ResponseWriter, r *http.Request, i interface{}, profile tap.Profile) {
	logger.Info(TykAPILogTag + " Starting OAuth Flow...")

	// Generate identity key match ID
	sso_key := tap.GenerateSSOKey(i.(goth.User))
	id_with_profile := t.profile.ID + "-" + sso_key
	// Check if key already exists

	value := ""
	logger.Debug("Store is: ", t.Store)
	logger.Debug("ID IS: ", id_with_profile)

	if !t.disableOneTokenPerAPI {
		fErr := t.Store.GetKey(id_with_profile, &value)
		if fErr == nil {
			// Key found
			logger.Warning(TykAPILogTag + " --> Token exists, invalidating")
			iErr, isAuthorized := t.API.InvalidateToken(t.dashboardUserAPICred, t.oauth.BaseAPIID, value)
			if iErr != nil {
				logger.WithField("isAuthorized", isAuthorized).WithField("returned-error", iErr).Error(TykAPILogTag + " ----> Token Invalidation failed.")

				//TODO: Should we return here??? the following call is against the gateway directly, so it's different credential.
				//TODO: The other action to auth token is calling the dash. why they are not the same?
				if !isAuthorized {
					logger.Error(TykAPILogTag + "Unauthorized user. Should exit.")
				}
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
		logger.Error("Failed to generate OAuth token ", oErr)
		fmt.Fprintf(w, "OAuth token generation failed")
		return
	}

	if resp == nil {
		logger.Error(TykAPILogTag + " --> Login failure. Request not allowed")
		fmt.Fprintf(w, "Login failed")
		return
	}

	if resp.AccessToken != "" {
		logger.Warning(TykAPILogTag + " --> Storing token reference")
		t.Store.SetKey(id_with_profile, resp.AccessToken)
	}

	if t.oauth.NoRedirect {
		asJson, jErr := json.Marshal(resp)
		if jErr != nil {
			logger.Error(TykAPILogTag+" --> Marshalling failure: ", jErr)
			fmt.Fprintf(w, "Data Failure")
		}

		logger.Info(TykAPILogTag + " --> No redirect, returning token...")
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, string(asJson))
		return
	}

	// After login, we need to redirect this user
	logger.Info(TykAPILogTag + " --> Running oauth redirect...")
	if resp.RedirectTo != "" {
		logger.Debug(TykAPILogTag+" --> URL is: ", resp.RedirectTo)
		http.Redirect(w, r, resp.RedirectTo, 301)
		return
	}
}

func (t *TykIdentityHandler) CompleteIdentityActionForTokenAuth(w http.ResponseWriter, r *http.Request, i interface{}, profile tap.Profile) {
	logger.Info(TykAPILogTag + " Starting Token Flow...")

	// Generate identity key match ID
	sso_key := tap.GenerateSSOKey(i.(goth.User))
	id_with_profile := t.profile.ID + "-" + sso_key
	// Check if key already exists

	value := ""
	logger.Debug("Store is: ", t.Store)
	logger.Debug("ID IS: ", id_with_profile)

	if !t.disableOneTokenPerAPI {
		fErr := t.Store.GetKey(id_with_profile, &value)
		if fErr == nil {
			// Key found
			logger.Warning(TykAPILogTag + " --> Token exists, invalidating")
			iErr, isAuthorized := t.API.InvalidateToken(t.dashboardUserAPICred, t.token.BaseAPIID, value)
			if iErr != nil {
				logger.Error(TykAPILogTag+" ----> Token Invalidation failed: ", iErr)

				logger.WithField("isAuthorized", isAuthorized).WithField("returned-error", iErr).Error(TykAPILogTag + " ----> Token Invalidation failed.")

				//TODO: Should we return here??? the following call is against the dashboard directly, so it will fail again.
				//TODO: The other action to auth token is calling the gateway. why they are not the same?
				if !isAuthorized {
					logger.Error(TykAPILogTag + "Unauthorized user. Should exit.")
					fmt.Fprintf(w, "Auth token generation failed due to invalid user credentials.")
					return
				}
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
		logger.Error("Failed to generate Auth token ", tErr)
		fmt.Fprintf(w, "Auth token generation failed")
		return
	}

	if resp == nil {
		logger.Error(TykAPILogTag + " --> Login failure. Request not allowed")
		fmt.Fprintf(w, "Login failed")
		return
	}

	if resp.KeyID != "" {
		logger.Warning(TykAPILogTag + " --> Storing token reference")
		t.Store.SetKey(id_with_profile, resp.KeyID)
	}

	// After login, we need to redirect this user
	if t.profile.ReturnURL != "" {
		logger.Info(TykAPILogTag + " --> Running auth redirect...")
		cleanURL := t.profile.ReturnURL + "#token=" + resp.KeyID
		logger.Debug(TykAPILogTag+" --> URL is: ", cleanURL)
		http.Redirect(w, r, cleanURL, 301)
		return
	}

	asJson, jErr := json.Marshal(resp)
	if jErr != nil {
		logger.Error(TykAPILogTag+" --> Marshalling failure: ", jErr)
		fmt.Fprintf(w, "Data Failure")
	}

	logger.Info(TykAPILogTag + " --> No redirect, returning token...")
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
