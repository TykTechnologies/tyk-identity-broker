package providers

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/xml"
	"net/http"
	"net/url"
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

type SAMLProvider struct {
	handler tap.IdentityHandler
	config  SAMLConfig
	profile tap.Profile
	m       *samlsp.Middleware
}

var middleware *samlsp.Middleware

type SAMLConfig struct {
	IDPMetadataURL      string
	CertFile            string
	KeyFile             string
	SAMLBaseURL         string
	ForceAuthentication bool
	SAMLBinding         string
}

func (s *SAMLProvider) Init(handler tap.IdentityHandler, profile tap.Profile, config []byte) error {
	//if an external logger was set, then lets reload it to inherit those configs
	onceReloadADLogger.Do(func() {
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
		keyPair, err := tls.LoadX509KeyPair(s.config.CertFile, s.config.KeyFile)
		if err != nil {
			log.Errorf("Error loading keypair: %v", err)
		}

		keyPair.Leaf, err = x509.ParseCertificate(keyPair.Certificate[0])
		if err != nil {
			SAMLLogger.Errorf("Error parsing certificate: %v", err)
		}

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

		opts := samlsp.Options{
			URL: *rootURL,
			Key: keyPair.PrivateKey.(*rsa.PrivateKey),
		}

		metadataURL := rootURL.ResolveReference(&url.URL{Path: "auth/" + s.profile.ID + "/saml/metadata"})
		acsURL := rootURL.ResolveReference(&url.URL{Path: "auth/" + s.profile.ID + "/saml/callback"})
		sloURL := rootURL.ResolveReference(&url.URL{Path: "auth/" + s.profile.ID + "/saml/slo"})

		SAMLLogger.Debugf("SP metadata URL: %v", metadataURL.String())
		SAMLLogger.Debugf("SP acs URL: %v", acsURL.String())

		var forceAuthn = s.config.ForceAuthentication

		sp := saml.ServiceProvider{
			EntityID:          metadataURL.String(),
			Key:               keyPair.PrivateKey.(*rsa.PrivateKey),
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

func (s *SAMLProvider) Handle(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
	s.m = middleware
	// If we try to redirect when the original request is the ACS URL we'll
	// end up in a loop. This is a programming error, so we panic here. In
	// general this means a 500 to the user, which is preferable to a
	// redirect loop.
	//log.Debug(s.m)
	if r.URL.Path == s.m.ServiceProvider.AcsURL.Path {
		http.Error(w, "Server error", http.StatusInternalServerError)
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// relayState is limited to 80 bytes but also must be integrity protected.
	// this means that we cannot use a JWT because it is way to long. Instead
	// we set a signed cookie that encodes the original URL which we'll check
	// against the SAML response when we get it.
	relayState, err := s.m.RequestTracker.TrackRequest(w, r, authReq.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

func (s *SAMLProvider) HandleCallback(w http.ResponseWriter, r *http.Request, onError func(tag string, errorMsg string, rawErr error, code int, w http.ResponseWriter, r *http.Request)) {
	s.m = middleware

	err := r.ParseForm()
	if err != nil {
		SAMLLogger.Error(err)
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
		s.m.OnError(w, r, err)
		return
	}
	rawData := make(map[string]interface{}, 0)
	for _, v := range assertion.AttributeStatements {
		for _, att := range v.Attributes {
			SAMLLogger.Debugf("attribute name: %v\n", att.Name)
			rawData[att.Name] = ""
			for _, vals := range att.Values {
				rawData[att.Name] = vals.Value
				SAMLLogger.Debugf("vals.value: %v\n ", vals.Value)
			}

		}
	}

	//this is going to be a nightmare of slight differences between IDPs
	var email string
	name := rawData["http://schemas.xmlsoap.org/ws/2005/05/identity/claims/givenname"].(string) + " " +
		rawData["http://schemas.xmlsoap.org/ws/2005/05/identity/claims/surname"].(string)

	if _, ok := rawData["http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress"]; ok {
		email = rawData["http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress"].(string)
	} else if _, ok := rawData["http://schemas.xmlsoap.org/ws/2005/05/identity/claims/"]; ok {
		email = rawData["http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name"].(string)
	}

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
