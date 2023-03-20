package backends

import (
	"context"
	"encoding/json"
	"github.com/TykTechnologies/storage/persistent"
	"github.com/TykTechnologies/storage/persistent/dbm"

	"github.com/TykTechnologies/tyk-identity-broker/log"
	"github.com/TykTechnologies/tyk-identity-broker/tap"
)

var mongoPrefix = "mongo-backend"
var mongoLogger = log.Get().WithField("prefix", mongoPrefix).Logger

type MongoBackend struct {
	Store      persistent.PersistentStorage
	Collection string
}

func (m MongoBackend) Init(interface{}) {

}

func (m MongoBackend) SetKey(key string, orgId string, value interface{}) error {

	profile := value.(tap.Profile)
	filter := dbm.DBM{}
	filter["ID"] = key
	if orgId != "" {
		filter["OrgID"] = orgId
	}

	// delete if exists, where matches the profile ID and org
	err := m.Store.Delete(context.Background(), profile, filter)
	if err != nil {
		if err.Error() != "not found" {
			mongoLogger.WithError(err).Error("setting profile in mongo")
		}
	}

	err = m.Store.Insert(context.Background(), profile)
	if err != nil {
		mongoLogger.WithError(err).Error("inserting profile in mongo")
	}

	return err
}

func (m MongoBackend) GetKey(key string, orgId string, val interface{}) error {

	filter := dbm.DBM{}
	filter["ID"] = key
	if orgId != "" {
		filter["OrgID"] = orgId
	}

	p := tap.Profile{}
	err := m.Store.Query(context.Background(), p, &p, filter)
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

	filter := dbm.DBM{}
	if orgId != "" {
		filter["OrgID"] = orgId
	}

	err := m.Store.Query(context.Background(), tap.Profile{}, &profiles, filter)
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
	filter := dbm.DBM{}
	filter["ID"] = key
	if orgId != "" {
		filter["OrgID"] = orgId
	}

	err := m.Store.Delete(context.Background(), tap.Profile{}, filter)
	if err != nil {
		mongoLogger.WithError(err).Error("removing profile")
	}

	return err
}
