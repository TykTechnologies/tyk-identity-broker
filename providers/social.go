/* package providers is a catch-all for all TAP auth provider types (e.g. social, active directory), if you are
extending TAP to use more providers, add them to this section */
package providers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Sirupsen/logrus"
	"tyk-identity-broker/tap"
	"tyk-identity-broker/toth"
	"tyk-identity-broker/tothic"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/bitbucket"
	"github.com/markbates/goth/providers/digitalocean"
	"github.com/markbates/goth/providers/dropbox"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/gplus"
	"github.com/markbates/goth/providers/linkedin"
	"github.com/markbates/goth/providers/twitter"
	"net/http"
	"strings"
)

var log = logrus.New()

// SocialLogTag is the log tag for the social provider
var SocialLogTag = "[SOCIAL AUTH]"

// Social is the identity handler for all social auth, it is a wrapper around Goth, and makes use of it's pluggable
// providers to provide a raft of social OAuth providers as SSO or Login delegates.
type Social struct {
	handler tap.IdentityHandler
	config  GothConfig
	toth    toth.TothInstance
	profile tap.Profile
}

// GothProviderConfig the configurations required for the individual goth providers
type GothProviderConfig struct {
	Name   string
	Key    string
	Secret string
}

// GothConfig is the main configuration object for the Social provider
type GothConfig struct {
	UseProviders    []GothProviderConfig
	CallbackBaseURL string
	FailureRedirect string
	CORS            bool
	CORSOrigin      string
	CORSHeaders     string
	CORSMaxAge      string
}

// Name returns the name of the provder
func (s *Social) Name() string {
	return "SocialProvider"
}

func (p *Social) GetProfile() tap.Profile {
	return p.profile
}

func (p *Social) GetHandler() tap.IdentityHandler {
	return p.handler
}

func (p *Social) GetLogTag() string {
	return ProxyLogTag
}

func (p *Social) GetCORS() bool {
	return p.config.CORS
}

func (p *Social) GetCORSOrigin() string {
	return p.config.CORSOrigin
}

func (p *Social) GetCORSHeaders() string {
	return p.config.CORSHeaders
}

func (p *Social) GetCORSMaxAge() string {
	return p.config.CORSMaxAge
}

// ProviderType returns the type of the provider, Social makes use of the reirect type, as
// it redirects the user to multiple locations in the flow
func (s *Social) ProviderType() tap.ProviderType {
	return tap.REDIRECT_PROVIDER
}

// UseCallback returns whether or not the callback URL is used for this profile. Social uses it.
func (s *Social) UseCallback() bool {
	return true
}

// Init will configure the social provider for this request.
func (s *Social) Init(handler tap.IdentityHandler, profile tap.Profile, config []byte) error {
	s.handler = handler
	s.profile = profile

	s.toth = toth.TothInstance{}
	s.toth.Init()

	unmarshallErr := json.Unmarshal(config, &s.config)
	if unmarshallErr != nil {
		return unmarshallErr
	}

	// TODO: Add more providers here
	gothProviders := []goth.Provider{}
	for _, provider := range s.config.UseProviders {
		switch provider.Name {
		case "gplus":
			gothProviders = append(gothProviders, gplus.New(provider.Key, provider.Secret, s.getCallBackURL(provider.Name)))

		case "github":
			gothProviders = append(gothProviders, github.New(provider.Key, provider.Secret, s.getCallBackURL(provider.Name)))

		case "twitter":
			gothProviders = append(gothProviders, twitter.New(provider.Key, provider.Secret, s.getCallBackURL(provider.Name)))

		case "linkedin":
			gothProviders = append(gothProviders, linkedin.New(provider.Key, provider.Secret, s.getCallBackURL(provider.Name)))

		case "dropbox":
			gothProviders = append(gothProviders, dropbox.New(provider.Key, provider.Secret, s.getCallBackURL(provider.Name)))

		case "digitalocean":
			gothProviders = append(gothProviders, digitalocean.New(provider.Key, provider.Secret, s.getCallBackURL(provider.Name)))

		case "bitbucket":
			gothProviders = append(gothProviders, bitbucket.New(provider.Key, provider.Secret, s.getCallBackURL(provider.Name)))
		}
	}

	s.toth.UseProviders(gothProviders...)
	return nil
}

// Handle is the main callback delegate for the generic auth flow
func (s *Social) Handle(w http.ResponseWriter, r *http.Request) (goth.User, error) {
	tothic.BeginAuthHandler(w, r, &s.toth)
	var user goth.User
	return user, errors.New("NOT IMPLEMENTED")
}

func (s *Social) checkConstraints(user interface{}) error {
	var thisUser goth.User
	thisUser = user.(goth.User)

	if s.profile.ProviderConstraints.Domain != "" {
		if !strings.HasSuffix(thisUser.Email, s.profile.ProviderConstraints.Domain) {
			return errors.New("Domain constraint failed, user domain does not match profile")
		}
	}

	if s.profile.ProviderConstraints.Group != "" {
		log.Warning("Social Auth does not support Group constraints")
	}

	return nil
}

// HandleCallback handles the callback from the OAuth provider
func (s *Social) HandleCallback(w http.ResponseWriter, r *http.Request, onError func(tag string, errorMsg string, rawErr error, code int, w http.ResponseWriter, r *http.Request)) {
	user, err := tothic.CompleteUserAuth(w, r, &s.toth)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	constraintErr := s.checkConstraints(user)
	if constraintErr != nil {
		if s.config.FailureRedirect == "" {
			onError(SocialLogTag, "Constraint failed", constraintErr, 400, w, r)
			return
		}

		http.Redirect(w, r, s.config.FailureRedirect, 301)
		return
	}

	// Complete login and redirect
	s.handler.CompleteIdentityAction(w, r, user, s.profile)
}

func (s *Social) getCallBackURL(provider string) string {
	return s.config.CallbackBaseURL + "/auth/" + s.profile.ID + "/" + provider + "/callback"
}
