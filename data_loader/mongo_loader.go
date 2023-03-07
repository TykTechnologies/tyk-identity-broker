package data_loader

import (
	"context"
	"encoding/json"
	"github.com/TykTechnologies/storage/persistent"
	"time"

	"github.com/TykTechnologies/tyk-identity-broker/tap"
)

var (
	mongoPrefix = "mongo"
)

// MongoLoaderConf is the configuration struct for a MongoLoader
type MongoLoaderConf struct {
	ClientOpts *persistent.ClientOpts
}

// MongoLoader implements DataLoader and will load TAP Profiles from a file
type MongoLoader struct {
	config    MongoLoaderConf
	store     persistent.PersistentStorage
	SkipFlush bool
}

type ProfilesBackup struct {
	Timestamp int           `bson:"timestamp" json:"timestamp"`
	Profiles  []tap.Profile `bson:"profiles" json:"profiles"`
}

// Init initialises the mongo loader
func (m *MongoLoader) Init(conf interface{}) error {
	mongoConfig := conf.(MongoLoaderConf)

	store, err := persistent.NewPersistentStorage(mongoConfig.ClientOpts)
	if err != nil {
		dataLogger.WithError(err).WithField("prefix", mongoPrefix).Error("failed to init MongoDB connection")
		time.Sleep(5 * time.Second)
		m.Init(conf)
	}

	m.store = store
	return err
}

// LoadIntoStore will load, unmarshal and copy profiles into a an AuthRegisterBackend
func (m *MongoLoader) LoadIntoStore(store tap.AuthRegisterBackend) error {
	var profiles []tap.Profile

	err := m.store.Query(context.Background(), tap.Profile{}, &profiles, nil)

	if err != nil {
		dataLogger.Error("error reading profiles from mongo: " + err.Error())
		return err
	}

	for _, profile := range profiles {
		inputErr := store.SetKey(profile.ID, profile.OrgID, profile)
		if inputErr != nil {
			dataLogger.WithField("error", inputErr).Error("Couldn't encode configuration")
		}
	}

	dataLogger.Info("Loaded profiles from Mongo")
	return nil
}

// Flush creates a backup of the current loaded config
func (m *MongoLoader) Flush(store tap.AuthRegisterBackend) error {
	//read all
	//save the changes in the main profile's collection, so empty and store as we don't know what was removed, updated or added
	updatedSet := store.GetAll("")

	//empty to store new changes
	err := m.store.DeleteWhere(context.Background(), tap.Profile{}, nil)
	if err != nil {
		dataLogger.WithError(err).Error("emptying profiles collection")
		return err
	}

	for _, p := range updatedSet {
		profile := tap.Profile{}
		switch p := p.(type) {
		case string:
			// we need to make this because redis return string instead objects
			if err := json.Unmarshal([]byte(p), &profile); err != nil {
				dataLogger.WithError(err).Error("un-marshaling interface for mongo flushing")
				return err
			}
		default:
			profile = p.(tap.Profile)
		}

		m.store.Insert(context.Background(), profile)
		if err != nil {
			dataLogger.WithError(err).Error("error refreshing profiles records in mongo")
			return err
		}
	}

	return nil
}
