package backends

import (
	"encoding/json"
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

func (m *MongoBackend) getCollection() *mgo.Collection {
	session := m.Db.Session.Copy()
	return session.DB("").C(m.Collection)
}

func (m MongoBackend) SetKey(key string,orgId string,value interface{}) error {
	profilesCollection := m.getCollection()
	defer profilesCollection.Database.Session.Close()

	filter := bson.M{"ID":key}
	if orgId != "" {
		filter["OrgID"] = orgId
	}
	// delete if exist, where matches the profile ID and org
	err := profilesCollection.Remove(filter)
	if err != nil {
		if err.Error() != "not found" {
			mongoLogger.WithError(err).Error("setting profile in mongo")
		}
	}

	err = profilesCollection.Insert(value)
	if err != nil {
		mongoLogger.WithError(err).Error("setting inserting in mongo: ")
	}

	return err
}

func (m MongoBackend) GetKey(key string,orgId string, val interface{}) error {
	profilesCollection := m.getCollection()
	defer profilesCollection.Database.Session.Close()

	filter := bson.M{"ID":key}
	if orgId != "" {
		filter["OrgID"] = orgId
	}

	p := tap.Profile{}
	err := profilesCollection.Find(filter).One(&p)
	if err != nil {
		if err.Error() != "not found" {
			mongoLogger.WithError(err).Error("error reading profile from mongo, key:", key)
		}
	}

	// Mongo doesn't parse well the nested map[string]interface{} so, we need to use json marshal/unmarshal
	// Mongo let those maps as bson.M
	data, err := json.Marshal(p)
	if err != nil {
		mongoLogger.WithError(err).Error("error marshaling profile")
		return err
	}

	if err := json.Unmarshal(data, &val); err != nil {
		mongoLogger.WithError(err).Error("error un-marshaling profile ")
		return err
	}

	return err
}

func (m MongoBackend) GetAll(orgId string) []interface{} {
	var profiles []tap.Profile

	filter := bson.M{}
	if orgId != "" {
		filter["OrgID"] = orgId
	}

	profilesCollection := m.getCollection()
	defer profilesCollection.Database.Session.Close()
	err := profilesCollection.Find(filter).All(&profiles)
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
	profilesCollection := m.getCollection()
	defer profilesCollection.Database.Session.Close()

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
