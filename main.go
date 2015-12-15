package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/gorilla/pat"
	"github.com/lonelycode/tyk-auth-proxy/backends"
	"github.com/lonelycode/tyk-auth-proxy/tap"
	"github.com/lonelycode/tyk-auth-proxy/tothic"
	"net/http"
)

var AuthConfigStore tap.AuthRegisterBackend
var IDHandler tap.IdentityHandler
var log = logrus.New()

var config Configuration

func initBackend(name string, configuration interface{}) {
	found := false

	switch name {
	case "in_memory":
		AuthConfigStore = &backends.InMemoryBackend{}
		found = true
	}

	if !found {
		log.Warning("[MAIN] No backend set!")
		AuthConfigStore = &backends.InMemoryBackend{}
	}

	AuthConfigStore.Init(configuration)
}

func setupTestConfig() {
	/// TEST ONLY

	// SOCIAL
	// ------
	// var testConf string = `
	// {
	// 	"UseProviders": [
	// 		{
	// 			"Name": "gplus",
	// 			"Key": "504206531762-lcdhc8vmveckktcbbevme0n2vgd5v0ve.apps.googleusercontent.com",
	// 			"Secret": "bIboXfuaJh1qnJHi0K_P1MyL"
	// 		}
	// 	],
	// 	"CallbackBaseURL": "http://sharrow.tyk.io:3010"
	// }`

	// testConfig := tap.Profile{
	// 	ID:              "1",
	// 	OrgID:           "TEST",
	// 	ActionType:      tap.GenerateOrLoginDeveloperProfile,
	// 	MatchedPolicyID: "1A",
	// 	Type:            tap.REDIRECT_PROVIDER,
	// 	ProviderName:    "SocialProvider",
	// 	ProviderConfig:  testConf,
	// 	ProviderConstraints: tap.ProfileConstraint{
	// 		Domain: "tyk.io",
	// 		Group:  "",
	// 	},
	// 	ReturnURL: "http://sharrow.tyk.io:3000/bounce",
	// }

	// LDAP
	// ----

	var testConf string = `
	{
		"LDAPServer": "localhost",
		"LDAPPort": "389",
		"LDAPUserDN": "cn=*USERNAME*,dc=test-ldap,dc=tyk,dc=io",
		"LDAPBaseDN": "dc=test-ldap,dc=tyk,dc=io",
		"LDAPFilter": "(cn=*USERNAME*)",
		"LDAPAttributes": [],
		"FailureRedirect": "http://sharrow.tyk.io:3000/failure",
		"SuccessRedirect": "http://sharrow.tyk.io:3000/bounce"
	}`

	testConfig := tap.Profile{
		ID:              "1",
		OrgID:           "TEST",
		ActionType:      tap.GenerateOrLoginDeveloperProfile,
		MatchedPolicyID: "1A",
		Type:            tap.PASSTHROUGH_PROVIDER,
		ProviderName:    "ADProvider",
		ProviderConfig:  testConf,
		ProviderConstraints: tap.ProfileConstraint{
			Domain: "",
			Group:  "",
		},
		ReturnURL: "http://sharrow.tyk.io:3000/bounce",
	}

	// Lets create some configurations!
	inputErr := AuthConfigStore.SetKey("1", testConfig)
	if inputErr != nil {
		log.Error("Couldn't encode configuration: ", inputErr)
	}

	/// END TEST INIT
}

func init() {
	loadConfig("tap.conf", &config)
	initBackend(config.BackEnd.Name, config.BackEnd.BackendSettings)

	// --- Testing
	setupTestConfig()
	// --- End test

	tothic.TothErrorHandler = HandleError
}

func main() {
	p := pat.New()
	p.Get("/auth/{id}/{provider}/callback", HandleAuthCallback)
	p.Post("/auth/{id}/{provider}/callback", HandleAuthCallback)
	p.Post("/auth/{id}/{provider}", HandleAuth)
	p.Get("/auth/{id}/{provider}", HandleAuth)

	log.Info("[MAIN] Listening...")
	http.ListenAndServe(":3010", p)
}
