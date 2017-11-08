package tyk

import (
	"bytes"
	"encoding/json"
	"errors"
	log "github.com/Sirupsen/logrus"
	"github.com/markbates/goth"
	"io"
	"io/ioutil"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Endpoint string   // A type for endpoints
type TykAPIName string // A type for Tyk API names (e.g. dashboard, gateway)

// EndpointConfig is a Configuration for an API Endpoint of one of the Tyk APIs
type EndpointConfig struct {
	Endpoint    string
	Port        string
	AdminSecret string
}

type OAuthResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
	RedirectTo  string `json:"redirect_to"`
	TokenType   string `json:"token_type"`
}

type TokenResponse struct {
	KeyID string `json:"key_id"`
}

// TykAPI is the main object (and configuration) of the Tyk API wrapper
type TykAPI struct {
	GatewayConfig   EndpointConfig
	DashboardConfig EndpointConfig
}

// PortalDeveloper represents a portal developer
type PortalDeveloper struct {
	Id            bson.ObjectId     `bson:"_id,omitempty" json:"id"`
	Email         string            `bson:"email" json:"email"`
	Password      string            `bson:"password" json:"password"`
	DateCreated   time.Time         `bson:"date_created" json:"date_created"`
	InActive      bool              `bson:"inactive" json:"inactive"`
	OrgId         string            `bson:"org_id" json:"org_id"`
	ApiKeys       map[string]string `bson:"api_keys" json:"api_keys"`
	Subscriptions map[string]string `bson:"subscriptions" json:"subscriptions"`
	Fields        map[string]string `bson:"fields" json:"fields"`
	Nonce         string            `bson:"nonce" json:"nonce"`
	SSOKey        string            `bson:"sso_key" json:"sso_key"`
}

// HashType is an encryption method for basic auth keys
type HashType string

// AccessSpecs define what URLS a user has access to an what methods are enabled
type AccessSpec struct {
	URL     string   `json:"url"`
	Methods []string `json:"methods"`
}

// AccessDefinition defines which versions of an API a key has access to
type AccessDefinition struct {
	APIName     string       `json:"api_name"`
	APIID       string       `json:"api_id"`
	Versions    []string     `json:"versions"`
	AllowedURLs []AccessSpec `bson:"allowed_urls"  json:"allowed_urls"` // mapped string MUST be a valid regex
}

// SessionState objects represent a current API session, mainly used for rate limiting.
type SessionState struct {
	LastCheck        int64                       `json:"last_check"`
	Allowance        float64                     `json:"allowance"`
	Rate             float64                     `json:"rate"`
	Per              float64                     `json:"per"`
	Expires          int64                       `json:"expires"`
	QuotaMax         int64                       `json:"quota_max"`
	QuotaRenews      int64                       `json:"quota_renews"`
	QuotaRemaining   int64                       `json:"quota_remaining"`
	QuotaRenewalRate int64                       `json:"quota_renewal_rate"`
	AccessRights     map[string]AccessDefinition `json:"access_rights"`
	OrgID            string                      `json:"org_id"`
	OauthClientID    string                      `json:"oauth_client_id"`
	OauthKeys        map[string]string           `json:"oauth_keys"`
	BasicAuthData    struct {
		Password string   `json:"password"`
		Hash     HashType `json:"hash_type"`
	} `json:"basic_auth_data"`
	JWTData struct {
		Secret string `json:"secret"`
	} `json:"jwt_data"`
	HMACEnabled   bool   `json:"hmac_enabled"`
	HmacSecret    string `json:"hmac_string"`
	IsInactive    bool   `json:"is_inactive"`
	ApplyPolicyID string `json:"apply_policy_id"`
	DataExpires   int64  `json:"data_expires"`
	Monitor       struct {
		TriggerLimits []float64 `json:"trigger_limits"`
	} `json:"monitor"`
	MetaData interface{} `json:"meta_data"`
	Tags     []string    `json:"tags"`
}

const (
	// Main endpoints used in this wrapper
	PORTAL_DEVS     Endpoint = "/api/portal/developers/email"
	PORTAL_DEVS_SSO Endpoint = "/api/portal/developers/ssokey"
	PORTAL_DEV      Endpoint = "/api/portal/developers"
	SSO             Endpoint = "/admin/sso"
	OAUTH_AUTHORIZE Endpoint = "tyk/oauth/authorize-client/"
	TOKENS          Endpoint = "/api/apis/{APIID}/keys"
	STANDARD_TOKENS Endpoint = "/api/keys"

	// Main APis used in this wrapper
	GATEWAY    TykAPIName = "gateway"
	DASH       TykAPIName = "dash"
	DASH_SUPER TykAPIName = "dash_super"

	HASH_PlainText HashType = ""
	HASH_BCrypt    HashType = "bcrypt"
)

// DispatchDashboard dispatches a request to the dashboard API and handles the response
func (t *TykAPI) DispatchDashboard(target Endpoint, method string, usercode string, body io.Reader) ([]byte, error) {
	preparedEndpoint := t.DashboardConfig.Endpoint + ":" + t.DashboardConfig.Port + string(target)

	log.Debug("Calling: ", preparedEndpoint)
	newRequest, err := http.NewRequest(method, preparedEndpoint, body)
	if err != nil {
		log.Error("Failed to create request")
		log.Error(err)
	}

	newRequest.Header.Add("authorization", usercode)
	c := &http.Client{}
	response, reqErr := c.Do(newRequest)

	if reqErr != nil {
		return []byte{}, reqErr
	}

	retBody, bErr := t.readBody(response)
	if bErr != nil {
		return []byte{}, bErr
	}

	log.Debug("GOT:", string(retBody))

	if response.StatusCode > 201 {
		log.Warning("Response code was: ", response.StatusCode)
		log.Warning("GOT:", string(retBody))
		return retBody, errors.New("Response code was not 200!")
	}

	return retBody, nil
}

func (t *TykAPI) readBody(response *http.Response) ([]byte, error) {
	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return []byte(""), err
	}

	return contents, nil

}

// DispatchDashboardSuper will dispatch a request to the dashbaord super-user API (admin)
func (t *TykAPI) DispatchDashboardSuper(target Endpoint, method string, body io.Reader) ([]byte, error) {
	preparedEndpoint := t.DashboardConfig.Endpoint + ":" + t.DashboardConfig.Port + string(target)

	log.Debug("Calling: ", preparedEndpoint)
	newRequest, err := http.NewRequest(method, preparedEndpoint, body)
	if err != nil {
		log.Error("Failed to create request")
		log.Error(err)
	}

	newRequest.Header.Add("admin-auth", t.DashboardConfig.AdminSecret)
	c := &http.Client{}
	response, reqErr := c.Do(newRequest)

	if reqErr != nil {
		return []byte{}, reqErr
	}

	retBody, bErr := t.readBody(response)
	if bErr != nil {
		return []byte{}, bErr
	}

	if response.StatusCode > 201 {
		log.Warning("Response code was: ", response.StatusCode)
		log.Warning("Returned: ", string(retBody))
		return retBody, errors.New("Response code was not 200!")
	}

	return retBody, nil
}

// DispatchGateway will dispatch a request to the gateway API
func (t *TykAPI) DispatchGateway(target Endpoint, method string, body io.Reader, ctype string) ([]byte, error) {
	preparedEndpoint := t.GatewayConfig.Endpoint + ":" + t.GatewayConfig.Port + string(target)

	log.Debug("Calling: ", preparedEndpoint)
	newRequest, err := http.NewRequest(method, preparedEndpoint, body)
	if err != nil {
		log.Error("Failed to create request")
		log.Error(err)
	}

	if ctype == "" {
		ctype = "application/json"
	}

	newRequest.Header.Add("x-tyk-authorization", t.GatewayConfig.AdminSecret)
	newRequest.Header.Add("content-type", ctype)
	c := &http.Client{}
	response, reqErr := c.Do(newRequest)

	if reqErr != nil {
		return []byte{}, reqErr
	}

	retBody, bErr := t.readBody(response)
	if bErr != nil {
		return []byte{}, bErr
	}

	if response.StatusCode > 201 {
		log.Warning("Response code was: ", response.StatusCode)
		return retBody, errors.New("Response code was not 200!")
	}

	log.Debug("API Response: ", string(retBody))

	return retBody, nil
}

// Dcode will unmarshal a request body, a bit redundant tbh
func (t *TykAPI) Decode(raw []byte, retVal interface{}) error {
	decErr := json.Unmarshal(raw, &retVal)
	return decErr
}

// DispatchAndDecode will select the API to target, dispatch the request, then decode ther esponse to return to the caller
func (t *TykAPI) DispatchAndDecode(target Endpoint, method string, APIName TykAPIName, retVal interface{}, creds string, body io.Reader, ctype string) error {
	var retBytes []byte
	var dispatchErr error

	switch APIName {
	case GATEWAY:
		retBytes, dispatchErr = t.DispatchGateway(target, method, body, ctype)
	case DASH:
		retBytes, dispatchErr = t.DispatchDashboard(target, method, creds, body)
	case DASH_SUPER:
		retBytes, dispatchErr = t.DispatchDashboardSuper(target, method, body)
	default:
		return errors.New("APIName must be one of GATEWAY, DASH or DASH_SUPER")
	}

	if dispatchErr != nil {
		return dispatchErr
	}

	t.Decode(retBytes, retVal)
	return nil
}

// CreateSSONonce will generate a single-signon nonce for the relevant part of the Tyk service (dashbaord or portal),
// nonces are single-use and expire after 60 seconds to prevent hijacking, they are only available during successful
// requests by redirecting the user. It is ecommended that SSL is used throughout
func (t *TykAPI) CreateSSONonce(target Endpoint, data interface{}) (interface{}, error) {
	target = Endpoint(strings.Join([]string{string(target)}, "/"))
	SSODataJSON, err := json.Marshal(data)

	if err != nil {
		return nil, err
	}

	var returnVal interface{}
	body := bytes.NewBuffer(SSODataJSON)
	dErr := t.DispatchAndDecode(Endpoint(target), "POST", DASH_SUPER, &returnVal, "", body, "")

	return returnVal, dErr
}

// GetDeveloper will retrieve a deverloper from the Advanced API using their Email address
func (t *TykAPI) GetDeveloper(UserCred string, DeveloperEmail string) (PortalDeveloper, error) {
	asStr := url.QueryEscape(DeveloperEmail)
	target := strings.Join([]string{string(PORTAL_DEVS), asStr}, "/")

	retUser := PortalDeveloper{}

	dErr := t.DispatchAndDecode(Endpoint(target), "GET", DASH, &retUser, UserCred, nil, "")

	return retUser, dErr
}

// GetDeveloperBySSOKey will retrieve a deverloper from the Advanced API using their SSO Key address
func (t *TykAPI) GetDeveloperBySSOKey(UserCred string, DeveloperEmail string) (PortalDeveloper, error) {
	asStr := url.QueryEscape(DeveloperEmail)
	target := strings.Join([]string{string(PORTAL_DEVS_SSO), asStr}, "/")

	retUser := PortalDeveloper{}

	dErr := t.DispatchAndDecode(Endpoint(target), "GET", DASH, &retUser, UserCred, nil, "")

	return retUser, dErr
}

// UpdateDeveloper will update a developer object using the advanced API
func (t *TykAPI) UpdateDeveloper(UserCred string, dev PortalDeveloper) error {
	target := strings.Join([]string{string(PORTAL_DEV), dev.Id.Hex()}, "/")

	retData := map[string]interface{}{}
	data, err := json.Marshal(dev)
	body := bytes.NewBuffer(data)

	if err != nil {
		return err
	}

	dErr := t.DispatchAndDecode(Endpoint(target), "PUT", DASH, &retData, UserCred, body, "")

	return dErr
}

// CreateDeveloper will create a developer using the advanced API
func (t *TykAPI) CreateDeveloper(UserCred string, dev PortalDeveloper) error {
	target := strings.Join([]string{string(PORTAL_DEV)}, "/")

	retData := map[string]interface{}{}
	data, err := json.Marshal(dev)
	body := bytes.NewBuffer(data)

	if err != nil {
		return err
	}

	dErr := t.DispatchAndDecode(Endpoint(target), "POST", DASH, &retData, UserCred, body, "")
	log.Debug("Returned: ", retData)

	return dErr
}

type OAuthMethod string

var Access OAuthMethod = "AccessToken"

func generateBasicTykSesion(baseAPIID, baseVersion, policyID, orgID string) SessionState {
	// Create a generic access token withour policy
	basicSessionState := SessionState{
		Allowance:        1,
		Rate:             1,
		Per:              1,
		Expires:          -1,
		QuotaMax:         1,
		QuotaRenews:      60,
		QuotaRemaining:   1,
		QuotaRenewalRate: 1,
		AccessRights:     map[string]AccessDefinition{},
		OrgID:            orgID,
		ApplyPolicyID:    policyID,
		MetaData:         map[string]interface{}{"Origin": "TAP"},
		Tags:             []string{"TykOrigin-TAP"},
	}

	accessEntry := AccessDefinition{
		APIName:  "Base",
		APIID:    baseAPIID,
		Versions: []string{baseVersion},
	}
	basicSessionState.AccessRights[baseAPIID] = accessEntry

	return basicSessionState
}

func (t *TykAPI) RequestOAuthToken(APIlistenPath, redirect_uri, responseType, clientId, secret, orgID, policyID, BaseAPIID string, userInfo interface{}) (*OAuthResponse, error) {
	// Create a generic access token withour policy
	basicSessionState := generateBasicTykSesion(BaseAPIID, "Default", policyID, orgID)
	basicSessionState.OauthClientID = clientId
	basicSessionState.MetaData.(map[string]interface{})["AuthProviderUserID"] = userInfo.(goth.User).UserID
	basicSessionState.MetaData.(map[string]interface{})["AuthProviderSource"] = userInfo.(goth.User).Provider
	basicSessionState.MetaData.(map[string]interface{})["AccessToken"] = userInfo.(goth.User).AccessToken
	basicSessionState.MetaData.(map[string]interface{})["AccessTokenSecret"] = userInfo.(goth.User).AccessTokenSecret

	/*

		Can be extracted in Global header settings as:

		X-Origin-Tyk: $tyk_meta.Origin
		X-Tyk-TAP-AccessToken: $tyk_meta.AccessToken
		X-Tyk-TAP-ID: $tyk_meta.AuthProviderUserID
		X-Tyk-TAP-Provider: $tyk_meta.AuthProviderSource

	*/

	keyDataJSON, err := json.Marshal(basicSessionState)

	if err != nil {
		return nil, err
	}

	if clientId == "" {
		return nil, errors.New("Requires client ID")
	}

	// Make the Auth request
	response := &OAuthResponse{}
	target := "/" + strings.Join([]string{APIlistenPath, string(OAUTH_AUTHORIZE)}, "/")
	data := "response_type=" + responseType
	data += "&client_id=" + clientId
	data += "&redirect_uri=" + redirect_uri
	data += "&key_rules=" + url.QueryEscape(string(keyDataJSON))

	log.Debug("Request data sent: ", data)

	body := bytes.NewBuffer([]byte(data))
	dErr := t.DispatchAndDecode(Endpoint(target), "POST", GATEWAY, response, "", body, "application/x-www-form-urlencoded")

	log.Debug("Returned: ", response)

	if dErr != nil {
		return nil, err
	}

	return response, nil
}

func (t *TykAPI) RequestStandardToken(orgID, policyID, BaseAPIID, UserCred string, expires int64, userInfo interface{}) (*TokenResponse, error) {
	// Create a generic access token withour policy
	basicSessionState := generateBasicTykSesion(BaseAPIID, "Default", policyID, orgID)
	basicSessionState.MetaData.(map[string]interface{})["AuthProviderUserID"] = userInfo.(goth.User).UserID
	basicSessionState.MetaData.(map[string]interface{})["AuthProviderSource"] = userInfo.(goth.User).Provider
	basicSessionState.MetaData.(map[string]interface{})["AccessToken"] = userInfo.(goth.User).AccessToken
	basicSessionState.MetaData.(map[string]interface{})["AccessTokenSecret"] = userInfo.(goth.User).AccessTokenSecret
	basicSessionState.Expires = time.Now().Add(time.Duration(expires) * time.Second).Unix()

	/*

		Can be extracted in Global header settings as:

		X-Origin-Tyk: $tyk_meta.Origin
		X-Tyk-TAP-AccessToken: $tyk_meta.AccessToken
		X-Tyk-TAP-ID: $tyk_meta.AuthProviderUserID
		X-Tyk-TAP-Provider: $tyk_meta.AuthProviderSource

	*/

	keyDataJSON, err := json.Marshal(basicSessionState)

	if err != nil {
		return nil, err
	}

	// Make the Auth request
	response := &TokenResponse{}
	target := strings.Join([]string{string(STANDARD_TOKENS)}, "/")
	data := keyDataJSON

	log.Debug("Request data sent: ", data)

	body := bytes.NewBuffer([]byte(data))
	dErr := t.DispatchAndDecode(Endpoint(target), "POST", DASH, response, UserCred, body, "")

	log.Debug("Returned: ", response)

	if dErr != nil {
		return nil, err
	}

	return response, nil
}

func (t *TykAPI) InvalidateToken(UserCred string, BaseAPI string, token string) error {
	target := strings.Join([]string{string(TOKENS), token}, "/")
	target = strings.Replace(target, "{APIID}", BaseAPI, 1)

	log.Debug("Target is: ", target)
	var reply interface{}
	oErr := t.DispatchAndDecode(Endpoint(target), "DELETE", DASH, &reply, UserCred, nil, "")

	return oErr
}
