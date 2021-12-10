package providers

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"encoding/xml"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"sync"

	"github.com/TykTechnologies/tyk/certs"

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
var CertManager = certs.NewCertificateManager(FileLoader{}, "", nil, false)

type SAMLProvider struct {
	handler tap.IdentityHandler
	config  SAMLConfig
	profile tap.Profile
	m       *samlsp.Middleware
}

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
	EntityId            string
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
	SAMLLogger.Debugf("Initializing SAML profile with config: %+v", s.config)
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

	if s.m == nil {
		SAMLLogger.Debug("Initialising middleware SAML")
		//needs to match the signing cert if IDP
		certs := CertManager.List([]string{s.config.CertLocation}, certs.CertificateAny)

		if len(certs) == 0 {
			SAMLLogger.Error("certificate was not loaded")
		} else {
			SAMLLogger.Debug("Certificate Loaded")
		}

		keyPair := certs[0]
		log.Debugf("KeyPair: %+v", keyPair)
		idpMetadataURL, err := url.Parse(s.config.IDPMetadataURL)
		if err != nil {
			SAMLLogger.Errorf("Error parsing IDP metadata URL: %v", err)
		} else {
			SAMLLogger.Debug("Parsing IDP metadata URL successful")
		}

		SAMLLogger.Debugf("IDPmetadataURL is: %v", idpMetadataURL.String())
		rootURL, err := url.Parse(s.config.SAMLBaseURL)
		if err != nil {
			SAMLLogger.Errorf("Error parsing SAMLBaseURL: %v", err)
		} else {
			SAMLLogger.Debugf("Parsing SAML Base URL successful, Root URL: %+v", rootURL)
		}

		httpClient := http.DefaultClient

		metadata, err := samlsp.FetchMetadata(context.TODO(), httpClient, *idpMetadataURL)
		if err != nil {
			SAMLLogger.Errorf("Error retrieving IDP Metadata: %v", err)
		} else {
			SAMLLogger.Debug("IDP Metadata retrieved successfully")
			SAMLLogger.Debugf("IDP Metadata: %+v", metadata)
		}

		SAMLLogger.Debugf("Root URL: %v", rootURL.String())
		if keyPair == nil {
			SAMLLogger.Error("profile certificate was not loaded")
			return
		} else {
			SAMLLogger.Debug("Profile Certificate was loaded")
		}
		var key *rsa.PrivateKey
		if keyPair.PrivateKey == nil {
			SAMLLogger.Error("Private Key is nil not rsa.PrivateKey")
		} else {
			key = keyPair.PrivateKey.(*rsa.PrivateKey)
			SAMLLogger.Debugf("Private Key is present in the certificate and was loaded")
		}

		opts := samlsp.Options{
			URL:               *rootURL,
			Key:               key,
			AllowIDPInitiated: true,
		}

		metadataURL := rootURL.ResolveReference(&url.URL{Path: "auth/" + s.profile.ID + "/saml/metadata"})
		acsURL := rootURL.ResolveReference(&url.URL{Path: "auth/" + s.profile.ID + "/saml/callback"})
		sloURL := rootURL.ResolveReference(&url.URL{Path: "auth/" + s.profile.ID + "/saml/slo"})

		SAMLLogger.Debugf("SP metadata URL: %v", metadataURL.String())
		SAMLLogger.Debugf("SP acs URL: %v", acsURL.String())

		var forceAuthn = s.config.ForceAuthentication
		sp := saml.ServiceProvider{
			// if s.config.EntityId is empty, it will default to metadataUrl
			EntityID:          s.config.EntityId,
			Key:               key,
			Certificate:       keyPair.Leaf,
			MetadataURL:       *metadataURL,
			AcsURL:            *acsURL,
			SloURL:            *sloURL,
			IDPMetadata:       metadata,
			ForceAuthn:        &forceAuthn,
			AllowIDPInitiated: true,
		}

		s.m = &samlsp.Middleware{
			ServiceProvider: sp,
			Binding:         s.config.SAMLBinding,
			OnError:         samlsp.DefaultOnError,
			Session:         samlsp.DefaultSessionProvider(opts),
		}
		s.m.RequestTracker = samlsp.DefaultRequestTracker(opts, &s.m.ServiceProvider)
	}

}

func (s *SAMLProvider) Handle(w http.ResponseWriter, r *http.Request, pathParams map[string]string, profile tap.Profile) {
	SAMLLogger.Debugf("Handling SAML request: %+v", r)
	if s.m == nil {
		SAMLLogger.Error("cannot process request, middleware not loaded")
		return
	} else {
		SAMLLogger.Debug("Using saml middleware already initialized")
	}

	// If we try to redirect when the original request is the ACS URL we'll
	// end up in a loop so just fail and error instead
	if r.URL.Path == s.m.ServiceProvider.AcsURL.Path {
		SAMLLogger.Debugf("request path is the same as SP ACSUrl, then redirecting to failing state. Url: %v", r.URL.Path)
		s.provideErrorRedirect(w, r)
		return
	}

	var binding, bindingLocation string
	if s.m.Binding != "" {
		SAMLLogger.Debugf("Middleware binding is not empty: %v ", s.m.Binding)
		binding = s.m.Binding
		bindingLocation = s.m.ServiceProvider.GetSSOBindingLocation(binding)
	} else {
		SAMLLogger.Debug("Middleware binding is empty, then initializing")
		binding = saml.HTTPRedirectBinding
		bindingLocation = s.m.ServiceProvider.GetSSOBindingLocation(binding)
		if bindingLocation == "" {
			binding = saml.HTTPPostBinding
			bindingLocation = s.m.ServiceProvider.GetSSOBindingLocation(binding)
		}
	}
	SAMLLogger.Debugf("Binding: %v", binding)
	SAMLLogger.Debugf("BindingLocation: %v", bindingLocation)

	SAMLLogger.Debug("Performing Authentication request to: %v", binding)
	SAMLLogger.Debugf("Service Provider details: %+v", s.m.ServiceProvider)

	authReq, err := s.m.ServiceProvider.MakeAuthenticationRequest(bindingLocation)
	if err != nil {
		SAMLLogger.Error("Making authentication request: %+v", err.Error())
		s.provideErrorRedirect(w, r)
		return
	}

	SAMLLogger.Debugf("Auth Request: %+v", authReq)
	// relayState is limited to 80 bytes but also must be integrity protected.
	// this means that we cannot use a JWT because it is way to long. Instead
	// we set a signed cookie that encodes the original URL which we'll check
	// against the SAML response when we get it.
	relayState, err := s.m.RequestTracker.TrackRequest(w, r, authReq.ID)
	if err != nil {
		SAMLLogger.Error("Tracking request: %+v", err.Error())
		s.provideErrorRedirect(w, r)
		return
	}
	SAMLLogger.Debugf("Relay State: %+v", relayState)

	if binding == saml.HTTPRedirectBinding {
		redirectURL := authReq.Redirect(relayState)
		SAMLLogger.Debugf("Binding is redirect, then redirecting to %v", redirectURL.String())
		w.Header().Add("Location", redirectURL.String())
		w.WriteHeader(http.StatusFound)
		return
	}
	if binding == saml.HTTPPostBinding {
		SAMLLogger.Debug("binding is a POST binding.")
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
	err := r.ParseForm()
	if err != nil {
		SAMLLogger.Errorf("Error parsing form: %v", err)
	}
	SAMLLogger.Debugf("HandleCallback called, req details: %+v", r)

	var possibleRequestIDs = make([]string, 0)

	if s.m.ServiceProvider.AllowIDPInitiated {
		SAMLLogger.Debug("allowing IDP initiated ID")
		possibleRequestIDs = append(possibleRequestIDs, "")
		SAMLLogger.Debugf("Possible Requests Ids: %+v", possibleRequestIDs)
	} else {
		SAMLLogger.Debug("IDP Initiated flow not allowed")
	}

	trackedRequests := s.m.RequestTracker.GetTrackedRequests(r)
	SAMLLogger.Debugf("Tracked Requests: %+v", trackedRequests)
	for _, tr := range trackedRequests {
		possibleRequestIDs = append(possibleRequestIDs, tr.SAMLRequestID)
	}
	SAMLLogger.Debugf("Possible requests IDs: %+v", possibleRequestIDs)

	assertion, err := s.m.ServiceProvider.ParseResponse(r, possibleRequestIDs)
	if err != nil {
		PrintErrorStruct(err)
		SAMLLogger.Error(err)
		s.provideErrorRedirect(w, r)
		return
	} else {
		SAMLLogger.Debugf("Assertion: %+v", assertion)
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
	SAMLLogger.Debugf("User: %+v", thisUser)
	s.handler.CompleteIdentityAction(w, r, thisUser, s.profile)
}

func (s *SAMLProvider) HandleMetadata(w http.ResponseWriter, r *http.Request) {
	SAMLLogger.Debug("HandleMetadata Called...")
	buf, _ := xml.MarshalIndent(s.m.ServiceProvider.Metadata(), "", "  ")
	w.Header().Set("Content-Type", "application/samlmetadata+xml")
	w.Write(buf)
	return
}

func (s *SAMLProvider) provideErrorRedirect(w http.ResponseWriter, r *http.Request) {
	SAMLLogger.Debugf("provideErrorRedirect called... req details:\n%+v\n", r)
	http.Redirect(w, r, s.config.FailureRedirect, 301)
	return
}

func PrintErrorStruct(err error) {
	if err == nil {
		return
	}
	e := reflect.ValueOf(err).Elem()
	typeOfT := e.Type()

	for i := 0; i < e.NumField(); i++ {
		f := e.Field(i)
		SAMLLogger.Debugf("%d: %s %s = %v\n", i, typeOfT.Field(i).Name, f.Type(), f.Interface())
	}
}
