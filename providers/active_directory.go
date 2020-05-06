/* package providers is a catch-all for all TAP auth provider types (e.g. social, active directory), if you are
extending TAP to use more providers, add them to this section */
package providers

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/go-ldap/ldap"
	"github.com/markbates/goth"

	logger "github.com/TykTechnologies/tyk-identity-broker/log"
	"github.com/TykTechnologies/tyk-identity-broker/tap"
	"github.com/sirupsen/logrus"
)

var onceReloadADLogger sync.Once
var ADLogTag =  "AD AUTH"
var ADLogger = log.WithField("prefix", ADLogTag)

// ADProvider is an auth delegation provider for LDAP protocol
type ADProvider struct {
	handler    tap.IdentityHandler
	config     ADConfig
	profile    tap.Profile
	connection *ldap.Conn
}

// ADConfig is the configuration object for an LDAP connector
type ADConfig struct {
	LDAPUseSSL             bool
	LDAPServer             string
	LDAPPort               string
	LDAPUserDN             string
	LDAPBaseDN             string
	LDAPFilter             string
	LDAPEmailAttribute     string
	LDAPFirstNameAttribute string
	LDAPLastNameAttribute  string
	LDAPAdminUser          string
	LDAPAdminPassword      string
	LDAPAttributes         []string
	LDAPSearchScope        int
	FailureRedirect        string
	DefaultDomain          string
	GetAuthFromBAHeader    bool
	SlugifyUserName        bool
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
	ADLogger.Debug("Connect: starting...")
	var err error
	sName := fmt.Sprintf("%s:%s", s.config.LDAPServer, s.config.LDAPPort)
	ADLogger.Debug("--> To: ", sName)
	if s.config.LDAPUseSSL {
		tlsconfig := &tls.Config{
			ServerName: s.config.LDAPServer,
		}
		s.connection, err = ldap.DialTLS("tcp", sName, tlsconfig)
	} else {
		s.connection, err = ldap.Dial("tcp", sName)
	}

	if err != nil {
		ADLogger.WithFields(logrus.Fields{
			"path":  sName,
			"error": err,
		}).Error("Failed to dial")
		return
	}
	ADLogger.Debug("Connect: finished...")
}

// Init initialises the handler with it's IdentityHandler (the interface handling actual account SSO on the target)
// profile - the Profile to use for this request and the specific configuration for the handler as a byte stream.
// The config is a byte stream as a hack so we do not need to type cast a map[string]interface{} manually from
// a JSON configuration
func (s *ADProvider) Init(handler tap.IdentityHandler, profile tap.Profile, config []byte) error {
	//if an external logger was set, then lets reload it to inherit those configs
	onceReloadADLogger.Do(func() {
		log = logger.Get()
		ADLogger = &logrus.Entry{Logger:log}
		ADLogger = ADLogger.Logger.WithField("prefix", ADLogTag)
	})

	s.handler = handler
	s.profile = profile

	unmarshallErr := json.Unmarshal(config, &s.config)
	if unmarshallErr != nil {
		return unmarshallErr
	}

	return nil
}

func (s *ADProvider) provideErrorRedirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, s.config.FailureRedirect, 301)
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

func (s *ADProvider) getUserData(username string, password string) (goth.User, error) {
	ADLogger.Info("Search: starting...")
	uname := username
	if s.config.SlugifyUserName {
		uname = Slug(username)
	}

	thisUser := goth.User{
		UserID:   uname,
		Provider: "ADProvider",
		RawData:  make(map[string]interface{}),
	}

	if s.config.LDAPFilter == "" {
		s.config.LDAPFilter = "(objectclass=*)"
	}

	DN := s.config.LDAPBaseDN
	if DN == "" {
		DN = s.prepDN(username)
	}

	ADLogger.WithFields(logrus.Fields{
		"DN":     DN,
		"Filter": s.prepFilter(username),
	}).Info("Running LDAP search")

	// LDAP search is inconsistent, defaulting to using username, assuming username is an email,
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
		ADLogger.WithFields(logrus.Fields{
			"error": err,
		}).Error("Failure in search")
		return thisUser, err
	}

	if len(sr.Entries) == 0 {
		return thisUser, errors.New("No users match given filter: " + s.prepFilter(username))
	}

	if len(sr.Entries) > 1 {
		return thisUser, errors.New("Filter matched multiple users")
	}

	entry := sr.Entries[0]

	if s.config.LDAPEmailAttribute == "" {
		s.config.LDAPEmailAttribute = "mail"
	}

	if s.config.LDAPFirstNameAttribute == "" {
		s.config.LDAPFirstNameAttribute = "givenName"
	}

	if s.config.LDAPLastNameAttribute == "" {
		s.config.LDAPLastNameAttribute = "sn"
	}

	if s.config.LDAPAdminUser != "" {
		bindErr := s.connection.Bind(entry.DN, password)
		if bindErr != nil {
			ADLogger.WithFields(logrus.Fields{
				"username": username,
				"error":    bindErr,
			}).Error("Bind failed for user")
			return thisUser, errors.New("Password not matched")
		}
		ADLogger.WithField("username", username).Info("User bind successful")
	}

	emailFound := false
	for _, j := range entry.Attributes {
		if j.Name == s.config.LDAPEmailAttribute {
			thisUser.Email = j.Values[0]
			emailFound = true
		}

		if j.Name == s.config.LDAPFirstNameAttribute {
			thisUser.FirstName = j.Values[0]
		}

		if j.Name == s.config.LDAPLastNameAttribute {
			thisUser.LastName = j.Values[0]
		}

		thisUser.RawData[j.Name] = j.Values[0]

	}

	if !emailFound {
		ADLogger.Warning("User email not found, generating from username")
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
func (s *ADProvider) Handle(w http.ResponseWriter, r *http.Request,pathParams map[string]string) {
	s.connect()

	username := r.FormValue("username")
	password := r.FormValue("password")

	if s.config.GetAuthFromBAHeader {
		username, password = ExtractBAUsernameAndPasswordFromRequest(r)
	}

	if username == "" || password == "" {
		ADLogger.Error("Login attempt with empty username or password")
		s.provideErrorRedirect(w, r)
		return
	}

	ADLogger.Debug("DN: ", s.prepDN(username))

	var bindErr error

	if s.config.LDAPAdminUser != "" {
		bindErr = s.connection.Bind(s.config.LDAPAdminUser, s.config.LDAPAdminPassword)
	} else {
		bindErr = s.connection.Bind(s.prepDN(username), password)
	}

	if bindErr != nil {
		if s.config.LDAPAdminUser != "" {
			ADLogger.WithFields(logrus.Fields{
				"username": s.config.LDAPAdminUser,
				"error":    bindErr,
			}).Error("Bind failed for user")
		} else {
			ADLogger.WithFields(logrus.Fields{
				"username": username,
				"error":    bindErr,
			}).Error("Bind failed for user")
		}
		s.provideErrorRedirect(w, r)
		return
	}

	if s.config.LDAPAdminUser != "" {
		ADLogger.WithField("username", username).Info("User bind successful")
	}

	user, uErr := s.getUserData(username, password)
	if uErr != nil {
		ADLogger.WithFields(logrus.Fields{
			"username": username,
			"error":    uErr,
		}).Error("Lookup failed for user")
		s.provideErrorRedirect(w, r)
		return
	}

	constraintErr := s.checkConstraints(user)
	if constraintErr != nil {
		ADLogger.Error("Constraint failed: ", constraintErr)
		s.provideErrorRedirect(w, r)
		return
	}

	s.handler.CompleteIdentityAction(w, r, user, s.profile)

	ADLogger.Debug("Closing connection")
	s.connection.Close()
}

func (s *ADProvider) checkConstraints(user interface{}) error {
	ADLogger.Debug("Constraints for AD must be set in DN")
	return nil
}

// HandleCallback is not used
func (s *ADProvider) HandleCallback(w http.ResponseWriter, r *http.Request, onError func(tag string, errorMsg string, rawErr error, code int, w http.ResponseWriter, r *http.Request)) {

	ADLogger.Warning("Callback not implemented for provider")

}
