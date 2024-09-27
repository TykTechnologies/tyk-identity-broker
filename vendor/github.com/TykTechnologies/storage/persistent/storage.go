package persistent

import (
	"errors"

	"github.com/TykTechnologies/storage/persistent/internal/driver/mongo"

	"github.com/TykTechnologies/storage/persistent/internal/driver/mgo"

	"github.com/TykTechnologies/storage/persistent/internal/types"
)

const (
	OfficialMongo string = "mongo-go"
	Mgo           string = "mgo"
)

type (
	ClientOpts        types.ClientOpts
	PersistentStorage types.PersistentStorage
)

// NewPersistentStorage returns a persistent storage object that uses the given driver
func NewPersistentStorage(opts *ClientOpts) (types.PersistentStorage, error) {
	clientOpts := types.ClientOpts(*opts)
	switch opts.Type {
	case OfficialMongo:
		return mongo.NewMongoDriver(&clientOpts)
	case Mgo:
		return mgo.NewMgoDriver(&clientOpts)
	default:
		return nil, errors.New("invalid driver")
	}
}
