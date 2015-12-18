package tap

import (
	"net/http"
)

type IdentityHandler interface {
	Init(interface{}) error
	CreateIdentity(interface{}) (string, error)
	LoginIdentity(string, string) (string, error)
	CompleteIdentityAction(http.ResponseWriter, *http.Request, interface{}, Profile)
}
