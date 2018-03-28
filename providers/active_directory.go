/* package providers is a catch-all for all TAP auth provider types (e.g. social, active directory), if you are
extending TAP to use more providers, add them to this section */
package providers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"crypto/tls"

	"github.com/Sirupsen/logrus"
	"github.com/go-ldap/ldap"
	"github.com/markbates/goth"

	"github.com/TykTechnologies/tyk-identity-broker/tap"
)

// ADProviderLogTag is the log tag for the active directory provider
var ADProviderLogTag = "[AD AUTH]"

// ADProvider is an auth delegation provider for LDAP protocol
type ADProvider struct {
	handler    tap.IdentityHandler
	config     ADConfig
	profile    tap.Profile
	connection *ldap.Conn
}

// ADConfig is the configuration object for an LDAP connector
type ADConfig struct {
	LDAPUseSSL          bool
	LDAPServer          string
	LDAPPort            string
	LDAPUserDN          string
	LDAPBaseDN          string
	LDAPFilter          string
	LDAPEmailAttribute  string
	LDAPAttributes      []string
	LDAPSearchScope     int
	FailureRedirect     string
	DefaultDomain       string
	GetAuthFromBAHeader bool
	SlugifyUserName     bool
}

// Name provides the name of the ID provider
func (s *ADProvider) Name() string {
	return "ADProvider"
}

// ProviderType returns the type of the provider, can be PASSTHROUGH_PROVIDER or REDIRECT dependin on the auth process
// LDAP is a pass -through provider, it will take authentication variables such as username and password and authenticate
// directly with the LDAP server with those values instead of delegating to a third-party such as OAuth.
func (s *ADProvider) ProviderType() tap.ProviderType {
	return tap.PASSTHROUGH_PROVIDER
}

// UseCallback signals whether this provider uses the callback endpoints
func (s *ADProvider) UseCallback() bool {
	return false
}

func (s *ADProvider) connect() {
	log.Debug(ADProviderLogTag + " Connect: starting...")
	var err error
	sName := fmt.Sprintf("%s:%s", s.config.LDAPServer, s.config.LDAPPort)
	log.Debug(ADProviderLogTag+" --> To: ", sName)
	if s.config.LDAPUseSSL {
		tlsconfig := &tls.Config{
			ServerName: s.config.LDAPServer,
		}
		s.connection, err = ldap.DialTLS("tcp", sName, tlsconfig)
	} else {
		s.connection, err = ldap.Dial("tcp", sName)
	}

	if err != nil {
		log.Error(ADProviderLogTag+" Failed to dial: ", err)
		return
	}
	log.Debug(ADProviderLogTag + " Connect: finished...")
}

// Init initialises the handler with it's IdentityHandler (the interface handling actual account SSO on the target)
// profile - the Profile to use for this request and the specific configuration for the handler as a byte stream.
// The config is a byte stream as a hack so we do not need to type cast a map[string]interface{} manually from
// a JSON configuration
func (s *ADProvider) Init(handler tap.IdentityHandler, profile tap.Profile, config []byte) error {
	s.handler = handler
	s.profile = profile

	unmarshallErr := json.Unmarshal(config, &s.config)
	if unmarshallErr != nil {
		return unmarshallErr
	}

	return nil
}

func (s *ADProvider) provideErrorRedirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, s.config.FailureRedirect, http.StatusMovedPermanently) //301
	return
}

func (s *ADProvider) prepFilter(thisUserName string) string {
	newFilter := strings.Replace(s.config.LDAPFilter, "*USERNAME*", thisUserName, -1)
	return newFilter
}

func (s *ADProvider) prepDN(thisUserName string) string {
	fullDN := s.config.LDAPUserDN
	newFilter := strings.Replace(fullDN, "*USERNAME*", thisUserName, -1)
	return newFilter
}

func (s *ADProvider) generateUsername(username string) string {
	var uname string
	if strings.Contains(username, "@") {
		uname = username
	} else {
		asSlug := Slug(username)
		domain := s.config.DefaultDomain
		if s.config.DefaultDomain == "" {
			domain = s.profile.OrgID + "-" + "ADProvider.com"
		}
		uname = asSlug + "@" + domain
	}
	return uname
}

func (s *ADProvider) getUserData(username string) (goth.User, error) {
	log.Info(ADProviderLogTag + " Search: starting...")
	uname := username
	if s.config.SlugifyUserName {
		uname = Slug(username)
	}

	thisUser := goth.User{
		UserID:   uname,
		Provider: "ADProvider",
	}

	if s.config.LDAPFilter == "" {
		log.Info(ADProviderLogTag + " LDAPFilter is blank, skipping")

		var attrs []string
		attrs = s.config.LDAPAttributes
		attrs = append(attrs, s.config.LDAPEmailAttribute)

		thisUser.Email = tap.GenerateSSOKey(thisUser)
		log.Info(ADProviderLogTag+" User Data:", thisUser)

		return thisUser, nil
	}

	DN := s.config.LDAPBaseDN
	if DN == "" {
		DN = s.prepDN(username)
	}

	log.Info(ADProviderLogTag + " Running LDAP search with DN:" + DN + " and Filter: " + s.prepFilter(username))
	// LDAP search is inconcistent, defaulting to using username, assuming username is an email,
	// otherwise we use an algo to create one
	search_request := ldap.NewSearchRequest(
		DN,
		s.config.LDAPSearchScope,
		ldap.DerefAlways,
		0,
		0,
		false,
		s.prepFilter(username),
		s.config.LDAPAttributes,
		nil)

	sr, err := s.connection.Search(search_request)
	if err != nil {
		log.Error(ADProviderLogTag+" Failure in search: ", err)
		return thisUser, err
	}

	if len(sr.Entries) == 0 {
		return thisUser, errors.New("No users match given filter: " + s.prepFilter(username))
	}

	if len(sr.Entries) > 1 {
		return thisUser, errors.New("Filter matched multiple users")
	}

	emailFound := false
	for _, entry := range sr.Entries {
		for _, j := range entry.Attributes {
			log.Info("Checking ", j.Name, "with ", s.config.LDAPEmailAttribute)
			if j.Name == s.config.LDAPEmailAttribute {
				thisUser.Email = j.Values[0]
				emailFound = true
				break
			}
		}
		if emailFound {
			break
		}
	}

	if !emailFound {
		log.Warning("User email not found, generating from username")
		if strings.Contains(username, "@") {
			thisUser.Email = username
		} else {
			thisUser.Email = username + "@" + s.profile.OrgID + "-" + "ADProvider.com"
		}
	}

	return thisUser, nil
}

// Handle is a delegate for the Http Handler used by the generic inbound handler, it will extract the username
// and password from the request and atempt to bind tot he AD host.
func (s *ADProvider) Handle(w http.ResponseWriter, r *http.Request) {
	log.Level = logrus.DebugLevel

	s.connect()

	username := r.FormValue("username")
	password := r.FormValue("password")

	if s.config.GetAuthFromBAHeader {
		username, password = ExtractBAUsernameAndPasswordFromRequest(r)
	}

	if username == "" || password == "" {
		log.Error(ADProviderLogTag + "Login attempt with empty username or password")
		s.provideErrorRedirect(w, r)
		return
	}

	log.Debug("DN: ", s.prepDN(username))

	bindErr := s.connection.Bind(s.prepDN(username), password)

	if bindErr != nil {
		log.Error(ADProviderLogTag+" Bind failed for user: ", username)
		log.Error(ADProviderLogTag+" --> Error was: ", bindErr)
		s.provideErrorRedirect(w, r)
		return
	}
	log.Info(ADProviderLogTag+" User bind successful: ", username)

	user, uErr := s.getUserData(username)
	if uErr != nil {
		log.Error(ADProviderLogTag+" Lookup failed for user: ", username)
		log.Error(ADProviderLogTag+" --> Error was: ", uErr)
		s.provideErrorRedirect(w, r)
		return
	}

	constraintErr := s.checkConstraints(user)
	if constraintErr != nil {
		log.Error(ADProviderLogTag+" Constraint failed: ", constraintErr)
		s.provideErrorRedirect(w, r)
		return
	}

	s.handler.CompleteIdentityAction(w, r, user, s.profile)

	log.Debug(ADProviderLogTag + " Closing connection")
	s.connection.Close()
}

func (s *ADProvider) checkConstraints(user interface{}) error {
	log.Debug(ADProviderLogTag + " Constraints for AD must be set in DN")
	return nil
}

// HandleCallback is not used
func (s *ADProvider) HandleCallback(w http.ResponseWriter, r *http.Request, onError func(tag string, errorMsg string, rawErr error, code int, w http.ResponseWriter, r *http.Request)) {

	log.Warning(ADProviderLogTag + " Callback not implemented for provider")

}
