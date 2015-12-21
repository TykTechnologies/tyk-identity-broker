package tyk

import (
	"bytes"
	"encoding/json"
	"errors"
	log "github.com/Sirupsen/logrus"
	"io"
	"io/ioutil"
	"labix.org/v2/mgo/bson"
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

const (
	// Main endpoints used in this wrapper
	PORTAL_DEVS     Endpoint = "/api/portal/developers/email"
	PORTAL_DEVS_SSO Endpoint = "/api/portal/developers/ssokey"
	PORTAL_DEV      Endpoint = "/api/portal/developers"
	SSO             Endpoint = "/admin/sso"

	// Main APis used in this wrapper
	GATEWAY    TykAPIName = "gateway"
	DASH       TykAPIName = "dash"
	DASH_SUPER TykAPIName = "dash_super"
)

// DispatchDashboard dispatches a request to the dashboard API and handles the response
func (t *TykAPI) DispatchDashboard(target Endpoint, method string, usercode string, body io.Reader) ([]byte, error) {
	preparedEndpoint := t.DashboardConfig.Endpoint + ":" + t.DashboardConfig.Port + string(target)

	log.Warning("Calling: ", preparedEndpoint)
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

	log.Warning("Calling: ", preparedEndpoint)
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
func (t *TykAPI) DispatchGateway(target Endpoint, method string, body io.Reader) ([]byte, error) {
	preparedEndpoint := t.GatewayConfig.Endpoint + ":" + t.GatewayConfig.Port + string(target)

	log.Warning("Calling: ", preparedEndpoint)
	newRequest, err := http.NewRequest(method, preparedEndpoint, body)
	if err != nil {
		log.Error("Failed to create request")
		log.Error(err)
	}

	newRequest.Header.Add("x-tyk-authorization", t.GatewayConfig.AdminSecret)
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

	return retBody, nil
}

// Dcode will unmarshal a request body, a bit redundant tbh
func (t *TykAPI) Decode(raw []byte, retVal interface{}) error {
	decErr := json.Unmarshal(raw, &retVal)
	return decErr
}

// DispatchAndDecode will select the API to target, dispatch the request, then decode ther esponse to return to the caller
func (t *TykAPI) DispatchAndDecode(target Endpoint, method string, APIName TykAPIName, retVal interface{}, creds string, body io.Reader) error {
	var retBytes []byte
	var dispatchErr error

	switch APIName {
	case GATEWAY:
		retBytes, dispatchErr = t.DispatchGateway(target, method, body)
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
	dErr := t.DispatchAndDecode(Endpoint(target), "POST", DASH_SUPER, &returnVal, "", body)

	return returnVal, dErr
}

// GetDeveloper will retrieve a deverloper from the Advanced API using their Email address
func (t *TykAPI) GetDeveloper(UserCred string, DeveloperEmail string) (PortalDeveloper, error) {
	asStr := url.QueryEscape(DeveloperEmail)
	target := strings.Join([]string{string(PORTAL_DEVS), asStr}, "/")

	retUser := PortalDeveloper{}

	dErr := t.DispatchAndDecode(Endpoint(target), "GET", DASH, &retUser, UserCred, nil)

	return retUser, dErr
}

// GetDeveloperBySSOKey will retrieve a deverloper from the Advanced API using their SSO Key address
func (t *TykAPI) GetDeveloperBySSOKey(UserCred string, DeveloperEmail string) (PortalDeveloper, error) {
	asStr := url.QueryEscape(DeveloperEmail)
	target := strings.Join([]string{string(PORTAL_DEVS_SSO), asStr}, "/")

	retUser := PortalDeveloper{}

	dErr := t.DispatchAndDecode(Endpoint(target), "GET", DASH, &retUser, UserCred, nil)

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

	dErr := t.DispatchAndDecode(Endpoint(target), "PUT", DASH, &retData, UserCred, body)

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

	dErr := t.DispatchAndDecode(Endpoint(target), "POST", DASH, &retData, UserCred, body)
	log.Info("Returned: ", retData)

	return dErr
}
