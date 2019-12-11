package data_loader

import (
	"github.com/Sirupsen/logrus"
	"github.com/TykTechnologies/tyk-identity-broker/tap"
	mgo "gopkg.in/mgo.v2"
	"strconv"
	"time"
)

var mongoPrefix = "mongo"
var profilesCollectionName = "profilesCollection"
var tibDbName = "db"

// MongoLoaderConf is the configuration struct for a MongoLoader
type MongoLoaderConf struct {
	dialInfo *mgo.DialInfo
}

// FileLoader implements DataLoader and will load TAP Profiles from a file
type MongoLoader struct {
	config MongoLoaderConf
	session *mgo.Session
}

// Init initialises the mongo loader
func (m *MongoLoader) Init(conf interface{}) error {

	var err error
	m.session, err = mgo.DialWithInfo(m.config.dialInfo)
	if err != nil {
		dataLogger.WithFields(logrus.Fields{
			"prefix": mongoPrefix,
			"error":    "Mongo connection failed:",
		}).Error("load failed!")

		time.Sleep(5 * time.Second)
		m.Init(conf)
	}
	return err
}

// LoadIntoStore will load, unmarshal and copy profiles into a an AuthRegisterBackend
func (m *MongoLoader) LoadIntoStore(store tap.AuthRegisterBackend) error {
	profiles := []tap.Profile{}
	err := m.session.DB(tibDbName).C(profilesCollectionName).Find(nil).All(&profiles)
	if err != nil {
		return err
	}

	var loaded int
	for _, profile := range profiles {
		inputErr := store.SetKey(profile.ID, profile)
		if inputErr != nil {
			dataLogger.WithField("error", inputErr).Error("Couldn't encode configuration")
		} else {
			loaded += 1
		}
	}
	return nil
}

//Flush creates a backup of the current loaded config
func (m *MongoLoader) Flush(store tap.AuthRegisterBackend) error{
	//read all
	oldSet := []tap.Profile{}
	err := m.session.DB(tibDbName).C(profilesCollectionName).Find(nil).All(&oldSet)
	if err != nil {
		return err
	}

	//create collection with the next name:
	ts := strconv.Itoa(int(time.Now().Unix()))
	bkCollectionName := "profiles_backup_" + ts + ".json"
	err = m.session.DB(tibDbName).C(bkCollectionName).Create(nil)
	if err != nil {
		return err
	}

	//put all the data there
	c := m.session.DB(tibDbName).C(bkCollectionName)
	err = c.Insert(&oldSet)
	if err != nil {
		return err
	}

	//save this in the current collection, so empty and store
	newSet := store.GetAll()
	profilesCollection := m.session.DB(tibDbName).C(profilesCollectionName)

	_, err = profilesCollection.RemoveAll(nil)
	err = profilesCollection.Insert(&newSet)
	if err != nil {
		return err
	}

	err = profilesCollection.Insert(&newSet)
	return err
}