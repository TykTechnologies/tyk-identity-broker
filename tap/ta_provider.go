/*
	package tap wraps a set of interfaces and object to provide a generic interface to a delegated authentication

proxy
*/
package tap

import (
	"net/http"
)

// TAProvider is an interface that defines an actual handler for a specific authentication provider. It can wrap
// largert libraries (such as Goth for social), or individual pass-throughs such as LDAP.
type TAProvider interface {
	Init(IdentityHandler, Profile, []byte) error
	Name() string
	ProviderType() ProviderType
	UseCallback() bool
	Handle(http.ResponseWriter, *http.Request, map[string]string, Profile)
	HandleCallback(http.ResponseWriter, *http.Request, func(tag string, errorMsg string, rawErr error, code int, w http.ResponseWriter, r *http.Request), Profile)
	HandleMetadata(http.ResponseWriter, *http.Request)
}
