package providers

import (
	"encoding/json"
	"fmt"
	"github.com/lonelycode/tyk-auth-proxy/tap"
	"github.com/lonelycode/tyk-auth-proxy/toth"
	"github.com/lonelycode/tyk-auth-proxy/tothic"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/gplus"
	"net/http"
)

type Social struct {
	handler tap.IdentityHandler
	config  GothConfig
	toth    toth.TothInstance
}

type GothProviderConfig struct {
	Name   string
	Key    string
	Secret string
}

type GothConfig struct {
	UseProviders []GothProviderConfig
}

func (s *Social) Name() string {
	return "SocialProvider"
}

func (s *Social) ProviderType() tap.ProviderType {
	return tap.REDIRECT_PROVIDER
}

func (s *Social) UseCallback() bool {
	return true
}

func (s *Social) Init(handler tap.IdentityHandler, config []byte) error {
	s.handler = handler

	s.toth = toth.TothInstance{}
	s.toth.Init()

	unmarshallErr := json.Unmarshal(config, &s.config)
	if unmarshallErr != nil {
		return unmarshallErr
	}

	gothProviders := []goth.Provider{}
	for _, provider := range s.config.UseProviders {
		switch provider.Name {
		case "gplus":
			gothProviders = append(gothProviders, gplus.New(provider.Key, provider.Secret, s.getCallBackURL(provider.Name)))
		}
	}

	s.toth.UseProviders(gothProviders...)
	return nil
}

func (s *Social) Handle(w http.ResponseWriter, r *http.Request) {
	tothic.BeginAuthHandler(w, r, &s.toth)
}

func (s *Social) HandleCallback(w http.ResponseWriter, r *http.Request) {
	// print our state string to the console
	fmt.Println(gothic.GetState(r))

	user, err := tothic.CompleteUserAuth(w, r, &s.toth)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	fmt.Println(user)
	fmt.Fprintln(w, "WHEEEE")
}

func (s *Social) getCallBackURL(provider string) string {
	return "http://sharrow.tyk.io:3000/auth/" + provider + "/callback"
}
