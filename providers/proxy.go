package providers

import (
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/jeffail/gabs"
	"tyk-identity-broker/tap"
	"github.com/markbates/goth"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"regexp"
)

var ProxyLogTag = "[PROXY PROVIDER] "

type ProxyHandlerConfig struct {
	TargetHost                         string
	OKCode                             int
	OKResponse                         string
	OKRegex                            string
	ResponseIsJson                     bool
	AccessTokenField                   string
	UsernameField                      string
	ExtractUserNameFromBasicAuthHeader bool
	CORS                               bool
	CORSOrigin                         string
	CORSHeaders                        string
	CORSMaxAge                         string
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
func (p *ProxyProvider) GetProfile() tap.Profile {
	return p.profile
}

func (p *ProxyProvider) GetHandler() tap.IdentityHandler {
	return p.handler
}

func (p *ProxyProvider) GetLogTag() string {
	return ProxyLogTag
}

func (p *ProxyProvider) GetCORS() bool {
	return p.config.CORS
}

func (p *ProxyProvider) GetCORSOrigin() string {
	return p.config.CORSOrigin
}

func (p *ProxyProvider) GetCORSHeaders() string {
	return p.config.CORSHeaders
}

func (p *ProxyProvider) GetCORSMaxAge() string {
	return p.config.CORSMaxAge
}


func (p *ProxyProvider) HandleCallback(http.ResponseWriter, *http.Request, func(tag string, errorMsg string, rawErr error, code int, w http.ResponseWriter, r *http.Request)) {
	return
}

func (p *ProxyProvider) Handle(rw http.ResponseWriter, r *http.Request) (goth.User, error){
	var user goth.User

	// copy the request to a target
	target, tErr := url.Parse(p.config.TargetHost)
	if tErr != nil {
		return user, fmt.Errorf(ProxyLogTag+"Failed to parse target URL: ", tErr)
	}
	thisProxy := httputil.NewSingleHostReverseProxy(target)

	// intercept the response
	recorder := httptest.NewRecorder()
	r.URL.Path = ""
	thisProxy.ServeHTTP(recorder, r)

	if recorder.Code >= 400 {
		return user, fmt.Errorf(ProxyLogTag+"Code was: ", recorder.Code)
	}
	// check against passing signal
	if p.config.OKCode != 0 {
		if recorder.Code != p.config.OKCode {
			return user, fmt.Errorf(ProxyLogTag+"Code was: ", recorder.Code, " expected: ", p.config.OKCode)
		}
	}

	thisBody, err := ioutil.ReadAll(recorder.Body)
	if p.config.OKResponse != "" {
		sEnc := b64.StdEncoding.EncodeToString(thisBody)
		if err != nil {
			return user, fmt.Errorf(ProxyLogTag + "Could not read body.")
		}

		if sEnc != p.config.OKResponse {
			shortStr := sEnc
			if len(sEnc) > 21 {
				shortStr = sEnc[:20] + "..."
			}
			return user, fmt.Errorf(ProxyLogTag+"Response was: '", shortStr, "' expected: '", p.config.OKResponse, "'")
		}
	}

	if p.config.OKRegex != "" {
		thisRegex, rErr := regexp.Compile(p.config.OKRegex)
		if rErr != nil {
			return user, fmt.Errorf(ProxyLogTag+"Regex failure: ", rErr)
		}

		found := thisRegex.MatchString(string(thisBody))

		if !found {
			return user, fmt.Errorf(ProxyLogTag + "Regex not found")
		}
	}

	uName := RandStringRunes(12)
	if p.config.ExtractUserNameFromBasicAuthHeader {
		thisU, _ := ExtractBAUsernameAndPasswordFromRequest(r)
		if thisU != "" {
			uName = thisU
		}
	}

	AccessToken := ""
	if p.config.ResponseIsJson {
		parsed, pErr := gabs.ParseJSON(thisBody)
		if pErr != nil {
			log.Warning(ProxyLogTag + "Parsing for access token field failed: ")
		} else {
			if p.config.AccessTokenField != "" {
				tok, fT := parsed.Path(p.config.AccessTokenField).Data().(string)
				if fT {
					AccessToken = tok
				}
			}
			if p.config.UsernameField != "" {
				thisU, fU := parsed.Path(p.config.UsernameField).Data().(string)
				if fU {
					uName = thisU
				}
			}
		}
	}

	user = goth.User{
		UserID:      uName,
		Provider:    p.Name(),
		AccessToken: AccessToken,
	}

	log.Debug("Username: ", user.UserID)
	log.Debug("Access token: ", user.AccessToken)

	return user, nil
}