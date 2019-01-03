package main

import (
	"encoding/json"
	"io/ioutil"

	"github.com/TykTechnologies/tyk-identity-broker/tyk-api"
	"github.com/kelseyhightower/envconfig"
)

const envPrefix = "TYK_IB"

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
func loadConfig(filePath string, configStruct *Configuration) {
	configuration, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Error("Couldn't load configuration file: ", err)
		failCount += 1
		if failCount < 3 {
			loadConfig(filePath, configStruct)
		} else {
			log.Fatal("Could not open configuration, giving up.")
		}
	} else {
		jsErr := json.Unmarshal(configuration, &configStruct)
		if jsErr != nil {
			log.Error("Couldn't unmarshal configuration: ", jsErr)
		}
	}

	log.Debug("[MAIN] Settings Struct: ", configStruct.TykAPISettings)

	if err = envconfig.Process(envPrefix, configStruct); err != nil {
		log.Errorf("Failed to process config env vars: %v", err)
	}
}
