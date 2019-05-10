package main

import (
	"encoding/json"
	"github.com/TykTechnologies/tyk-identity-broker/tothic"
	"io/ioutil"

	"github.com/TykTechnologies/tyk-identity-broker/tyk-api"
	"github.com/kelseyhightower/envconfig"
)

var failCount int

type IdentityBackendSettings struct {
	MaxIdle       int
	MaxActive     int
	Database      int
	Password      string
	EnableCluster bool
	Hosts         map[string]string
}

// Configuration holds all configuration settings for TAP
type Configuration struct {
	Secret     string
	Port       int
	ProfileDir string
	BackEnd    struct {
		ProfileBackendSettings  interface{}
		IdentityBackendSettings IdentityBackendSettings
	}
	TykAPISettings    tyk.TykAPI
	HttpServerOptions struct {
		UseSSL   bool
		CertFile string
		KeyFile  string
	}
	SSLInsecureSkipVerify bool
}

// loadConfig will load the config from a file
func loadConfig(filePath string, conf *Configuration) {
	configuration, err := ioutil.ReadFile(filePath)
	if err != nil {
		logger.Error("Couldn't load configuration file: ", err)
		failCount += 1
		if failCount < 3 {
			loadConfig(filePath, conf)
		} else {
			logger.Fatal("Could not open configuration, giving up.")
		}
	} else {
		jsErr := json.Unmarshal(configuration, conf)
		if jsErr != nil {
			logger.Error("Couldn't unmarshal configuration: ", jsErr)
		}
	}

	if err = envconfig.Process(tothic.EnvPrefix, conf); err != nil {
		logger.Errorf("Failed to process config env vars: %v", err)
	}

	logger.Debug("[MAIN] Settings Struct: ", conf.TykAPISettings)
}
