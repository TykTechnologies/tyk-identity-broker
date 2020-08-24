package backends

import (
	"github.com/TykTechnologies/tyk-identity-broker/log"
	"github.com/TykTechnologies/tyk-identity-broker/tap"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
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

	profile := value.(*tap.Profile)

	// delete if exist, where matches the profile ID and org
	err := profilesCollection.Remove(bson.M{"ID":key,"OrgID":profile.OrgID})
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
	err := profilesCollection.Find(bson.M{"ID":key}).One(val)
	if err != nil {
		mongoLogger.Error("error reading profiles from mongo: " + err.Error())
	}
	return err
}

func (m MongoBackend) GetAll() []interface{} {
	var profiles []tap.Profile
	err := m.Db.C(m.Collection).Find(nil).All(&profiles)
	if err != nil {
		mongoLogger.Error("error reading profiles from mongo: " + err.Error())
	}

	result := make([]interface{}, len(profiles))
	for i, profile := range profiles {
		result[i] = profile
	}

	return result
}

func (m MongoBackend) DeleteKey(key string) error {
	profilesCollection := m.Db.C(m.Collection)

	err := profilesCollection.Remove(bson.M{"ID": key})
	if err != nil {
		mongoLogger.WithError(err).Error("removing profile")
	}

	return err
}
