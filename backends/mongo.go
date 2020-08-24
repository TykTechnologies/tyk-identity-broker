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

func (m MongoBackend) SetKey(key string,orgId string,value interface{}) error {
	profilesCollection := m.Db.C(m.Collection)

	filter := bson.M{"ID":key}
	if orgId != "" {
		filter["OrgID"] = orgId
	}
	// delete if exist, where matches the profile ID and org
	err := profilesCollection.Remove(filter)
	if err != nil {
		mongoLogger.WithError(err).Error("error setting profile in mongo: ")
	}

	err = profilesCollection.Insert(value)
	if err != nil {
		mongoLogger.WithError(err).Error("error setting profile in mongo: ")
	}

	return err
}

func (m MongoBackend) GetKey(key string,orgId string, val interface{}) error {
	profilesCollection := m.Db.C(m.Collection)

	filter := bson.M{"ID":key}
	if orgId != "" {
		filter["OrgID"] = orgId
	}

	err := profilesCollection.Find(filter).One(val)
	if err != nil {
		mongoLogger.Error("error reading profiles from mongo: " + err.Error())
	}
	return err
}

func (m MongoBackend) GetAll(orgId string) []interface{} {
	var profiles []tap.Profile

	filter := bson.M{}
	if orgId != "" {
		filter["OrgID"] = orgId
	}

	err := m.Db.C(m.Collection).Find(filter).All(&profiles)
	if err != nil {
		mongoLogger.Error("error reading profiles from mongo: " + err.Error())
	}

	result := make([]interface{}, len(profiles))
	for i, profile := range profiles {
		result[i] = profile
	}

	return result
}

func (m MongoBackend) DeleteKey(key string, orgId string) error {
	profilesCollection := m.Db.C(m.Collection)

	filter := bson.M{"ID":key}
	if orgId != "" {
		filter["OrgID"] = orgId
	}

	err := profilesCollection.Remove(filter)
	if err != nil {
		mongoLogger.WithError(err).Error("removing profile")
	}

	return err
}
