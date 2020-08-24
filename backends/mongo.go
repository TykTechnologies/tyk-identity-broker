package backends

import (
	"github.com/TykTechnologies/tyk-identity-broker/log"
	"github.com/TykTechnologies/tyk-identity-broker/tap"
	"gopkg.in/mgo.v2"
)

var mongoPrefix = "mongo-backend"
var mongoLogger = log.Get().WithField("prefix", mongoPrefix).Logger
type MongoBackend struct{
	Db *mgo.Database
	Collection string
}

func (m MongoBackend) Init(interface{}) {

}

func (m MongoBackend) SetKey(key string,value interface{}) error {
	profilesCollection := m.Db.C(m.Collection)

	profile := value.(tap.Profile)
	// delete if exist, where matches the profile ID and org
	err := profilesCollection.Remove(tap.Profile{ID:key,OrgID:profile.OrgID})
	if err != nil {
		mongoLogger.WithError(err).Error("error setting profile in mongo: ")
	}

	err = profilesCollection.Insert(value)
	if err != nil {
		mongoLogger.WithError(err).Error("error setting profile in mongo: ")
	}

	return err
}

func (m MongoBackend) GetKey(key string, val interface{}) error {
	profilesCollection := m.Db.C(m.Collection)
	profile := tap.Profile{ID:key}

	err := profilesCollection.Find(profile).One(&val)
	if err != nil {
		mongoLogger.Error("error reading profiles from mongo: " + err.Error())
	}
	return err
}

func (m MongoBackend) GetAll() []interface{} {
	var profiles []interface{}
	err := m.Db.C(m.Collection).Find(nil).All(&profiles)
	if err != nil {
		mongoLogger.Error("error reading profiles from mongo: " + err.Error())
	}
	return profiles
}

func (m MongoBackend) DeleteKey(key string) error {
	profilesCollection := m.Db.C(m.Collection)

	profile := tap.Profile{ID:key}
	err := profilesCollection.Remove(profile)

	if err != nil {
		mongoLogger.WithError(err).Error("removing profile")
	}

	return err
}
