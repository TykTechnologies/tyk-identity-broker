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

type MinimalApiDef struct {
	APIs []struct {
		ApiDefinition struct {
			ID string `json:"id"`
		} `json:"api_definition"`
	} `json:"apis"`
}

type EndpointConfig struct {
	Endpoint    string
	Port        string
	AdminSecret string
}

type TykAPI struct {
	GatewayConfig   EndpointConfig
	DashboardConfig EndpointConfig
}

type ApiDocument struct {
	ApiHumanName string `json:"api_human_name"`
	APIID        string `json:"api_id"`
}

type EventConfig struct {
	Webhook string `bson:"webhook" json:"webhook"`
	Email   string `bson:"email" json:"email"`
	Redis   bool   `bson:"redis" json:"redis"`
}

type OrganisationDocument struct {
	Id             bson.ObjectId          `json:"id,omitempty"`
	OwnerName      string                 `json:"owner_name"`
	OwnerSlug      string                 `json:"owner_slug"`
	CNAMEEnabled   bool                   `json:"cname_enabled"`
	CNAME          string                 `json:"cname"`
	Apis           []ApiDocument          `json:"apis"`
	DeveloperQuota int                    `json:"developer_quota"`
	DeveloperCount int                    `json:"developer_count"`
	HybridEnabled  bool                   `json:"hybrid_enabled"`
	Events         map[string]EventConfig `bson:"event_options" json:"event_options"`
}

type SessionState struct {
	Allowance        float64     `json:"allowance"`
	Rate             float64     `json:"rate"`
	Per              float64     `json:"per"`
	Expires          int64       `json:"expires"`
	QuotaMax         int64       `json:"quota_max"`
	QuotaRenews      int64       `json:"quota_renews"`
	QuotaRemaining   int64       `json:"quota_remaining"`
	QuotaRenewalRate int64       `json:"quota_renewal_rate"`
	OrgID            string      `json:"org_id"`
	IsInactive       bool        `json:"is_inactive"`
	MetaData         interface{} `json:"meta_data"`
	DataExpires      int64       `json:"data_expires"`
	Monitor          struct {
		TriggerLimits []float64 `json:"trigger_limits"`
	} `json:"monitor"`
}

type TykDashboardUser struct {
	FirstName    string        `json:"first_name"`
	LastName     string        `json:"last_name"`
	EmailAddress string        `json:"email_address"`
	Password     string        `json:"password"`
	OrgID        string        `json:"org_id"`
	Active       bool          `json:"active"`
	Id           bson.ObjectId `json:"id,omitempty"`
	AccessKey    string        `json:"access_key"`
}

type TykDashboardUserlist struct {
	Users []TykDashboardUser `json:"users"`
}

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
}

type ReturnDataStruct struct {
	Data  []interface{}
	Pages int
}

type Endpoint string
type TykAPIName string

const (
	ORGANISATION_KEY Endpoint = "/tyk/org/keys"
	ORGANISATIONS    Endpoint = "/admin/organisations"
	SUPER_USERS      Endpoint = "/admin/users"
	USERS            Endpoint = "/api/users"
	PASSWORD         Endpoint = "/api/users/USERID/actions/reset"
	APIS             Endpoint = "/api/apis"
	PORTAL_CONFIG    Endpoint = "/api/portal/configuration"
	PORTAL_PAGE      Endpoint = "/api/portal/pages"
	PORTAL_CSS       Endpoint = "/api/portal/css"
	PORTAL_MENUS     Endpoint = "/api/portal/menus"
	PORTAL_DEVS      Endpoint = "/api/portal/developers/email"
	PORTAL_DEV       Endpoint = "/api/portal/developers"
	SSO              Endpoint = "/admin/sso"

	GATEWAY    TykAPIName = "gateway"
	DASH       TykAPIName = "dash"
	DASH_SUPER TykAPIName = "dash_super"
)

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

func (t *TykAPI) Decode(raw []byte, retVal interface{}) error {
	decErr := json.Unmarshal(raw, &retVal)
	return decErr
}

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

func (t *TykAPI) GetOrg(orgId string) (OrganisationDocument, error) {
	thisOrg := OrganisationDocument{}
	target := strings.Join([]string{string(ORGANISATIONS), orgId}, "/")
	err := t.DispatchAndDecode(Endpoint(target), "GET", DASH_SUPER, &thisOrg, "", nil)

	return thisOrg, err
}

func (t *TykAPI) DeleteOrg(orgId string) error {
	var ret interface{}
	target := strings.Join([]string{string(ORGANISATIONS), orgId}, "/")
	err := t.DispatchAndDecode(Endpoint(target), "DELETE", DASH_SUPER, &ret, "", nil)
	log.Info(ret)
	return err
}

func (t *TykAPI) GetOrgKey(orgId string) (SessionState, error) {
	thisSession := SessionState{}
	target := strings.Join([]string{string(ORGANISATION_KEY), orgId}, "/")
	err := t.DispatchAndDecode(Endpoint(target), "GET", GATEWAY, &thisSession, "", nil)

	return thisSession, err
}

func (t *TykAPI) SetOrgKey(orgId string, thisSession SessionState) error {
	target := strings.Join([]string{string(ORGANISATION_KEY), orgId}, "/")

	orgDataJson, err := json.Marshal(thisSession)

	if err != nil {
		return err
	}

	body := bytes.NewBuffer(orgDataJson)
	var reply interface{}
	oErr := t.DispatchAndDecode(Endpoint(target), "PUT", GATEWAY, &reply, "", body)

	return oErr
}

func (t *TykAPI) UpdateOrg(orgId string, thisOrg OrganisationDocument) error {
	target := strings.Join([]string{string(ORGANISATIONS), orgId}, "/")
	var statusMessage interface{}

	orgDataJson, err := json.Marshal(thisOrg)

	if err != nil {
		return err
	}

	body := bytes.NewBuffer(orgDataJson)
	dErr := t.DispatchAndDecode(Endpoint(target), "PUT", DASH_SUPER, &statusMessage, "", body)

	return dErr
}

func (t *TykAPI) CreateOrg(thisOrg OrganisationDocument) (string, error) {
	target := strings.Join([]string{string(ORGANISATIONS)}, "/")
	statusMessage := make(map[string]string)

	orgDataJson, err := json.Marshal(thisOrg)

	if err != nil {
		return "", err
	}

	body := bytes.NewBuffer(orgDataJson)
	dErr := t.DispatchAndDecode(Endpoint(target), "POST", DASH_SUPER, &statusMessage, "", body)

	return statusMessage["Meta"], dErr
}

func (t *TykAPI) CreateUser(thisUser TykDashboardUser) (string, error) {
	target := strings.Join([]string{string(SUPER_USERS)}, "/")
	statusMessage := make(map[string]string)

	userDataJson, err := json.Marshal(thisUser)

	if err != nil {
		return "", err
	}

	body := bytes.NewBuffer(userDataJson)
	dErr := t.DispatchAndDecode(Endpoint(target), "POST", DASH_SUPER, &statusMessage, "", body)

	return statusMessage["Message"], dErr
}

func (t *TykAPI) GetUsers(UserCred string) (TykDashboardUserlist, error) {
	target := strings.Join([]string{string(USERS)}, "/")
	// userList := make([]TykDashboardUser, 0)

	retList := TykDashboardUserlist{}

	dErr := t.DispatchAndDecode(Endpoint(target), "GET", DASH, &retList, UserCred, nil)

	return retList, dErr
}

func (t *TykAPI) GetAPIS(UserCred string) (MinimalApiDef, error) {
	target := strings.Join([]string{string(APIS)}, "/")
	// userList := make([]TykDashboardUser, 0)

	retList := MinimalApiDef{}

	dErr := t.DispatchAndDecode(Endpoint(target), "GET", DASH, &retList, UserCred, nil)

	return retList, dErr
}

func (t *TykAPI) DeleteUser(UserCred string, UserId string) error {
	target := strings.Join([]string{string(USERS), UserId}, "/")
	// userList := make([]TykDashboardUser, 0)

	var ret interface{}

	dErr := t.DispatchAndDecode(Endpoint(target), "DELETE", DASH, &ret, UserCred, nil)
	log.Info(ret)
	return dErr
}

func (t *TykAPI) DeleteAPI(UserCred string, ApiID string) error {
	target := strings.Join([]string{string(APIS), ApiID}, "/")
	// userList := make([]TykDashboardUser, 0)

	var ret interface{}

	dErr := t.DispatchAndDecode(Endpoint(target), "DELETE", DASH, &ret, UserCred, nil)
	log.Info(ret)

	return dErr
}

func (t *TykAPI) UpdateUserPassword(UserCred, UserId, NewPass string) error {
	target := strings.Join([]string{string(PASSWORD)}, "/")
	target = strings.Replace(target, "USERID", UserId, 1)

	type UserPassword struct {
		NewPassword string `json:"new_password"`
	}

	thisPass := UserPassword{NewPass}
	userDataJson, err := json.Marshal(thisPass)

	if err != nil {
		return err
	}

	var returnVal interface{}
	body := bytes.NewBuffer(userDataJson)
	dErr := t.DispatchAndDecode(Endpoint(target), "POST", DASH, &returnVal, UserCred, body)

	return dErr
}

func (t *TykAPI) SetOrgStatus(orgId string, disabled bool) error {
	orgSession, oErr := t.GetOrgKey(orgId)
	if oErr != nil {
		return oErr
	}

	orgSession.IsInactive = disabled

	orgSetErr := t.SetOrgKey(orgId, orgSession)
	if orgSetErr != nil {
		return orgSetErr
	}

	return nil
}

func (t *TykAPI) ChangeOrgQuota(orgId string, newSessionState SessionState, devQuota int, hybridEnabled bool, enableCNAME bool) error {
	// Set the KeyCount
	orgSetErr := t.SetOrgKey(orgId, newSessionState)
	if orgSetErr != nil {
		return orgSetErr
	}

	log.Info("Set ORG Data for: ", orgId)

	thisOrgData, getErr := t.GetOrg(orgId)
	if getErr != nil {
		return getErr
	}

	if thisOrgData.DeveloperCount >= devQuota {
		return errors.New("Failed to update developer quota, org has more devs than new quota!")
	}

	log.Info("Setting dashboard data for org")

	thisOrgData.DeveloperQuota = devQuota
	thisOrgData.HybridEnabled = hybridEnabled
	thisOrgData.CNAMEEnabled = enableCNAME
	updateErr := t.UpdateOrg(orgId, thisOrgData)

	log.Info("Done.")
	if updateErr != nil {
		return updateErr
	}

	return nil
}

func (t *TykAPI) GetOrgDevCount(orgId string) (int, error) {
	thisOrgData, getErr := t.GetOrg(orgId)
	if getErr != nil {
		return -1, getErr
	}

	return thisOrgData.DeveloperCount, nil
}

func (t *TykAPI) GenericPortalCreate(UserCred, OrgId string, data interface{}, target Endpoint) error {
	target = Endpoint(strings.Join([]string{string(target)}, "/"))
	userDataJson, err := json.Marshal(data)

	if err != nil {
		return err
	}

	var returnVal interface{}
	body := bytes.NewBuffer(userDataJson)
	dErr := t.DispatchAndDecode(Endpoint(target), "POST", DASH, &returnVal, UserCred, body)

	return dErr
}

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

func (t *TykAPI) GetDeveloper(UserCred string, DeveloperEmail string) (PortalDeveloper, error) {
	asStr := url.QueryEscape(DeveloperEmail)
	target := strings.Join([]string{string(PORTAL_DEVS), asStr}, "/")

	retUser := PortalDeveloper{}

	dErr := t.DispatchAndDecode(Endpoint(target), "GET", DASH, &retUser, UserCred, nil)

	return retUser, dErr
}

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

func (t *TykAPI) CreateDeveloper(UserCred string, dev PortalDeveloper) error {
	target := strings.Join([]string{string(PORTAL_DEV)}, "/")

	retData := map[string]interface{}{}
	data, err := json.Marshal(dev)
	body := bytes.NewBuffer(data)

	if err != nil {
		return err
	}

	dErr := t.DispatchAndDecode(Endpoint(target), "POST", DASH, &retData, UserCred, body)

	return dErr
}
