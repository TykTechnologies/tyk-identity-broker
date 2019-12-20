package configuration

import (
	"encoding/json"
	logger "github.com/TykTechnologies/tyk-identity-broker/log"
	"github.com/TykTechnologies/tyk-identity-broker/tothic"
	"io/ioutil"

	"github.com/TykTechnologies/tyk-identity-broker/tyk-api"
	"github.com/kelseyhightower/envconfig"
)

var failCount int
var log = logger.Get()
var mainLogger = log.WithField("prefix", "CONFIG")

const (
	MONGO = "mongo"
	FILE  = "file"
)

type IdentityBackendSettings struct {
	MaxIdle       int
	MaxActive     int
	Database      int
	Password      string
	EnableCluster bool
	Hosts         map[string]string
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

type Storage struct {
	StorageType string     `json:"storage_type" mapstructure:"storage_type"`
	MongoConf   *MongoConf `json:"mongo" mapstructure:"mongo"`
}

type Backend struct {
	ProfileBackendSettings  interface{}
	IdentityBackendSettings IdentityBackendSettings
}

// Configuration holds all configuration settings for TAP
type Configuration struct {
	Secret     string
	Port       int
	ProfileDir string
	BackEnd    Backend
	TykAPISettings    tyk.TykAPI
	HttpServerOptions struct {
		UseSSL   bool
		CertFile string
		KeyFile  string
	}
	SSLInsecureSkipVerify bool
	Storage               *Storage
}

//LoadConfig will load the config from a file
func LoadConfig(filePath string, conf *Configuration) {
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
	mainLogger.Debug("Settings Struct: ", conf.TykAPISettings)
}
