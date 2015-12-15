package identityHandlers

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/lonelycode/tyk-auth-proxy/tap"
	"net/http"
)

var log = logrus.New()
var DummyLogTag string = "[DUMMY ID HANDLER]"

type DummyIdentityHandler struct{}

func (d DummyIdentityHandler) CreateIdentity(i interface{}) (string, error) {
	log.Info("[DUMMY-ID-HANDLER]  Creating identity for: ", i)
	return "", nil
}

func (d DummyIdentityHandler) LoginIdentity(user string, pass string) (string, error) {
	log.Info("[DUMMY-ID-HANDLER]  Logging in identity: ", user)
	return "12345", nil
}

func (d DummyIdentityHandler) CompleteIdentityAction(w http.ResponseWriter, r *http.Request, i interface{}, profile tap.Profile) {
	d.CreateIdentity(i)
	nonce, _ := d.LoginIdentity("DUMMY", "DUMMY")

	// After login, we need to redirect this user
	log.Debug(DummyLogTag + " --> Running redirect...")
	if profile.ReturnURL != "" {
		newURL := profile.ReturnURL + "?nonce=" + nonce
		http.Redirect(w, r, newURL, 301)
		return
	}

	log.Warning(DummyLogTag + "No return URL found, redirect failed.")
	fmt.Fprintf(w, "Success! (Have you set a return URL?)")
}
