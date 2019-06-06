/* package identityHandlers provides a collection of handlers for target systems,
these handlers create accounts and sso tokens */
package identityHandlers

import (
	"fmt"
	logger "github.com/TykTechnologies/tyk-identity-broker/log"
	"github.com/TykTechnologies/tyk-identity-broker/tap"
	"net/http"
)

var log = logger.Get()
var DummyLogTag string = "[DUMMY ID HANDLER]"
var dummyLogger = log.WithField("prefix", "DUMMY ID HANDLER")

// DummyIdentityHandler is a dummy hndler, use for testing
type DummyIdentityHandler struct{}

// Init will set up the configuration of the handler
func (d DummyIdentityHandler) Init(conf interface{}) error {
	return nil
}

// Dummy method
func (d DummyIdentityHandler) CreateIdentity(i interface{}) (string, error) {
	dummyLogger.Info("Creating identity for: ", i)
	return "", nil
}

// Dummy method
func (d DummyIdentityHandler) LoginIdentity(user string, pass string) (string, error) {
	dummyLogger.Info("Logging in identity: ", user)
	return "12345", nil
}

// CompleteIdentityAction is called when an authenticated callback event is triggered, it should speak to
// the target system and generate / login the user. In this case it redirects the user to the ReturnURL.
func (d DummyIdentityHandler) CompleteIdentityAction(w http.ResponseWriter, r *http.Request, i interface{}, profile tap.Profile) {
	d.CreateIdentity(i)
	nonce, _ := d.LoginIdentity("DUMMY", "DUMMY")

	// After login, we need to redirect this user
	dummyLogger.Debug("--> Running redirect...")
	if profile.ReturnURL != "" {
		newURL := profile.ReturnURL + "?nonce=" + nonce
		http.Redirect(w, r, newURL, 301)
		return
	}

	dummyLogger.Warning("No return URL found, redirect failed.")
	fmt.Fprintf(w, "Success! (Have you set a return URL?)")
}
