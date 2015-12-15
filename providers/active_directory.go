package providers

import (
	"encoding/json"
	"fmt"
	"github.com/lonelycode/go-ldap"
	"github.com/lonelycode/tyk-auth-proxy/tap"
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
	LDAPServer      string
	LDAPPort        string
	LDAPUserDN      string
	LDAPBaseDN      string
	LDAPFilter      string
	LDAPAttributes  []string
	FailureRedirect string
	SuccessRedirect string
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
	log.Warning(ADProviderLogTag + " Connect: starting...")
	var err error
	sName := fmt.Sprintf("%s:%s", s.config.LDAPServer, s.config.LDAPPort)
	log.Warning(ADProviderLogTag+" --> To: ", sName)
	s.connection, err = ldap.Dial("tcp", sName)
	if err != nil {
		log.Error(ADProviderLogTag+" Failed to dial: ", err)
		return
	}
	log.Warning(ADProviderLogTag + " Connect: finished...")
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

func (s *ADProvider) getUserData(username string) (interface{}, error) {
	log.Warning(ADProviderLogTag + " Search: starting...")

	search_request := ldap.NewSearchRequest(
		s.config.LDAPBaseDN,
		ldap.ScopeSingleLevel,
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
		return nil, err
	}

	for _, i := range sr.Entries {
		log.Info(i)
	}
	log.Warning(ADProviderLogTag+" User Data:", sr)
	log.Warning(ADProviderLogTag+" Search:", search_request.Filter, "-> num of entries = ", len(sr.Entries))
	return sr, nil
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

	log.Warning(ADProviderLogTag + " Closing connection")
	closeFail := s.connection.Close()
	if closeFail != nil {
		log.Error(ADProviderLogTag+" Closing failed! ", closeFail)
	}
}

func (s *ADProvider) checkConstraints(user interface{}) error {
	log.Warning(ADProviderLogTag + " Constraints for AD must be set in DN")
	return nil
}

func (s *ADProvider) HandleCallback(w http.ResponseWriter, r *http.Request, onError func(tag string, errorMsg string, rawErr error, code int, w http.ResponseWriter, r *http.Request)) {

	log.Warning(ADProviderLogTag + " Callback not implemented for provider")

}
