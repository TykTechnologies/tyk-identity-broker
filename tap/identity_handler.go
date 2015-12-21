/* package tap wraps a set of interfaces and object to provide a generic interface to a delegated authentication
proxy */
package tap

import (
	"github.com/markbates/goth"
	"net/http"
)

// IdentityHandler provides an interface that provides a generic way to handle the creation / login of an SSO
// session for a specific provider, it should generate users, tokens and SSO sesisons for whatever target system
// is being used off the back of a delegated authentication provider such as GPlus.
type IdentityHandler interface {
	Init(interface{}) error
	CompleteIdentityAction(http.ResponseWriter, *http.Request, interface{}, Profile)
}

func GenerateSSOKey(user goth.User) string {
	return user.UserID + "@" + user.Provider
}
