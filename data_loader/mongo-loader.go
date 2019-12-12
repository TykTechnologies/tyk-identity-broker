package data_loader

import (
	"crypto/tls"
	"github.com/Sirupsen/logrus"
	"github.com/TykTechnologies/tyk-identity-broker/tap"
	"gopkg.in/mgo.v2"
	"net"
	"strconv"
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
	session *mgo.Session
}

// Init initialises the mongo loader
func (m *MongoLoader) Init(conf interface{}) error {
	m.config = conf.(MongoLoaderConf)

	var err error
	m.session, err = mgo.DialWithInfo(m.config.DialInfo)
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
	var profiles []tap.Profile
	database := m.config.DialInfo.Database
	err := m.session.DB(database).C(profilesCollectionName).Find(nil).All(&profiles)
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

	oldSet := []tap.Profile{}
	database := m.config.DialInfo.Database

	err := m.session.DB(database).C(profilesCollectionName).Find(nil).All(&oldSet)
	if err != nil {
		return err
	}

	//create collection with the next name:
	ts := strconv.Itoa(int(time.Now().Unix()))
	bkCollectionName := "profiles_backup_" + ts

	err = m.session.DB(database).C(bkCollectionName).Create(&mgo.CollectionInfo{})
	if err != nil {
		return err
	}

	//put all the data there
	collection := m.session.DB(database).C(bkCollectionName)
	for _, profile := range oldSet{
		err = collection.Insert(profile)
		if err != nil {
			return err
		}
	}

	//save this in the current collection, so empty and store
	newSet := store.GetAll()
	profilesCollection := m.session.DB(database).C(profilesCollectionName)

	_, err = profilesCollection.RemoveAll(nil)
	for _, profile := range newSet {
		err = profilesCollection.Insert(&profile)
		if err != nil {
			return err
		}
	}

	return nil
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