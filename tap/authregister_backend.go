/*
	package tap wraps a set of interfaces and object to provide a generic interface to a delegated authentication

proxy
*/
package tap

import (
	"github.com/TykTechnologies/storage/persistent/model"
)

// ProfileStore is the minimum interface required to look up authentication
// profiles.  Implement this when your host application owns profile CRUD and
// you only need GetTapProfile to work (read path only).
type ProfileStore interface {
	GetKey(key string, orgId string, val interface{}) error
}

// KVStore is the minimum interface required for OAuth session state and nonce
// tokens.  tothic and TykIdentityHandler both write and delete ephemeral keys,
// so all three mutating methods are required.
type KVStore interface {
	GetKey(key string, orgId string, val interface{}) error
	SetKey(key string, orgId string, val interface{}) error
	DeleteKey(key string, orgId string) error
}

// AuthRegisterBackend is the full storage interface used throughout TIB.
// Any value implementing AuthRegisterBackend automatically satisfies both
// ProfileStore and KVStore.
type AuthRegisterBackend interface {
	Init(interface{}) error
	SetKey(key string, orgId string, val interface{}) error
	GetKey(key string, orgId string, val interface{}) error
	GetAll(orgId string) []interface{}
	DeleteKey(key string, orgId string) error
}

type DBObject interface {
	SetDBID(id model.ObjectID)
}
