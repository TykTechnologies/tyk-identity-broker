package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/pat"
	"github.com/lonelycode/tyk-auth-proxy/providers"
	"github.com/lonelycode/tyk-auth-proxy/tap"
)

var AuthConfigStore AuthRegisterBackend

func initBackend(name string) {
	found := false

	switch name {
	case "in_memory":
		AuthConfigStore = backends.InMemoryBackend{}
		found = true
	}

	if !found {
		fmt.Println("No backend set!")
		AuthConfigStore = backends.InMemoryBackend{}
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
				"Key": "504206531762-e3nk43d2svtut98odmknrclf6aa1hd4n.apps.googleusercontent.com",
				"Secret": "kRqL0F0ysPiM2sv-oyEwkw2F"
			}
		]
	}`

	testConfig := tap.Profile{
		ID:              "1",
		OrgID:           "TEST",
		ActionType:      tap.GenerateOrLoginDeveloperProfile,
		MatchedPolicyID: "1",
		Type:            tap.REDIRECT_PROVIDER,
		ProviderName:    "SocialProvider",
		ProviderConfig:  config,
	}

	// Lets create some configurations!
	AuthConfigStore.SetKey("configs", []tap.Profile{testConfig})

	/// END TEST INIT
}

func main() {

	var config string = `
	{
		"UseProviders": [
			{
				"Name": "gplus",
				"Key": "504206531762-e3nk43d2svtut98odmknrclf6aa1hd4n.apps.googleusercontent.com",
				"Secret": "kRqL0F0ysPiM2sv-oyEwkw2F"
			}
		]
	}
	`
	var theseConfigs []tap.Profile
	theseConfigs := AuthConfigStore.GetKey("configs")

	p := pat.New()

	for _, conf := range theseConfigs {
		var thisProvider tap.TAProvider

		switch conf.ProviderName {
		case "SocialProvider":
			thisProvider = providers.Social{}
		}

		var thisIdentityHandler IdentityHandler

		switch conf.ActionType {
		case GenerateOrLoginDeveloperProfile:
			thisIdentityHandler = tap.DummyIdentityHandler{} // TODO: Change These
		case GenerateOrLoginUserProfile:
			thisIdentityHandler = tap.DummyIdentityHandler{} // TODO: Change These
		}

		thisProvider.Init(thisIdentityHandler, []byte(conf.ProviderConfig))
	}

	p.Get("/auth/{id}/{provider}/callback", HandleAuthCallback)
	p.Get("/auth/{id}/{provider}", HandleAuth)

	fmt.Println("Listening...")
	http.ListenAndServe(":3000", p)
}
