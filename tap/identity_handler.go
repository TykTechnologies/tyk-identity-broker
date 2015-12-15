package tap

import (
	"net/http"
)

type IdentityHandler interface {
	CreateIdentity(interface{}) (string, error)
	LoginIdentity(string, string) (string, error)
	CompleteIdentityAction(http.ResponseWriter, *http.Request, interface{}, Profile)
}
