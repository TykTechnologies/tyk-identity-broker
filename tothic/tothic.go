/*
Package Tothic wraps Gothic behaviour for multi-tenant usage. Package gothic wraps common behaviour when using Goth. This makes it quick, and easy, to get up
and running with Goth. Of course, if you want complete control over how things flow, in regards
to the authentication process, feel free and use Goth directly.

See https://github.com/markbates/goth/examples/main.go to see this in action.
*/
package tothic

import (
	"errors"
	logger "github.com/TykTechnologies/tyk-identity-broker/log"
	"github.com/TykTechnologies/tyk-identity-broker/toth"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"net/http"
	"os"
)

// SessionName is the key used to access the session store.
const SessionName = "_gothic_session"

const EnvPrefix = "TYK_IB"

var log = logger.Get()

var TothErrorHandler func(string, string, error, int, http.ResponseWriter, *http.Request)

// Store can/should be set by applications using gothic. The default is a cookie store.
var Store sessions.Store

func init() {
	key := KeyFromEnv()
	Store = sessions.NewCookieStore([]byte(key))
}

func KeyFromEnv() (key string) {
	// To handle deprecation
	key = os.Getenv("SESSION_SECRET")
	temp := os.Getenv(EnvPrefix + "_SESSION_SECRET")
	if temp != "" {
		if key != "" {
			log.Warn("SESSION_SECRET is deprecated, TYK_IB_SESSION_SECRET overrides it when you set both.")
		}
		key = temp
	}

	if key == "" && temp == "" {
		log.Warn("toth/tothic: no TYK_IB_SESSION_SECRET environment variable is set. The default cookie store is not available and any calls will fail. Ignore this warning if you are using a different store.")
	}

	return
}

/*
BeginAuthHandler is a convienence handler for starting the authentication process.
It expects to be able to get the name of the provider from the query parameters
as either "provider" or ":provider".

BeginAuthHandler will redirect the user to the appropriate authentication end-point
for the requested provider.

See https://github.com/markbates/goth/examples/main.go to see this in action.
*/
func BeginAuthHandler(res http.ResponseWriter, req *http.Request, toth *toth.TothInstance) {
	url, err := GetAuthURL(res, req, toth)
	if err != nil {
		//res.WriteHeader(http.StatusBadRequest)
		//fmt.Fprintln(res, err)
		TothErrorHandler("[TOTHIC]", err.Error(), err, http.StatusBadRequest, res, req)
		return
	}

	http.Redirect(res, req, url, http.StatusTemporaryRedirect)
}

// GetState gets the state string associated with the given request
// This state is sent to the provider and can be retrieved during the
// callback.
var GetState = func(req *http.Request) string {
	return "state"
}

/*
GetAuthURL starts the authentication process with the requested provided.
It will return a URL that should be used to send users to.

It expects to be able to get the name of the provider from the query parameters
as either "provider" or ":provider".

I would recommend using the BeginAuthHandler instead of doing all of these steps
yourself, but that's entirely up to you.
*/
func GetAuthURL(res http.ResponseWriter, req *http.Request, toth *toth.TothInstance) (string, error) {

	providerName, err := GetProviderName(req)
	if err != nil {
		return "", err
	}

	provider, err := toth.GetProvider(providerName)
	if err != nil {
		return "", err
	}
	sess, err := provider.BeginAuth(GetState(req))
	if err != nil {
		return "", err
	}

	url, err := sess.GetAuthURL()
	if err != nil {
		return "", err
	}

	session, _ := Store.Get(req, SessionName)
	session.Values[SessionName] = sess.Marshal()
	err = session.Save(req, res)
	if err != nil {
		return "", err
	}

	return url, err
}

/*
CompleteUserAuth does what it says on the tin. It completes the authentication
process and fetches all of the basic information about the user from the provider.

It expects to be able to get the name of the provider from the query parameters
as either "provider" or ":provider".

See https://github.com/markbates/goth/examples/main.go to see this in action.
*/
var CompleteUserAuth = func(res http.ResponseWriter, req *http.Request, toth *toth.TothInstance) (goth.User, error) {

	providerName, err := GetProviderName(req)
	if err != nil {
		return goth.User{}, err
	}

	provider, err := toth.GetProvider(providerName)
	if err != nil {
		return goth.User{}, err
	}

	session, _ := Store.Get(req, SessionName)

	if session.Values[SessionName] == nil {
		return goth.User{}, errors.New("could not find a matching session for this request")
	}

	sess, err := provider.UnmarshalSession(session.Values[SessionName].(string))
	if err != nil {
		return goth.User{}, err
	}

	_, err = sess.Authorize(provider, req.URL.Query())

	if err != nil {
		return goth.User{}, err
	}

	return provider.FetchUser(sess)
}

// GetProviderName is a function used to get the name of a provider
// for a given request. By default, this provider is fetched from
// the URL query string. If you provide it in a different way,
// assign your own function to this variable that returns the provider
// name for your request.
var GetProviderName = getProviderName

func getProviderName(req *http.Request) (string, error) {
	provider := mux.Vars(req)["provider"]
	if provider == "" {
		provider = mux.Vars(req)[":provider"]
	}
	if provider == "" {
		return provider, errors.New("you must select a provider")
	}
	return provider, nil
}
