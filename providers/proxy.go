package providers

import (
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/lonelycode/tyk-auth-proxy/tap"
	"github.com/markbates/goth"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"regexp"
)

type ProxyHandlerConfig struct {
	TargetHost string
	OKCode     int
	OKResponse string
	OKRegex    string
}

type ProxyProvider struct {
	handler tap.IdentityHandler
	config  ProxyHandlerConfig
	profile tap.Profile
}

func (p *ProxyProvider) Init(handler tap.IdentityHandler, profile tap.Profile, config []byte) error {
	p.handler = handler
	p.profile = profile

	unmarshallErr := json.Unmarshal(config, &p.config)
	if unmarshallErr != nil {
		return unmarshallErr
	}

	return nil
}

func (p *ProxyProvider) Name() string {
	return "ProxyProvider"
}

func (p *ProxyProvider) ProviderType() tap.ProviderType {
	return tap.PASSTHROUGH_PROVIDER
}

func (p *ProxyProvider) UseCallback() bool {
	return false
}

func (p *ProxyProvider) respondFailure(rw http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(rw, "Authentication Failed")
}

func (p *ProxyProvider) Handle(rw http.ResponseWriter, r *http.Request) {
	// copy the request to a target
	target, tErr := url.Parse(p.config.TargetHost)
	if tErr != nil {
		log.Error("Failed to parse target URL: ", tErr)
		p.respondFailure(rw, r)
		return
	}
	thisProxy := httputil.NewSingleHostReverseProxy(target)

	// intercept the response
	recorder := httptest.NewRecorder()
	r.URL.Path = ""
	thisProxy.ServeHTTP(recorder, r)

	if recorder.Code >= 400 {
		log.Error("Code was: ", recorder.Code)
		p.respondFailure(rw, r)
		return
	}
	// check against passing signal
	if p.config.OKCode != 0 {
		if recorder.Code != p.config.OKCode {
			log.Error("Code was: ", recorder.Code, " expected: ", p.config.OKCode)
			p.respondFailure(rw, r)
			return
		}
	}

	thisBody, err := ioutil.ReadAll(recorder.Body)
	if p.config.OKResponse != "" {
		sEnc := b64.StdEncoding.EncodeToString(thisBody)
		if err != nil {
			log.Error("Could not read body.")
			p.respondFailure(rw, r)
			return
		}

		if sEnc != p.config.OKResponse {
			log.Error("Response was: ", sEnc, " expected: ", p.config.OKResponse)
			p.respondFailure(rw, r)
			return
		}
	}

	if p.config.OKRegex != "" {
		thisRegex, rErr := regexp.Compile(p.config.OKRegex)
		if rErr != nil {
			log.Error("Regex failure: ", rErr)
			p.respondFailure(rw, r)
			return
		}

		found := thisRegex.MatchString(string(thisBody))

		if !found {
			log.Error("Regex not found")
			p.respondFailure(rw, r)
			return
		}
	}

	thisUser := goth.User{
		UserID:   RandStringRunes(12),
		Provider: p.Name(),
	}

	// Complete the identity action
	p.handler.CompleteIdentityAction(rw, r, thisUser, p.profile)
}

func (p *ProxyProvider) HandleCallback(http.ResponseWriter, *http.Request, func(tag string, errorMsg string, rawErr error, code int, w http.ResponseWriter, r *http.Request)) {
	return
}
