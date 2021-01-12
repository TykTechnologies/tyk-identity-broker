package providers

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"encoding/xml"
	"github.com/TykTechnologies/tyk/certs"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/markbates/goth"

	"github.com/crewjam/saml"

	"github.com/crewjam/saml/samlsp"

	logger "github.com/TykTechnologies/tyk-identity-broker/log"
	"github.com/sirupsen/logrus"

	"github.com/TykTechnologies/tyk-identity-broker/tap"
)

var onceReloadSAMLLogger sync.Once
var SAMLLogTag = "SAML AUTH"
var SAMLLogger = log.WithField("prefix", SAMLLogTag)

// certManager will fallback as files as default
var CertManager = certs.NewCertificateManager(nil, "", nil)

type SAMLProvider struct {
	handler tap.IdentityHandler
	config  SAMLConfig
	profile tap.Profile
	m       *samlsp.Middleware
}

var middleware *samlsp.Middleware

type SAMLConfig struct {
	IDPMetadataURL      string
	CertLocation        string
	SAMLBaseURL         string
	ForceAuthentication bool
	SAMLBinding         string
	SAMLEmailClaim      string
	SAMLForenameClaim   string
	SAMLSurnameClaim    string
	FailureRedirect     string
}

func (s *SAMLProvider) Init(handler tap.IdentityHandler, profile tap.Profile, config []byte) error {
	//if an external logger was set, then lets reload it to inherit those configs
	onceReloadSAMLLogger.Do(func() {
		log = logger.Get()
		SAMLLogger = &logrus.Entry{Logger: log}
		SAMLLogger = SAMLLogger.Logger.WithField("prefix", SAMLLogTag)
	})

	s.handler = handler
	s.profile = profile
	unmarshalErr := json.Unmarshal(config, &s.config)

	if unmarshalErr != nil {
		return unmarshalErr
	}
	s.initialiseSAMLMiddleware()

	return nil
}

func (s *SAMLProvider) Name() string {
	return "SAMLProvider"
}

func (s *SAMLProvider) ProviderType() tap.ProviderType {
	return tap.REDIRECT_PROVIDER
}

func (s *SAMLProvider) UseCallback() bool {
	return true
}

func (s *SAMLProvider) initialiseSAMLMiddleware() {
	if middleware == nil {

		SAMLLogger.Debug("Initialising middleware SAML")
		//needs to match the signing cert if IDP
		certs := CertManager.List([]string{s.config.CertLocation}, certs.CertificateAny)

		if len(certs) == 0 {
			SAMLLogger.Error("certificate was not loaded")
		}

		keyPair := certs[0]
		idpMetadataURL, err := url.Parse(s.config.IDPMetadataURL)
		if err != nil {
			SAMLLogger.Errorf("Error parsing IDP metadata URL: %v", err)
		}

		SAMLLogger.Debugf("IDPmetadataURL is: %v", idpMetadataURL.String())
		rootURL, err := url.Parse(s.config.SAMLBaseURL)
		if err != nil {
			SAMLLogger.Errorf("Error parsing SAMLBaseURL: %v", err)
		}

		httpClient := http.DefaultClient

		metadata, err := samlsp.FetchMetadata(context.TODO(), httpClient, *idpMetadataURL)
		if err != nil {
			SAMLLogger.Errorf("Error retrieving IDP Metadata: %v", err)
		}

		SAMLLogger.Debugf("Root URL: %v", rootURL.String())
		if keyPair == nil {
			SAMLLogger.Error("profile certificate was not loaded")
			return
		}
		var key *rsa.PrivateKey
		if keyPair.PrivateKey == nil {
			SAMLLogger.Error("Private Key is nil not rsa.PrivateKey")
		} else {
			key = keyPair.PrivateKey.(*rsa.PrivateKey)
		}

		opts := samlsp.Options{
			URL: *rootURL,
			Key: key,
		}

		metadataURL := rootURL.ResolveReference(&url.URL{Path: "auth/" + s.profile.ID + "/saml/metadata"})
		acsURL := rootURL.ResolveReference(&url.URL{Path: "auth/" + s.profile.ID + "/saml/callback"})
		sloURL := rootURL.ResolveReference(&url.URL{Path: "auth/" + s.profile.ID + "/saml/slo"})

		SAMLLogger.Debugf("SP metadata URL: %v", metadataURL.String())
		SAMLLogger.Debugf("SP acs URL: %v", acsURL.String())

		var forceAuthn = s.config.ForceAuthentication

		sp := saml.ServiceProvider{
			EntityID:          metadataURL.String(),
			Key:               key,
			Certificate:       keyPair.Leaf,
			MetadataURL:       *metadataURL,
			AcsURL:            *acsURL,
			SloURL:            *sloURL,
			IDPMetadata:       metadata,
			ForceAuthn:        &forceAuthn,
			AllowIDPInitiated: true,
		}

		middleware = &samlsp.Middleware{
			ServiceProvider: sp,
			Binding:         s.config.SAMLBinding,
			OnError:         samlsp.DefaultOnError,
			Session:         samlsp.DefaultSessionProvider(opts),
		}
		middleware.RequestTracker = samlsp.DefaultRequestTracker(opts, &middleware.ServiceProvider)
	}

}

func (s *SAMLProvider) Handle(w http.ResponseWriter, r *http.Request, pathParams map[string]string, profile tap.Profile) {
	if middleware == nil {
		SAMLLogger.Error("cannot process request, middleware not loaded")
		return
	}

	s.m = middleware
	// If we try to redirect when the original request is the ACS URL we'll
	// end up in a loop so just fail and error instead
	if r.URL.Path == s.m.ServiceProvider.AcsURL.Path {
		s.provideErrorRedirect(w, r)
		return
	}

	var binding, bindingLocation string
	if s.m.Binding != "" {
		binding = s.m.Binding
		bindingLocation = s.m.ServiceProvider.GetSSOBindingLocation(binding)
	} else {
		binding = saml.HTTPRedirectBinding
		bindingLocation = s.m.ServiceProvider.GetSSOBindingLocation(binding)
		if bindingLocation == "" {
			binding = saml.HTTPPostBinding
			bindingLocation = s.m.ServiceProvider.GetSSOBindingLocation(binding)
		}
	}
	SAMLLogger.Debugf("Binding: %v", binding)
	SAMLLogger.Debugf("BindingLocation: %v", bindingLocation)

	authReq, err := s.m.ServiceProvider.MakeAuthenticationRequest(bindingLocation)
	if err != nil {
		s.provideErrorRedirect(w, r)
		return
	}

	// relayState is limited to 80 bytes but also must be integrity protected.
	// this means that we cannot use a JWT because it is way to long. Instead
	// we set a signed cookie that encodes the original URL which we'll check
	// against the SAML response when we get it.
	relayState, err := s.m.RequestTracker.TrackRequest(w, r, authReq.ID)
	if err != nil {
		s.provideErrorRedirect(w, r)
		return
	}

	if binding == saml.HTTPRedirectBinding {
		redirectURL := authReq.Redirect(relayState)
		w.Header().Add("Location", redirectURL.String())
		w.WriteHeader(http.StatusFound)
		return
	}
	if binding == saml.HTTPPostBinding {
		w.Header().Add("Content-Security-Policy", ""+
			"default-src; "+
			"script-src 'sha256-AjPdJSbZmeWHnEc5ykvJFay8FTWeTeRbs9dutfZ0HqE='; "+
			"reflected-xss block; referrer no-referrer;")
		w.Header().Add("Content-type", "text/html")
		w.Write([]byte(`<!DOCTYPE html><html><body>`))
		w.Write(authReq.Post(relayState))
		w.Write([]byte(`</body></html>`))
		return
	}
}

func (s *SAMLProvider) HandleCallback(w http.ResponseWriter, r *http.Request, onError func(tag string, errorMsg string, rawErr error, code int, w http.ResponseWriter, r *http.Request), profile tap.Profile) {
	s.m = middleware

	err := r.ParseForm()
	if err != nil {
		SAMLLogger.Errorf("Error parsing form: %v", err)
	}

	var possibleRequestIDs = make([]string, 0)
	if s.m.ServiceProvider.AllowIDPInitiated {
		SAMLLogger.Debug("allowing IDP initiated ID")
		possibleRequestIDs = append(possibleRequestIDs, "")
	}

	trackedRequests := s.m.RequestTracker.GetTrackedRequests(r)
	for _, tr := range trackedRequests {
		possibleRequestIDs = append(possibleRequestIDs, tr.SAMLRequestID)
	}
	assertion, err := s.m.ServiceProvider.ParseResponse(r, possibleRequestIDs)
	if err != nil {
		s.provideErrorRedirect(w, r)
		return
	}
	rawData := make(map[string]interface{}, 0)
	var str strings.Builder
	for _, v := range assertion.AttributeStatements {
		for _, att := range v.Attributes {
			SAMLLogger.Debugf("attribute name: %v\n", att.Name)
			rawData[att.Name] = ""
			for _, vals := range att.Values {
				str.WriteString(vals.Value + " ")
				SAMLLogger.Debugf("vals.value: %v\n ", vals.Value)
			}
			rawData[att.Name] = strings.TrimSuffix(str.String(), " ")
			str.Reset()
		}
	}

	//this is going to be a nightmare of slight differences between IDPs
	// so lets make it configurable with a sensible backup
	var email string
	emailClaim := s.config.SAMLEmailClaim
	if emailClaim == "" {
		emailClaim = "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress"
	}

	if _, ok := rawData[emailClaim]; ok {
		email = rawData[emailClaim].(string)
	} else if _, ok := rawData["http://schemas.xmlsoap.org/ws/2005/05/identity/claims/"]; ok {
		email = rawData["http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name"].(string)
	}

	givenNameClaim := s.config.SAMLForenameClaim
	surnameClaim := s.config.SAMLSurnameClaim

	if givenNameClaim == "" {
		givenNameClaim = "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/givenname"
	}

	if surnameClaim == "" {
		surnameClaim = "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/surname"
	}
	name := rawData[givenNameClaim].(string) + " " +
		rawData[surnameClaim].(string)

	thisUser := goth.User{
		UserID:   name,
		Email:    email,
		Provider: "SAMLProvider",
		RawData:  rawData,
	}
	s.handler.CompleteIdentityAction(w, r, thisUser, s.profile)
}

func (s *SAMLProvider) HandleMetadata(w http.ResponseWriter, r *http.Request) {
	s.m = middleware

	buf, _ := xml.MarshalIndent(s.m.ServiceProvider.Metadata(), "", "  ")
	w.Header().Set("Content-Type", "application/samlmetadata+xml")
	w.Write(buf)
	return
}

func (s *SAMLProvider) provideErrorRedirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, s.config.FailureRedirect, 301)
	return
}
