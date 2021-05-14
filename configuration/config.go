package configuration

import (
	"encoding/json"
	"io/ioutil"

	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"

	logger "github.com/TykTechnologies/tyk-identity-broker/log"
	"github.com/TykTechnologies/tyk-identity-broker/tothic"
	tyk "github.com/TykTechnologies/tyk-identity-broker/tyk-api"
)

var failCount int
var log = logger.Get()
var mainLoggerTag = "CONFIG"
var mainLogger = log.WithField("prefix", mainLoggerTag)

const (
	MONGO = "mongo"
	FILE  = "file"
)

type IdentityBackendSettings struct {
	MaxIdle               int
	MaxActive             int
	Database              int
	Username              string
	Password              string
	EnableCluster         bool
	Hosts                 map[string]string // Deprecated: Use Addrs instead.
	Addrs                 []string
	UseSSL                bool
	SSLInsecureSkipVerify bool
	MasterName            string
	SentinelPassword      string
}

type MongoConf struct {
	DbName                     string `json:"db_name" mapstructure:"db_name"`
	MongoURL                   string `json:"mongo_url" mapstructure:"mongo_url"`
	MongoUseSSL                bool   `json:"mongo_use_ssl" mapstructure:"mongo_use_ssl"`
	MongoSSLInsecureSkipVerify bool   `json:"mongo_ssl_insecure_skip_verify" mapstructure:"mongo_ssl_insecure_skip_verify"`
	MaxInsertBatchSizeBytes    int    `json:"max_insert_batch_size_bytes" mapstructure:"max_insert_batch_size_bytes"`
	MaxDocumentSizeBytes       int    `json:"max_document_size_bytes" mapstructure:"max_document_size_bytes"`
	CollectionCapMaxSizeBytes  int    `json:"collection_cap_max_size_bytes" mapstructure:"collection_cap_max_size_bytes"`
	CollectionCapEnable        bool   `json:"collection_cap_enable" mapstructure:"collection_cap_enable"`
}

// Storage object to configure the storage where the profiles lives in
// it can be extended to work with other loaders. As file Loader is the default
// then we dont read the file path from here
type Storage struct {
	StorageType string     `json:"storage_type" mapstructure:"storage_type"`
	MongoConf   *MongoConf `json:"mongo" mapstructure:"mongo"`
}

// FileLoaderConf is the configuration struct for a FileLoader, takes a filename as main init
type FileLoaderConf struct {
	FileName   string
	ProfileDir string
}

type Backend struct {
	ProfileBackendSettings  interface{}
	IdentityBackendSettings IdentityBackendSettings
}

// Configuration holds all configuration settings for TAP
type Configuration struct {
	Secret string
	Port   int
	ProfileDir        string
	BackEnd           Backend
	TykAPISettings    tyk.TykAPI
	HttpServerOptions struct {
		UseSSL                bool
		CertFile              string
		KeyFile               string
		SSLInsecureSkipVerify bool
	}
	SSLInsecureSkipVerify bool
	Storage               *Storage
}

// LoadConfig will load the config from a file
func LoadConfig(filePath string, conf *Configuration) {
	log = logger.Get()
	mainLogger = &logrus.Entry{Logger: log}
	mainLogger = mainLogger.Logger.WithField("prefix", mainLoggerTag)

	configuration, err := ioutil.ReadFile(filePath)
	if err != nil {
		mainLogger.Error("Couldn't load configuration file: ", err)
		failCount += 1
		if failCount < 3 {
			LoadConfig(filePath, conf)
		} else {
			mainLogger.Fatal("Could not open configuration, giving up.")
		}
	} else {
		jsErr := json.Unmarshal(configuration, conf)
		if jsErr != nil {
			mainLogger.Error("Couldn't unmarshal configuration: ", jsErr)
		}
	}

	if err = envconfig.Process(tothic.EnvPrefix, conf); err != nil {
		mainLogger.Errorf("Failed to process config env vars: %v", err)
	}
	mainLogger.Debugf("\nConfig Loaded: %+v \n", conf)
	mainLogger.Debugf("\n Storage conf: %+v \n", conf.Storage)
	mainLogger.Debug("Settings Struct: ", conf.TykAPISettings)
}
