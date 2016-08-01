package providers

import (
	"encoding/json"
	"tyk-identity-broker/tap"
	"tyk-identity-broker/tap/identity-handlers"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

var proxyConfig_CODE string = `
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

var proxyConfig_BODY string = `
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

var proxyConfig_REGEX string = `
	{
		"TargetHost" : "http://lonelycode.com",
		"OKCode"     : 0,
		"OKResponse" : "",
		"OKRegex"    : "Code, for one",
		"ResponseIsJson": false,
		"AccessTokenField": "",
		"UsernameField": "",
		"ExrtactUserNameFromBasicAuthHeader": false
	}
`

var BODYFAILURE_STR string = "Authentication Failed"

func getProfile(profileConfig string) tap.Profile {
	provConf := ProxyHandlerConfig{}
	json.Unmarshal([]byte(profileConfig), &provConf)

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

	thisConf := proxyConfig_CODE
	thisProfile := getProfile(thisConf)
	thisProvider := ProxyProvider{}

	thisProvider.Init(identityHandlers.DummyIdentityHandler{}, thisProfile, []byte(thisConf))

	recorder := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/", nil)

	if err != nil {
		t.Fatal(err)
	}

	thisProvider.Handle(recorder, req)
	thisBody, err := ioutil.ReadAll(recorder.Body)

	if recorder.Code != 401 {
		t.Error("Expected 401 as key val, got: ", recorder.Code)
	}

	if string(thisBody) != BODYFAILURE_STR {
		t.Error("Body string incorrect, is: ", thisBody)
	}

}

func TestProxyProvider_BadBody(t *testing.T) {

	thisConf := proxyConfig_BODY
	thisProfile := getProfile(thisConf)
	thisProvider := ProxyProvider{}

	thisProvider.Init(identityHandlers.DummyIdentityHandler{}, thisProfile, []byte(thisConf))

	recorder := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/", nil)

	if err != nil {
		t.Fatal(err)
	}

	thisProvider.Handle(recorder, req)
	thisBody, err := ioutil.ReadAll(recorder.Body)

	if recorder.Code != 401 {
		t.Error("Expected 401 as key val, got: ", recorder.Code)
	}

	if string(thisBody) != BODYFAILURE_STR {
		t.Error("Body string incorrect, is: ", thisBody)
	}

}

func TestProxyProvider_GoodRegex(t *testing.T) {

	thisConf := proxyConfig_REGEX
	thisProfile := getProfile(thisConf)
	thisProvider := ProxyProvider{}

	thisProvider.Init(identityHandlers.DummyIdentityHandler{}, thisProfile, []byte(thisConf))

	recorder := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/", nil)

	if err != nil {
		t.Fatal(err)
	}

	thisProvider.Handle(recorder, req)

	if recorder.Code != 200 {
		t.Error("Expected 200 as key val, got: ", recorder.Code)
	}

}
