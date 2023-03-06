package data_loader

import (
	"github.com/TykTechnologies/storage/persistent"
	"github.com/TykTechnologies/tyk-identity-broker/configuration"
	logger "github.com/TykTechnologies/tyk-identity-broker/log"
	"github.com/TykTechnologies/tyk-identity-broker/tap"
	"github.com/sirupsen/logrus"
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
		// map from tib mongo conf structure to persistent.ClientOpts
		connectionConf := persistent.ClientOpts{
			ConnectionString:      mongoConf.MongoURL,
			UseSSL:                mongoConf.MongoUseSSL,
			SSLInsecureSkipVerify: mongoConf.MongoSSLInsecureSkipVerify,
			SessionConsistency:    mongoConf.SessionConsistency,
			Type:                  persistent.Mgo,
		}

		loaderConf = MongoLoaderConf{
			ClientOpts: &connectionConf,
		}
	default:
		//default: FILE
		dataLoader = &FileLoader{}
		loaderConf = configuration.FileLoaderConf{
			FileName:   ProfileFilename,
			ProfileDir: config.ProfileDir,
		}
	}

	err := dataLoader.Init(loaderConf)
	return dataLoader, err
}
