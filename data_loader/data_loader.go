package data_loader

import (
	"github.com/sirupsen/logrus"
	"gopkg.in/mgo.v2"

	"github.com/TykTechnologies/tyk-identity-broker/configuration"
	logger "github.com/TykTechnologies/tyk-identity-broker/log"
	"github.com/TykTechnologies/tyk-identity-broker/tap"
)

var log = logger.Get()
var dataLoaderLoggerTag = "TIB DATA LOADER"
var dataLogger = log.WithField("prefix", dataLoaderLoggerTag)

// DataLoader is an interface that defines how data is loaded from a source into a AuthRegisterBackend interface store
type DataLoader interface {
	Init(conf interface{}) error
	LoadIntoStore(tap.AuthRegisterBackend) error
	Flush(tap.AuthRegisterBackend) error
}

func reloadDataLoaderLogger() {
	log = logger.Get()
	dataLogger = &logrus.Entry{Logger: log}
	dataLogger = dataLogger.Logger.WithField("prefix", dataLoaderLoggerTag)
}

func CreateMongoLoaderFromConnection(db *mgo.Database) DataLoader {
	var dataLoader DataLoader

	reloadDataLoaderLogger()
	dataLogger.Info("Set mongo loader for TIB")
	dataLoader = &MongoLoader{Db: db}

	return dataLoader
}

func CreateDataLoader(config configuration.Configuration, ProfileFilename string) (DataLoader, error) {
	var dataLoader DataLoader
	var loaderConf interface{}
	reloadDataLoaderLogger()

	//default storage
	storageType := configuration.FILE

	if config.Storage != nil {
		storageType = config.Storage.StorageType
	}

	switch storageType {
	case configuration.MONGO:
		dataLoader = &MongoLoader{}

		mongoConf := config.Storage.MongoConf
		dialInfo, err := MongoDialInfo(mongoConf.MongoURL, mongoConf.MongoUseSSL, mongoConf.MongoSSLInsecureSkipVerify)
		if err != nil {
			dataLogger.Error("Error getting mongo settings: " + err.Error())
			return nil, err
		}
		loaderConf = MongoLoaderConf{
			DialInfo: dialInfo,
		}
	default:
		//default: FILE
		dataLoader = &FileLoader{}
		//pDir := path.Join(config.ProfileDir, *ProfileFilename)
		loaderConf = configuration.FileLoaderConf{
			FileName:   ProfileFilename,
			ProfileDir: config.ProfileDir,
		}
	}

	err := dataLoader.Init(loaderConf)
	return dataLoader, err
}
