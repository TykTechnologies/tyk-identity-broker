/* package tap wraps a set of interfaces and object to provide a generic interface to a delegated authentication
proxy */
package providers

import (
	"net/http"
	"tyk-identity-broker/tap"
	"github.com/markbates/goth"
)

// TAProvider is an interface that defines an actual handler for a specific authentication provider. It can wrap
// largert libraries (such as Goth for social), or individual pass-throughs such as LDAP.
type TAProvider interface {
	Init(tap.IdentityHandler, tap.Profile, []byte) error
	Name() string
	ProviderType() tap.ProviderType
	UseCallback() bool
	Handle(http.ResponseWriter, *http.Request) (goth.User, error)
	HandleCallback(http.ResponseWriter, *http.Request, func(tag string, errorMsg string, rawErr error, code int, w http.ResponseWriter, r *http.Request))
	GetProfile() tap.Profile
	GetHandler() tap.IdentityHandler
	GetCORS() bool
	GetCORSOrigin() string
	GetCORSHeaders() string
	GetCORSMaxAge() string
	GetLogTag() string
}
