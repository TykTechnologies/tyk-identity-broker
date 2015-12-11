package tap

import (
	"net/http"
)

type TAProvider interface {
	Init(IdentityHandler, string) error
	Name() string
	ProviderType() ProviderType
	UseCallback() bool
	Handle(http.ResponseWriter, *http.Request)
	HandleCallback(http.ResponseWriter, *http.Request)
}
