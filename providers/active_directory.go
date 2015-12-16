package providers

import (
	"encoding/json"
	"fmt"
	"github.com/lonelycode/go-ldap"
	"github.com/lonelycode/tyk-auth-proxy/tap"
	"github.com/markbates/goth"
	"net/http"
	"strings"
)

var ADProviderLogTag = "[AD AUTH]"

type ADProvider struct {
	handler    tap.IdentityHandler
	config     ADConfig
	profile    tap.Profile
	connection *ldap.Conn
}

type ADConfig struct {
	LDAPServer         string
	LDAPPort           string
	LDAPUserDN         string
	LDAPBaseDN         string
	LDAPFilter         string
	LDAPEmailAttribute string
	LDAPAttributes     []string
	FailureRedirect    string
	SuccessRedirect    string
	DefaultDomain      string
}

func (s *ADProvider) Name() string {
	return "ADProvider"
}

func (s *ADProvider) ProviderType() tap.ProviderType {
	return tap.PASSTHROUGH_PROVIDER
}

func (s *ADProvider) UseCallback() bool {
	return false
}

func (s *ADProvider) connect() {
	log.Debug(ADProviderLogTag + " Connect: starting...")
	var err error
	sName := fmt.Sprintf("%s:%s", s.config.LDAPServer, s.config.LDAPPort)
	log.Debug(ADProviderLogTag+" --> To: ", sName)
	s.connection, err = ldap.Dial("tcp", sName)
	if err != nil {
		log.Error(ADProviderLogTag+" Failed to dial: ", err)
		return
	}
	log.Debug(ADProviderLogTag + " Connect: finished...")
}

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

func (s *ADProvider) getUserData(username string) (goth.User, error) {
	log.Debug(ADProviderLogTag + " Search: starting...")
	thisUser := goth.User{
		UserID:   username,
		Provider: "ADProvider",
	}
	var attrs []string
	attrs = s.config.LDAPAttributes
	attrs = append(attrs, s.config.LDAPEmailAttribute)

	// LDAP search is inconcistent, defaulting to using username, assuming username is an email,
	// otherwise we use an algo to create one

	// search_request := ldap.NewSearchRequest(
	// 	s.config.LDAPBaseDN,
	// 	ldap.ScopeSingleLevel,
	// 	ldap.DerefAlways,
	// 	0,
	// 	0,
	// 	false,
	// 	s.prepFilter(username),
	// 	s.config.LDAPAttributes,
	// 	nil)

	// sr, err := s.connection.Search(search_request)
	// if err != nil {
	// 	log.Error(ADProviderLogTag+" Failure in search: ", err)
	// 	return thisUser, err
	// }

	// emailFound := false
	// for _, entry := range sr.Entries {
	// 	for _, j := range entry.Attributes {
	// 		log.Debug("Checking ", j.Name, "with ", s.config.LDAPEmailAttribute)
	// 		if j.Name == s.config.LDAPEmailAttribute {
	// 			thisUser.Email = j.Values[0]
	// 			emailFound = true
	// 			break
	// 		}
	// 	}
	// 	if emailFound {
	// 		break
	// 	}
	// }

	// if !emailFound {
	// 	log.Warning("User email not found, generating from username")
	// 	if strings.Contains(username, "@") {
	// 		thisUser.Email = username
	// 	} else {
	// 		thisUser.Email = username + "@" + s.profile.OrgID + "-" + "ADProvider.com"
	// 	}
	// }

	thisUser.Email = s.generateUsername(username)

	log.Debug(ADProviderLogTag+" User Data:", thisUser)
	//log.Debug(ADProviderLogTag+" Search:", search_request.Filter, "-> num of entries = ", len(sr.Entries))
	return thisUser, nil
}

func (s *ADProvider) Handle(w http.ResponseWriter, r *http.Request) {
	s.connect()

	username := r.FormValue("username")
	password := r.FormValue("password")
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
	closeFail := s.connection.Close()
	if closeFail != nil {
		log.Error(ADProviderLogTag+" Closing failed! ", closeFail)
	}
}

func (s *ADProvider) checkConstraints(user interface{}) error {
	log.Debug(ADProviderLogTag + " Constraints for AD must be set in DN")
	return nil
}

func (s *ADProvider) HandleCallback(w http.ResponseWriter, r *http.Request, onError func(tag string, errorMsg string, rawErr error, code int, w http.ResponseWriter, r *http.Request)) {

	log.Warning(ADProviderLogTag + " Callback not implemented for provider")

}
