/*
	package tap wraps a set of interfaces and object to provide a generic interface to a delegated authentication

proxy
*/
package tap

import (
	"github.com/TykTechnologies/storage/persistent/model"
)

// AuthRegisterBackend is an interface to provide storage for profiles loaded into TAP
type AuthRegisterBackend interface {
	Init(interface{})
	SetKey(key string, orgId string, val interface{}) error
	GetKey(key string, orgId string, val interface{}) error
	GetAll(orgId string) []interface{}
	DeleteKey(key string, orgId string) error
}

type DBObject interface {
	SetDBID(id model.ObjectID)
}
