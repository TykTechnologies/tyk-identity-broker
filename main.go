package main

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/gorilla/pat"
	"github.com/lonelycode/tyk-auth-proxy/backends"
	"github.com/lonelycode/tyk-auth-proxy/tap"
	"net/http"
)

var AuthConfigStore tap.AuthRegisterBackend
var log = logrus.New()

func initBackend(name string) {
	found := false

	switch name {
	case "in_memory":
		AuthConfigStore = &backends.InMemoryBackend{}
		found = true
	}

	if !found {
		fmt.Println("No backend set!")
		AuthConfigStore = &backends.InMemoryBackend{}
	}

	AuthConfigStore.Init()
}

func init() {

	/// TEST ONLY

	initBackend("in_memory")

	var config string = `
	{
		"UseProviders": [
			{
				"Name": "gplus",
				"Key": "504206531762-lcdhc8vmveckktcbbevme0n2vgd5v0ve.apps.googleusercontent.com",
				"Secret": "bIboXfuaJh1qnJHi0K_P1MyL"
			}
		]
	}`

	testConfig := tap.Profile{
		ID:              "1",
		OrgID:           "TEST",
		ActionType:      tap.GenerateOrLoginDeveloperProfile,
		MatchedPolicyID: "1A",
		Type:            tap.REDIRECT_PROVIDER,
		ProviderName:    "SocialProvider",
		ProviderConfig:  config,
	}

	// Lets create some configurations!
	AuthConfigStore.SetKey("1", testConfig)

	/// END TEST INIT
}

func main() {
	p := pat.New()
	p.Get("/auth/{id}/{provider}/callback", HandleAuthCallback) // TODO: WRITE THESE!!!
	p.Get("/auth/{id}/{provider}", HandleAuth)

	fmt.Println("Listening...")
	http.ListenAndServe(":3000", p)
}
