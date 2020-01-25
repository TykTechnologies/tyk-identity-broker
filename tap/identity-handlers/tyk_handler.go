/* package identityHandlers provides a collection of handlers for target systems,
these handlers create accounts and sso tokens */
package identityHandlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/markbates/goth"
	uuid "github.com/satori/go.uuid"

	"github.com/TykTechnologies/tyk-identity-broker/tap"
	tyk "github.com/TykTechnologies/tyk-identity-broker/tyk-api"
)

var tykHandlerLogger = log.WithField("prefix", "TYK ID HANDLER")

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
	GroupID      string
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

	tykHandlerLogger.Error("Action: ", action)
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
			tykHandlerLogger.Debug("Found Oauth configuration, loading...")
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
				tykHandlerLogger.Error("Base API is empty!")
				return errors.New("Base API cannot be empty")
			}
			t.token.BaseAPIID = tokenSettings.(map[string]interface{})["BaseAPIID"].(string)

			if tokenSettings.(map[string]interface{})["Expires"] == nil {
				tykHandlerLogger.Warning("No expiry found - defaulting to 3600 seconds")
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

	tykHandlerLogger.Debugf("Creating identity for user: %#v", i.(goth.User))

	thisModule, modErr := mapActionToModule(t.profile.ActionType)
	if modErr != nil {
		tykHandlerLogger.Error("Failed to assign module: ", modErr)
		return "", modErr
	}

	gUser, ok := i.(goth.User)
	email := ""
	displayName := ""
	groupID := ""
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

		groupID = t.profile.DefaultUserGroupID

		if t.profile.CustomUserGroupField != "" {
			groups := ""
			if gUser.RawData[t.profile.CustomUserGroupField] != nil {
				groups = gUser.RawData[t.profile.CustomUserGroupField].(string)
			}

			for _, group := range strings.Split(groups, " ") {
				if gid, ok := t.profile.UserGroupMapping[group]; ok {
					groupID = gid
				}
			}
		}
	}

	accessRequest := SSOAccessData{
		ForSection:   thisModule,
		OrgID:        t.profile.OrgID,
		EmailAddress: email,
		DisplayName:  displayName,
		GroupID:      groupID,
	}

	returnVal, ssoEndpoint, retErr := t.API.CreateSSONonce(t.dashboardUserAPICred, accessRequest)

	tykHandlerLogger.WithField("return_value", returnVal).Debugf("Returned from %s endpoint", ssoEndpoint)
	if retErr != nil {
		tykHandlerLogger.WithField("return_value", returnVal).Error("API Response error: ", retErr)
		return "", retErr
	}

	asMapString := returnVal.(map[string]interface{})

	return asMapString["Meta"].(string), nil
}

// CompleteIdentityActionForDashboard handles a dashboard identity. No ise is created, only an SSO login session
func (t *TykIdentityHandler) CompleteIdentityActionForDashboard(w http.ResponseWriter, r *http.Request, i interface{}, profile tap.Profile) {
	nonce, nErr := t.CreateIdentity(i)

	if nErr != nil {
		tykHandlerLogger.WithField("error", nErr).Error("Nonce creation failed")
		fmt.Fprintf(w, "Login failed")
		return
	}

	// After login, we need to redirect this user
	tykHandlerLogger.Debug("--> Running redirect...")
	if profile.ReturnURL != "" {
		newURL := profile.ReturnURL + "?nonce=" + nonce
		tykHandlerLogger.Debugln("--> redirecting to URL: " + newURL)
		http.Redirect(w, r, newURL, 301)
		return
	}

	tykHandlerLogger.Error("No return URL found, cannot redirect. (Check why no URL redirect on the profile) ")
	fmt.Fprintf(w, "Check with your admin why there's no URI defined")
}

// CompleteIdentityActionForPortal will generate an identity for a portal user based, so it will AddOrUpdate that
// user depnding on if they exist or not and validate the login using a one-time nonce.
func (t *TykIdentityHandler) CompleteIdentityActionForPortal(w http.ResponseWriter, r *http.Request, i interface{}, profile tap.Profile) {
	// Create a nonce
	tykHandlerLogger.Info("Creating nonce")
	nonce, nErr := t.CreateIdentity(i)

	if nErr != nil {
		tykHandlerLogger.Error("Nonce creation failed: ", nErr)
		fmt.Fprintf(w, "Login failed")
		return
	}

	user := i.(goth.User)
	if t.profile.CustomUserIDField != "" {
		if user.RawData[t.profile.CustomUserIDField] != nil {
			user.UserID = user.RawData[t.profile.CustomUserIDField].(string)
		}
	}
	// Check if user exists
	sso_key := tap.GenerateSSOKey(user)
	tykHandlerLogger.Debug("sso_key = ", sso_key)

	thisUser, retErr, isAuthorised := t.API.GetDeveloperBySSOKey(t.dashboardUserAPICred, sso_key)
	if !isAuthorised {
		tykHandlerLogger.WithField("returned_error", retErr).Error("User is unauthorized.")
		fmt.Fprintf(w, "Login failed")
		return
	}
	if retErr != nil {
		tykHandlerLogger.WithField("returned_error", retErr).Info("User not found, creating new record.")

		// If not, create user
		tykHandlerLogger.Info("Creating user")
		tykHandlerLogger.WithField("user_name", user.Email).Debug()

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
			tykHandlerLogger.WithField("error", createErr).Error("failed to create user!")
			fmt.Fprintf(w, "Login failed")
			return
		}
	} else {
		tykHandlerLogger.Debug("Returned: ", thisUser)

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
			tykHandlerLogger.WithField("error", updateErr).Error("Failed to update user!")
			fmt.Fprintf(w, "Login failed")
			return
		}
	}

	// After login, we need to redirect this user
	tykHandlerLogger.Info("--> Running redirect...")
	if profile.ReturnURL != "" {
		newURL := profile.ReturnURL + "?nonce=" + nonce
		tykHandlerLogger.Info("--> URL With NONCE is: ", newURL)
		http.Redirect(w, r, newURL, 301)
		return
	}

	tykHandlerLogger.Warning("No return URL found, redirect failed.")
	fmt.Fprintf(w, "Success! (Have you set a return URL?)")
}

func (t *TykIdentityHandler) CompleteIdentityActionForOAuth(w http.ResponseWriter, r *http.Request, i interface{}, profile tap.Profile) {
	tykHandlerLogger.Info("Starting OAuth Flow...")

	// Generate identity key match ID
	sso_key := tap.GenerateSSOKey(i.(goth.User))
	id_with_profile := t.profile.ID + "-" + sso_key
	// Check if key already exists

	value := ""
	tykHandlerLogger.Debug("Store is: ", t.Store)
	tykHandlerLogger.Debug("ID IS: ", id_with_profile)

	if !t.disableOneTokenPerAPI {
		fErr := t.Store.GetKey(id_with_profile, &value)
		if fErr == nil {
			// Key found
			tykHandlerLogger.Warning("--> Token exists, invalidating")
			iErr, isAuthorized := t.API.InvalidateToken(t.dashboardUserAPICred, t.oauth.BaseAPIID, value)
			if iErr != nil {
				tykHandlerLogger.WithField("isAuthorized", isAuthorized).WithField("returned-error", iErr).Error("----> Token Invalidation failed.")

				//TODO: Should we return here??? the following call is against the gateway directly, so it's different credential.
				//TODO: The other action to auth token is calling the dash. why they are not the same?
				if !isAuthorized {
					tykHandlerLogger.Error("Unauthorized user. Should exit.")
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
		tykHandlerLogger.WithField("error", oErr).Error("Failed to generate OAuth token")
		fmt.Fprintf(w, "OAuth token generation failed")
		return
	}

	if resp == nil {
		tykHandlerLogger.Error("--> Login failure. Request not allowed")
		fmt.Fprintf(w, "Login failed")
		return
	}

	if resp.AccessToken != "" {
		tykHandlerLogger.Warning("--> Storing token reference")
		t.Store.SetKey(id_with_profile, resp.AccessToken)
	}

	if t.oauth.NoRedirect {
		asJson, jErr := json.Marshal(resp)
		if jErr != nil {
			tykHandlerLogger.WithField("error", jErr).Error("--> Marshalling failure")
			fmt.Fprintf(w, "Data Failure")
		}

		tykHandlerLogger.Info("--> No redirect, returning token...")
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, string(asJson))
		return
	}

	// After login, we need to redirect this user
	tykHandlerLogger.Info("--> Running oauth redirect...")
	if resp.RedirectTo != "" {
		tykHandlerLogger.Debug("--> URL is: ", resp.RedirectTo)
		http.Redirect(w, r, resp.RedirectTo, 301)
		return
	}
}

func (t *TykIdentityHandler) CompleteIdentityActionForTokenAuth(w http.ResponseWriter, r *http.Request, i interface{}, profile tap.Profile) {
	tykHandlerLogger.Info("Starting Token Flow...")

	// Generate identity key match ID
	sso_key := tap.GenerateSSOKey(i.(goth.User))
	id_with_profile := t.profile.ID + "-" + sso_key
	// Check if key already exists

	value := ""
	tykHandlerLogger.Debug("Store is: ", t.Store)
	tykHandlerLogger.Debug("ID IS: ", id_with_profile)

	if !t.disableOneTokenPerAPI {
		fErr := t.Store.GetKey(id_with_profile, &value)
		if fErr == nil {
			// Key found
			tykHandlerLogger.Warning("--> Token exists, invalidating")
			iErr, isAuthorized := t.API.InvalidateToken(t.dashboardUserAPICred, t.token.BaseAPIID, value)
			if iErr != nil {
				tykHandlerLogger.WithField("isAuthorized", isAuthorized).WithField("returned-error", iErr).Error(" ----> Token Invalidation failed.")

				//TODO: Should we return here??? the following call is against the dashboard directly, so it will fail again.
				//TODO: The other action to auth token is calling the gateway. why they are not the same?
				if !isAuthorized {
					tykHandlerLogger.Error("Unauthorized user. Should exit.")
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
		tykHandlerLogger.WithField("error", tErr).Error("Failed to generate Auth token")
		fmt.Fprintf(w, "Auth token generation failed")
		return
	}

	if resp == nil {
		tykHandlerLogger.Error("--> Login failure. Request not allowed")
		fmt.Fprintf(w, "Login failed")
		return
	}

	if resp.KeyID != "" {
		tykHandlerLogger.Warning("--> Storing token reference")
		t.Store.SetKey(id_with_profile, resp.KeyID)
	}

	// After login, we need to redirect this user
	if t.profile.ReturnURL != "" {
		tykHandlerLogger.Info("--> Running auth redirect...")
		cleanURL := t.profile.ReturnURL + "#token=" + resp.KeyID
		tykHandlerLogger.Debug("--> URL is: ", cleanURL)
		http.Redirect(w, r, cleanURL, 301)
		return
	}

	asJson, jErr := json.Marshal(resp)
	if jErr != nil {
		tykHandlerLogger.WithField("error", jErr).Error("--> Marshalling failure")
		fmt.Fprintf(w, "Data Failure")
	}

	tykHandlerLogger.Info("--> No redirect, returning token...")
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
