package providers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/matryer/is"

	"github.com/TykTechnologies/tyk-identity-broker/tap"
	identityHandlers "github.com/TykTechnologies/tyk-identity-broker/tap/identity-handlers"
)

const proxyConfig_CODE = `
	{
		"TargetHost" : "http://lonelycode.com/doesnotexist",
		"OKCode"     : 200,
		"OKResponse" : "",
		"OKRegex"    : "",
		"ResponseIsJson": false,
		"AccessTokenField": "",
		"UsernameField": "",
		"ExrtactUserNameFromBasicAuthHeader": false
	}
`

const proxyConfig_BODY = `
	{
		"TargetHost" : "http://lonelycode.com",
		"OKCode"     : 0,
		"OKResponse" : "This wont match",
		"OKRegex"    : "",
		"ResponseIsJson": false,
		"AccessTokenField": "",
		"UsernameField": "",
		"ExrtactUserNameFromBasicAuthHeader": false
	}
`

const proxyConfig_REGEX = `
	{
		"TargetHost" : "https://lonelycode.com",
		"OKCode"     : 0,
		"OKResponse" : "",
		"OKRegex"    : "digital hippie",
		"ResponseIsJson": false,
		"AccessTokenField": "",
		"UsernameField": "",
		"ExrtactUserNameFromBasicAuthHeader": false
	}
`

const BODYFAILURE_STR = "Authentication Failed"

func getProfile(t *testing.T, profileConfig string) tap.Profile {
	t.Helper()

	is := is.New(t)

	provConf := ProxyHandlerConfig{}
	is.NoErr(json.Unmarshal([]byte(profileConfig), &provConf))

	thisProfile := tap.Profile{
		ID:                    "1",
		OrgID:                 "12345",
		ActionType:            "GenerateTemporaryAuthToken",
		Type:                  "passthrough",
		ProviderName:          "ProxyProvider",
		ProviderConfig:        provConf,
		IdentityHandlerConfig: new(interface{}),
	}

	return thisProfile
}

func TestProxyProvider_BadCode(t *testing.T) {
	is := is.New(t)

	thisConf := proxyConfig_CODE
	thisProfile := getProfile(t, thisConf)
	thisProvider := ProxyProvider{}

	is.NoErr(thisProvider.Init(identityHandlers.DummyIdentityHandler{}, thisProfile, []byte(thisConf)))

	recorder := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	thisProvider.Handle(recorder, req, nil, thisProfile)
	thisBody, err := ioutil.ReadAll(recorder.Body)
	is.NoErr(err)

	if recorder.Code != 401 {
		t.Fatalf("Expected 401 response code, got '%d'", recorder.Code)
	}

	if string(thisBody) != BODYFAILURE_STR {
		t.Fatalf("Body string '%s' is incorrect", thisBody)
	}
}

func TestProxyProvider_BadBody(t *testing.T) {
	is := is.New(t)

	thisConf := proxyConfig_BODY
	thisProfile := getProfile(t, thisConf)
	thisProvider := ProxyProvider{}

	is.NoErr(thisProvider.Init(identityHandlers.DummyIdentityHandler{}, thisProfile, []byte(thisConf)))

	recorder := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	thisProvider.Handle(recorder, req, nil, thisProfile)
	thisBody, err := ioutil.ReadAll(recorder.Body)
	is.NoErr(err)

	if recorder.Code != 401 {
		t.Fatalf("Expected 401 response code, got '%d'", recorder.Code)
	}

	if string(thisBody) != BODYFAILURE_STR {
		t.Fatalf("Body string '%s' is incorrect", thisBody)
	}
}

func TestProxyProvider_GoodRegex(t *testing.T) {
	is := is.New(t)

	thisConf := proxyConfig_REGEX
	thisProfile := getProfile(t, thisConf)
	thisProvider := ProxyProvider{}

	is.NoErr(thisProvider.Init(identityHandlers.DummyIdentityHandler{}, thisProfile, []byte(thisConf)))

	recorder := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	thisProvider.Handle(recorder, req, nil, thisProfile)

	if recorder.Code != 200 {
		t.Fatalf("Expected 200 response code, got '%v'", recorder.Code)
	}
}
