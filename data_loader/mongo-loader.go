package data_loader

import (
	"crypto/tls"
	"github.com/Sirupsen/logrus"
	"github.com/TykTechnologies/tyk-identity-broker/tap"
	"gopkg.in/mgo.v2"
	"net"
	"time"
)

var mongoPrefix = "mongo"
var profilesCollectionName = "profilesCollection"

// MongoLoaderConf is the configuration struct for a MongoLoader
type MongoLoaderConf struct {
	DialInfo *mgo.DialInfo
}

// MongoLoader implements DataLoader and will load TAP Profiles from a file
type MongoLoader struct {
	config MongoLoaderConf
	Db *mgo.Database
}

type ProfilesBackup struct {
	Timestamp int  `bson:"timestamp" json:"timestamp"`
	Profiles []tap.Profile `bson:"profiles" json:"profiles"`
}

// Init initialises the mongo loader
func (m *MongoLoader) Init(conf interface{}) error {
	m.config = conf.(MongoLoaderConf)

	var err error
	session, err := mgo.DialWithInfo(m.config.DialInfo)
	if err != nil {
		dataLogger.WithFields(logrus.Fields{
			"prefix": mongoPrefix,
			"error":    "Mongo connection failed:",
		}).Error("load failed!")

		time.Sleep(5 * time.Second)
		m.Init(conf)
	}
	m.Db = session.DB("")
	return err
}

// LoadIntoStore will load, unmarshal and copy profiles into a an AuthRegisterBackend
func (m *MongoLoader) LoadIntoStore(store tap.AuthRegisterBackend) error {
	var profiles []tap.Profile

	err := m.Db.C(profilesCollectionName).Find(nil).All(&profiles)
	if err != nil {
		dataLogger.Error("error reading profiles from mongo: "+err.Error())
		return err
	}

	for _, profile := range profiles {
		inputErr := store.SetKey(profile.ID, profile)
		if inputErr != nil {
			dataLogger.WithField("error", inputErr).Error("Couldn't encode configuration")
		}
	}

	dataLogger.Info("Loaded profiles from Mongo")
	return nil
}

//Flush creates a backup of the current loaded config
func (m *MongoLoader) Flush(store tap.AuthRegisterBackend) error{
	//read all

	//save the changes in the main profiles collection, so empty and store as we dont know what was removed, updated or added
	updatedSet := store.GetAll()
	profilesCollection := m.Db.C(profilesCollectionName)

	//empty to store new changes
	_, err := profilesCollection.RemoveAll(nil)
	if err != nil {
		return err
	}
	err = profilesCollection.Insert(updatedSet...)

	return err
}


func MongoDialInfo(mongoURL string, useSSL bool, SSLInsecureSkipVerify bool) (dialInfo *mgo.DialInfo, err error) {

	if dialInfo, err = mgo.ParseURL(mongoURL); err != nil {
		return dialInfo, err
	}

	if useSSL {
		dialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
			tlsConfig := &tls.Config{}
			if SSLInsecureSkipVerify {
				tlsConfig.InsecureSkipVerify = true
			}
			return tls.Dial("tcp", addr.String(), tlsConfig)
		}
	}

	return dialInfo, err
}